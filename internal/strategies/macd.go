package strategies

import (
	"errors"
	"time"

	"github.com/markcheno/go-talib"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) processMACDStrategy(binSize string) error { // nolint:funlen,unused
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(binSize)
	count := cfg.MacdSlowCount + cfg.MacdSigCount

	fromTime, err := utils.FromTime(time.Now().UTC(), binSize, count)
	if err != nil {
		return err
	}

	candles, err := s.tradeAPI.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Symbol,
		BinSize:   binSize,
		Count:     int32(count),
		StartTime: fromTime.Format(bitmex.TradeTimeFormat),
	})
	if err != nil {
		return err
	}

	closes := s.fetchCloses(candles)
	if len(closes) == 0 {
		return errors.New("all candles invalid")
	}

	lastCandleTS, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		return err
	}

	timestamps, err := s.fetchTSFromCandles(candles)
	if err != nil {
		return err
	}

	macd := s.tradeCalc.CalcMACD(
		closes,
		s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdFastCount, talib.EMA,
		s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount, talib.EMA,
		s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSigCount, talib.WMA,
	)
	signal := models.Signal{
		MACDFast:           s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdFastCount,
		MACDSlow:           s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount,
		MACDSig:            s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSigCount,
		BinSize:            bin,
		Timestamp:          lastCandleTS,
		SignalType:         models.MACD,
		SignalValue:        macd.Sig,
		MACDValue:          macd.Value,
		MACDHistogramValue: macd.HistogramValue,
	}
	_, err = s.db.SaveSignal(signal)
	if err != nil {
		return err
	}

	macdDiverg, err := s.processMACDSignals(bin, timestamps, candles)
	if err != nil {
		return err
	}

	s.log.Infof("processMACDSignals defined -->: %#v", macdDiverg)
	if macdDiverg.bearDiverg {
		s.log.Infof("place sell order-------------->macd")
		//	err := s.placeBitmexOrder(types.SideSell)
		//	if err != nil {
		//		return err
		//	}
	}
	if macdDiverg.bullDiverg {
		s.log.Infof("place buy order-------------->macd")
		//	err := s.placeBitmexOrder(types.SideBuy)
		//	if err != nil {
		//		return err
		//	}
	}

	return nil
}

// TODO test this
func (s *Strategies) processMACDSignals( // nolint:unused
	binSize models.BinSize, timestamps []time.Time, candles []bitmex.TradeBuck,
) (result struct {
	bearDiverg bool
	bullDiverg bool
}, err error) {
	signals, err := s.db.GetSignalsByTS(models.MACD, binSize, timestamps)
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
		bearDivergIsDefined, err := s.defineBearDivergence(binSize, filteredCandles, signals)
		if err != nil {
			return result, err
		}
		result.bearDiverg = bearDivergIsDefined
	}
	if lastSignal.MACDHistogramValue < 0 {
		bullDivergenceIsDefined, err := s.defineBullDivergence(binSize, filteredCandles, signals)
		if err != nil {
			return result, err
		}
		result.bullDiverg = bullDivergenceIsDefined
	}

	return result, nil
}

func (s *Strategies) validateDefineMechanism( // nolint:unused
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

func (s *Strategies) defineBullDivergence( // nolint:unused
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
	if minHistVals.secondMin <= minHistVals.firstMin {
		s.log.Debugf("bullDivergence not confirmed by signals")
		return false, nil
	}
	if fCandles[minHistVals.secondMinIndx].Close > fCandles[minHistVals.firstMinIndx].Close {
		s.log.Debug("bull divergence not confirmed by candles")
		s.log.Debugf("first min histogram val:%v", minHistVals.firstMin)
		s.log.Debugf("first close val:%v", fCandles[minHistVals.firstMinIndx].Close)
		s.log.Debugf("second min histogram val:%v", minHistVals.secondMin)
		s.log.Debugf("second close val:%v", fCandles[minHistVals.secondMinIndx].Close)
		return false, nil
	}

	s.log.Debug("bull divergence confirmed")
	s.log.Debugf("first min histogram val:%v", minHistVals.firstMin)
	s.log.Debugf("first close val:%v", fCandles[minHistVals.firstMinIndx].Close)
	s.log.Debugf("second min histogram val:%v", minHistVals.secondMin)
	s.log.Debugf("second close val:%v", fCandles[minHistVals.secondMinIndx].Close)

	return true, nil
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

func (s *Strategies) defineBearDivergence( // nolint:unused
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
	// check histogram bearDiverg
	maxHistVals, err := s.findTwoMax(timeFrame, signals)
	if err != nil {
		return false, err
	}
	s.log.Debugf("findTwoMax max hist vals: %#v", maxHistVals)
	if maxHistVals.secondMax >= maxHistVals.firstMax {
		// exit - bearDiverg not confirmed
		s.log.Debugf("bearDivergence not confirmed by signals")
		return false, nil
	}
	// check price reversal
	if fCandles[maxHistVals.secondMaxIndx].Close < fCandles[maxHistVals.firstMaxIndx].Close {
		// exit - bearDiverg not confirmed
		s.log.Debugf("bearDivergence not confirmed by candles")
		s.log.Debugf("first max histogram val:%v", maxHistVals.firstMax)
		s.log.Debugf("first close val:%v", fCandles[maxHistVals.firstMaxIndx].Close)
		s.log.Debugf("second max histogram val:%v", maxHistVals.secondMax)
		s.log.Debugf("second close val:%v", fCandles[maxHistVals.secondMaxIndx].Close)
		return false, nil
	}

	s.log.Debugf("bearDivergence is confirmed")
	s.log.Debugf("first max histogram val:%v", maxHistVals.firstMax)
	s.log.Debugf("first close val:%v", fCandles[maxHistVals.firstMaxIndx].Close)
	s.log.Debugf("second max histogram val:%v", maxHistVals.secondMax)
	s.log.Debugf("second close val:%v", fCandles[maxHistVals.secondMaxIndx].Close)

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

func (s *Strategies) filterCandlesBySignals( // nolint:unused
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
				copyCandle := candle
				filteredCandles[i] = &copyCandle
				break
			}
		}
	}

	return filteredCandles, nil
}
