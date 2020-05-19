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

func (s *Strategies) SignalsInit() error {
	for _, binSize := range s.cfg.GlobStrategies.GetBinSizes() {
		err := s.binProcess(binSize)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Strategies) binProcess(binSize string) error {
	binType, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	count := s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount * 2

	startTime, err := utils.FromTime(time.Now().UTC(), binSize, count)
	if err != nil {
		return err
	}

	candles, err := s.tradeApi.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Symbol,
		BinSize:   binSize,
		Count:     int32(count),
		StartTime: startTime.Format(bitmex.TradeTimeFormat),
	})
	if err != nil {
		return err
	}
	err = s.checkCloses(candles)
	if err != nil {
		return err
	}

	cache := s.candlesCaches.GetCache(binType)
	if cache != nil {
		cache.StoreBatch(candles)
	}

	closes := s.fetchCloses(candles)
	if len(closes) < s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount {
		return errors.New("candles less than macd slow count")
	}

	var step int
	for i := len(candles) - s.cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSigCount; i < len(candles); i++ {
		// need for save signal into db
		candle := candles[i]
		timestamp, err := time.Parse(
			tradeapi.TradeBucketedTimestampLayout,
			candle.Timestamp,
		)
		if err != nil {
			return err
		}

		// MACD
		err = s.macdSave(timestamp, binType, closes[:i], i)
		if err != nil {
			return err
		}

		// RSI
		err = s.rsiSave(timestamp, binType, closes, i)
		if err != nil {
			return err
		}

		// Others
		err = s.otherSignals(timestamp, binType, closes, i)
		if err != nil {
			return err
		}

		step++
	}

	return nil
}

func (s *Strategies) macdSave(
	timestamp time.Time, size models.BinSize, closes []float64, step int) error {

	macd := s.tradeCalc.CalcMACD(
		closes,
		s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdFastCount,
		talib.EMA,
		s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSlowCount,
		talib.EMA,
		s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSigCount,
		talib.WMA,
	)
	_, err := s.db.SaveSignal(models.Signal{
		BinSize:            size,
		MACDFast:           s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdFastCount,
		MACDSlow:           s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSlowCount,
		MACDSig:            s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSigCount,
		Timestamp:          timestamp,
		SignalType:         models.MACD,
		SignalValue:        macd.Sig,
		MACDValue:          macd.Value,
		MACDHistogramValue: macd.HistogramValue,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Strategies) rsiSave(
	timestamp time.Time, size models.BinSize, closes []float64, step int) error {
	signals, err := s.db.GetSignalsByTs(models.RSI, size, []time.Time{timestamp})
	if err != nil {
		return err
	}
	if len(signals) != 0 {
		if signals[0].SignalValue >= float64(s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMaxBorder) {
			s.rsiPrev.maxBorderInProc = true
			s.rsiPrev.minBorderInProc = false
			s.log.Infof("max border is overcome up")
		} else if signals[0].SignalValue <= float64(s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMinBorder) {
			s.rsiPrev.minBorderInProc = true
			s.rsiPrev.maxBorderInProc = false
			s.log.Infof("min border is overcome - down")
		} else {
			s.rsiPrev.maxBorderInProc = false
			s.rsiPrev.minBorderInProc = false
		}
		s.log.Warnf("rsi signal by ts: %v, and binsize: %v already exist", timestamp, size)
		return nil
	}

	rsiValues := closes[:step]
	rsi := s.tradeCalc.CalcRSI(rsiValues, s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiCount-1)
	_, err = s.db.SaveSignal(models.Signal{
		N:           s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiCount,
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})

	if rsi.Value >= float64(s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMaxBorder) {
		s.rsiPrev.maxBorderInProc = true
		s.rsiPrev.minBorderInProc = false
		s.log.Infof("max border is overcome up")
	} else if rsi.Value <= float64(s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMinBorder) {
		s.rsiPrev.minBorderInProc = true
		s.rsiPrev.maxBorderInProc = false
		s.log.Infof("min border is overcome - down")
	} else {
		s.rsiPrev.maxBorderInProc = false
		s.rsiPrev.minBorderInProc = false
	}
	return err
}

func (s *Strategies) otherSignals(timestamp time.Time, size models.BinSize, closes []float64, step int) error {
	if step < s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).GetCandlesCount {
		return nil
	}
	sigs, err := s.db.GetSignalsByTs(models.BolingerBand, size, []time.Time{timestamp})
	if err != nil {
		return err
	}
	if len(sigs) != 0 {
		s.log.Warnf("other signals by ts: %v, and binsize: %v already exist", timestamp, size)
		return nil
	}

	maStartIndx, maStopIndx := step-s.cfg.GlobStrategies.GetCfgByBinSize(size.String()).GetCandlesCount, step
	values := closes[maStartIndx:maStopIndx]
	signals := s.tradeCalc.CalcSignals(values, talib.EMA)
	_, err = s.db.SaveSignal(models.Signal{
		N:           len(values),
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.SMA,
		SignalValue: signals.SMA,
	})
	if err != nil {
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           len(values),
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.EMA,
		SignalValue: signals.EMA,
	})
	if err != nil {
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           len(values),
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.WMA,
		SignalValue: signals.WMA,
	})
	if err != nil {
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:          len(values),
		BinSize:    size,
		Timestamp:  timestamp,
		SignalType: models.BolingerBand,
		BBTL:       signals.BB.TL,
		BBML:       signals.BB.ML,
		BBBL:       signals.BB.BL,
	})
	if err != nil {
		return err
	}

	return nil
}
