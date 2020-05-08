package strategies

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) processMACDStrategy(binSize string) error {
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	fromTime, err := utils.FromTime(time.Now().UTC(), binSize, s.cfg.Strategies.MacdSlowCount)
	if err != nil {
		return err
	}

	candles, err := s.tradeApi.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Currency,
		BinSize:   binSize,
		Count:     int32(s.cfg.Strategies.MacdSlowCount),
		StartTime: fromTime.Format(bitmex.TradeTimeFormat),
	})
	if err != nil {
		return err
	}

	closes := s.fetchCloses(candles)
	if len(closes) == 0 {
		return errors.New("all candles invalid")
	}

	lastCandleTs, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)

	timestamps, err := s.fetchTsFromCandles(candles)
	if err != nil {
		return err
	}

	macdLastTsBySigCount := timestamps[len(timestamps)-s.cfg.Strategies.MacdSigCount:]

	signals, err := s.db.GetSignalsByTs([]models.SignalType{models.MACD}, []models.BinSize{bin}, macdLastTsBySigCount)
	if err != nil {
		return err
	}
	fastValues := closes[len(closes)-s.cfg.Strategies.MacdFastCount:]

	macd := s.tradeCalc.CalculateMACD(fastValues, closes, s.fetchMacdHistVals(signals), trademath.EMAIndication)
	signal := models.Signal{
		MACDFast:           s.cfg.Strategies.MacdFastCount,
		MACDSlow:           s.cfg.Strategies.MacdSlowCount,
		MACDSig:            s.cfg.Strategies.MacdSigCount,
		BinSize:            bin,
		Timestamp:          lastCandleTs,
		SignalType:         models.MACD,
		SignalValue:        macd.Sig,
		MACDHistogramValue: macd.HistogramValue,
	}
	_, err = s.db.SaveSignal(signal)
	if err != nil {
		return err
	}

	signals = append(signals, &signal)

	macdDiverg, err := s.processMACDSignals(bin, timestamps, candles)
	if err != nil {
		return err
	}

	// todo send order
	s.log.Infof("processMACDSignals defined -->: %#v", macdDiverg)
	return nil
}

// TODO test this
func (s *Strategies) processMACDSignals(
	binSize models.BinSize, timestamps []time.Time, candles []bitmex.TradeBuck,
) (result struct {
	divergence  bool
	convergence bool
}, err error) {
	signals, err := s.db.GetSignalsByTs([]models.SignalType{models.MACD}, []models.BinSize{binSize}, timestamps)
	if err != nil {
		return result, err
	}

	filteredCandles, err := s.filterCandlesBySignals(candles, signals)
	if err != nil {
		return result, err
	}

	if len(filteredCandles) == 0 {
		s.log.Debugf("processMACDSignals exit, by signals - candles not exist")
		return result, nil
	}

	lastSignal := signals[len(signals)-1]
	if lastSignal.MACDHistogramValue > 0 {
		divergenceIsDefined, err := s.defineDivergence(binSize, filteredCandles, signals)
		if err != nil {
			return result, err
		}
		result.divergence = divergenceIsDefined
		s.log.Debug(divergenceIsDefined)
	}
	if lastSignal.MACDHistogramValue < 0 {
		convergenceIsDefined, err := s.defineConvergence(binSize, filteredCandles, signals)
		if err != nil {
			return result, err
		}
		result.convergence = convergenceIsDefined
		s.log.Debug(convergenceIsDefined)
	}

	return result, nil
}

func (s *Strategies) validateDefineMechanism(
	fCandles []*bitmex.TradeBuck, signals []*models.Signal,
) error {
	if len(signals) == 0 || len(fCandles) == 0 {
		return errors.New("empty signals or filtered candles")
	}
	if len(signals) != len(fCandles) {
		return errors.New("signals count not equal filtered candles count")
	}

	return nil
}

func (s *Strategies) defineConvergence(
	binSize models.BinSize, fCandles []*bitmex.TradeBuck, signals []*models.Signal,
) (bool, error) {
	err := s.validateDefineMechanism(fCandles, signals)
	if err != nil {
		return false, err
	}
	timeFrame := s.macdTimeFrameDefine(binSize)
	if timeFrame == 0 || len(signals) < timeFrame {
		return false, errors.New("time frame is 0 or count signals less than time frame")
	}
	minHistVals, err := s.findTwoMin(timeFrame, signals)
	if err != nil {
		return false, err
	}
	s.log.Debug(minHistVals)

	return false, nil
}

func (s *Strategies) findTwoMin(tframe int, signals []*models.Signal) (result struct {
	firstMinIndx  int
	firstMin      float64
	secondMinIndx int
	secondMin     float64
}, err error) {
	if tframe == 0 || len(signals) == 0 {
		return result, errors.New("empty time frame or signals")
	}
	if tframe > len(signals) {
		return result, errors.New("signals less than time frame")
	}

	var (
		lastNegativeCount int
		lastPositiveIndx  int
	)
	for i, signal := range signals {
		if signal.MACDHistogramValue > 0 {
			lastPositiveIndx = i
			signals[i] = nil
			lastNegativeCount = 0
			continue
		}
		lastNegativeCount++
	}

	if len(signals)-1 == lastPositiveIndx {
		return result, errors.New("last signal histogram value is positive")
	}

	if lastNegativeCount < tframe {
		return result, errors.New("negative signals count less than time frame")
	}

	for i := 0; i < len(signals); i++ {
		if i <= lastPositiveIndx {
			continue
		}
		signal := signals[i]
		if signal.MACDHistogramValue <= result.firstMin {
			result.secondMin = result.firstMin
			result.secondMinIndx = result.firstMinIndx
			result.firstMin = signal.MACDHistogramValue
			result.firstMinIndx = i
		}
		if signal.MACDHistogramValue < result.secondMin && signal.MACDHistogramValue > result.firstMin {
			result.secondMin = signal.MACDHistogramValue
			result.secondMinIndx = i
		}
	}

	return result, err
}

func (s *Strategies) defineDivergence(
	binSize models.BinSize, fCandles []*bitmex.TradeBuck, signals []*models.Signal,
) (bool, error) {
	err := s.validateDefineMechanism(fCandles, signals)
	if err != nil {
		return false, err
	}
	timeFrame := s.macdTimeFrameDefine(binSize)
	if timeFrame == 0 || len(signals) < timeFrame {
		return false, errors.New("time frame is 0 or count signals less than time frame")
	}
	// check histogram divergence
	maxHistVals, err := s.findTwoMax(timeFrame, signals)
	if err != nil {
		return false, err
	}
	s.log.Debugf("findTwoMax max hist vals: %#v", maxHistVals)
	if maxHistVals.secondMax >= maxHistVals.firstMax {
		// exit - divergence not confirmed
		s.log.Debugf("divergence not confirmed by signals")
		return false, nil
	}
	// check price reversal
	if fCandles[maxHistVals.secondMaxIndx].Close < fCandles[maxHistVals.firstMaxIndx].Close {
		// exit - divergence not confirmed
		s.log.Debugf("divergence not confirmed by candles")
		return false, nil
	}

	return true, nil
}

func (s *Strategies) findTwoMax(tframe int, signals []*models.Signal) (result struct {
	firstMaxIndx  int
	firstMax      float64
	secondMaxIndx int
	secondMax     float64
}, err error) {
	if tframe == 0 || len(signals) == 0 {
		return result, errors.New("empty time frame or signals")
	}
	if tframe > len(signals) {
		return result, errors.New("signals less than time frame")
	}

	var (
		lastPositiveCount int
		lastNegativeIndx  int
	)
	for i, signal := range signals {
		if signal.MACDHistogramValue < 0 {
			lastNegativeIndx = i
			signals[i] = nil
			lastPositiveCount = 0
			continue
		}
		lastPositiveCount++
	}

	if len(signals)-1 == lastNegativeIndx {
		return result, errors.New("last signal histogram value is negative")
	}

	if lastPositiveCount < tframe {
		return result, errors.New("positive signals count less than time frame")
	}

	for i := 0; i < len(signals); i++ {
		if i <= lastNegativeIndx {
			continue
		}
		signal := signals[i]
		if signal.MACDHistogramValue >= result.firstMax {
			result.secondMax = result.firstMax
			result.secondMaxIndx = result.firstMaxIndx
			result.firstMax = signal.MACDHistogramValue
			result.firstMaxIndx = i
		}
		if signal.MACDHistogramValue > result.secondMax && signal.MACDHistogramValue < result.firstMax {
			result.secondMax = signal.MACDHistogramValue
			result.secondMaxIndx = i
		}
	}

	return result, err
}

func (s *Strategies) filterCandlesBySignals(
	candles []bitmex.TradeBuck, signals []*models.Signal,
) ([]*bitmex.TradeBuck, error) {
	var filteredCandles = make([]*bitmex.TradeBuck, len(signals))
	for i, signal := range signals {
		for _, candle := range candles {
			timestamp, err := time.Parse(tradeapi.TradeBucketedTimestampLayout, candle.Timestamp)
			if err != nil {
				return nil, err
			}
			if timestamp.Equal(signal.Timestamp) {
				filteredCandles[i] = &candle
				break
			}
		}
	}

	return filteredCandles, nil
}
