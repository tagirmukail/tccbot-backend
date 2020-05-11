package strategies

import (
	"errors"
	"fmt"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/types"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) processBBStrategy(binSize string, count int32) error {
	var (
		err     error
		signals *trademath.Singals
		candles []bitmex.TradeBuck
	)
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	for i := 0; i < 5; i++ {
		candles, signals, err = s.processBBStrategyCandles(binSize, count)
		if err != nil {
			continue
		}
		break
	}

	err = s.processBBSignals(candles, signals, bin)
	if err != nil {
		return err
	}

	return nil
}

func (s *Strategies) processBBStrategyCandles(binSize string, count int32) ([]bitmex.TradeBuck, *trademath.Singals, error) {
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		return nil, nil, fmt.Errorf("processBBStrategyCandles models.ToBinSize error: %v", err)
	}
	from, err := utils.FromTime(time.Now().UTC(), binSize, int(count))
	if err != nil {
		return nil, nil, err
	}
	candles, err := s.tradeApi.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Symbol,
		BinSize:   binSize,
		Count:     count,
		StartTime: from.Format(bitmex.TradeTimeFormat),
	})
	if err != nil {
		s.log.Errorf("processBBStrategyCandles tradeApi.GetBitmex().GetTradeBucketed error: %v", err)
		return nil, nil, err
	}
	if len(candles) < s.cfg.Strategies.BBLastCandlesCount {
		s.log.Debug("processBBStrategyCandles there are fewer candles than necessary for the signal bolinger band")
		return nil, nil, errors.New("there are fewer candles than necessary for the signal bolinger band")
	}
	closes := s.fetchCloses(candles)
	if len(closes) == 0 {
		s.log.Debug("processBBStrategyCandles closes is empty")
		return nil, nil, errors.New("all candles invalid")
	}
	lastCandleTs, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		s.log.Debugf("processBBStrategyCandles last candle timestamp parse error: %v", err)
		return nil, nil, err
	}
	signals := s.tradeCalc.CalculateSignals(closes)
	err = s.saveSignals(lastCandleTs, bin, int(count), &signals)
	if err != nil {
		s.log.Debugf("processBBStrategyCandles db.SaveSignal error: %v", err)
		return nil, nil, err
	}

	s.log.Debugf("processed signals - close: %v for bin size:%s - signals: %#v", candles[len(candles)-1].Close, binSize, signals)
	return candles, &signals, err
}

func (s *Strategies) processBBSignals(
	candles []bitmex.TradeBuck, signals *trademath.Singals, binSize models.BinSize,
) error {
	s.log.Infof("start process bolinger band signals for bin_size:%v", binSize)
	err := s.retryProcess(candles, binSize, s.processBB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Strategies) saveSignals(timestamp time.Time, bin models.BinSize, n int, signals *trademath.Singals) error {
	_, err := s.db.SaveSignal(models.Signal{
		N:          n,
		BinSize:    bin,
		Timestamp:  timestamp,
		SignalType: models.BolingerBand,
		BBTL:       signals.BB.TL,
		BBML:       signals.BB.ML,
		BBBL:       signals.BB.BL,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal bolinger band error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.SMA,
		SignalValue: signals.SMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal sma error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.EMA,
		SignalValue: signals.EMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal ema error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.WMA,
		SignalValue: signals.WMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal wma error: %v", err)
		return err
	}
	return nil
}

func (s *Strategies) processBB(
	candles []bitmex.TradeBuck, binSize models.BinSize,
) error {
	lastCandles := s.fetchLastCandlesForBB(candles)
	if len(lastCandles) == 0 {
		err := errors.New("processBB last candles fo BB signal is empty")
		s.log.Debug(err)
		return err
	}

	lastCandlesTs, err := s.fetchTsFromCandles(lastCandles)
	if err != nil {
		return err
	}

	lastSignals, err := s.db.GetSignalsByTs([]models.SignalType{models.BolingerBand}, []models.BinSize{binSize}, lastCandlesTs)
	if err != nil {
		return err
	}

	s.processTrend(binSize, lastCandles, lastSignals)
	return nil
}

func (s *Strategies) processTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) {
	candlesCount, signalsCount := len(candles), len(signals)
	if candlesCount != signalsCount || candlesCount == 0 || signalsCount == 0 {
		s.log.Debugf("processTrend - count candles:%d, but signals:%d", candlesCount, signalsCount)
		return
	}

	s.log.Infof("start process bolinger band trend")
	firstCandle, firstSignal := candles[0], signals[0]
	if firstCandle.Close > firstSignal.BBTL {
		// uptrend detection
		s.upTrend(binSize, candles[1:], signals[1:])
	}
	if firstCandle.Close < firstSignal.BBBL {
		// downtrend detection
		s.downTrend(binSize, candles[1:], signals[1:])
	}
	s.log.Infof("finish process bolinger band trend")
	return
}

func (s *Strategies) upTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) {
	s.log.Infof("start process bolinger band up trend")

	if len(candles) == 0 || len(signals) == 0 {
		s.log.Debugf("upTrend empty candles or signals")
		return
	}
	for i, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		if candle.Close <= signals[i].BBTL {
			// uptrend broken
			s.log.Debugf("uptrend broken, #bin_size:%v, #close:%v, #bbtl:%v", binSize, candle.Close, signals[i].BBTL)
			return
		}
	}

	s.log.Infof("uptrend successfully completed for bin_size:%v and close:%v", binSize, candles[len(candles)-1].Close)

	for i := 0; i < s.cfg.Strategies.RetryProcessCount; i++ {
		err := s.placeBitmexOrder(types.SideSell, models.BolingerBand)
		if err != nil {
			s.log.Warnf("placeBitmexOrder sell failed: %v", err)
			continue
		}
		break
	}
}

func (s *Strategies) downTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) {
	s.log.Infof("start process bolinger band down trend")

	if len(candles) == 0 || len(signals) == 0 {
		s.log.Debugf("downTrend empty candles or signals")
		return
	}
	for i, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		if candle.Close >= signals[i].BBBL {
			// downtrend broken
			s.log.Debugf("downtrend broken, #bin_size:%v, #close:%v, #bbtl:%v", binSize, candle.Close, signals[i].BBTL)
			return
		}
	}
	s.log.Infof("downtrend successfully completed for bin_size:%v and close:%v", binSize, candles[len(candles)-1].Close)

	for i := 0; i < s.cfg.Strategies.RetryProcessCount; i++ {
		err := s.placeBitmexOrder(types.SideBuy, models.BolingerBand)
		if err != nil {
			s.log.Warnf("placeBitmexOrder buy failed: %v", err)
			continue
		}
		break
	}
}
