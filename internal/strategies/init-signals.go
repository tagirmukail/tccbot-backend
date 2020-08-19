package strategies

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/config"

	"github.com/markcheno/go-talib"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) SignalsInit() error {
	cfg, err := s.configurator.GetConfig()
	if err != nil {
		s.log.Fatal(err)
	}
	for _, binSize := range cfg.GlobStrategies.GetBinSizes() {
		err := s.binProcess(cfg, binSize)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Strategies) binProcess(cfg *config.GlobalConfig, binSize string) error {
	binType, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	count := cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount * 2

	startTime, err := utils.FromTime(time.Now().UTC(), binSize, count)
	if err != nil {
		return err
	}

	candles, err := s.tradeAPI.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    cfg.ExchangesSettings.Bitmex.Symbol,
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
	if len(closes) < cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSlowCount {
		return errors.New("candles less than macd slow count")
	}

	var step int
	for i := len(candles) - cfg.GlobStrategies.GetCfgByBinSize(binSize).MacdSigCount; i < len(candles); i++ {
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
		err = s.macdSave(cfg, timestamp, binType, closes[:i])
		if err != nil {
			return err
		}

		// RSI
		err = s.rsiSave(cfg, timestamp, binType, closes, i)
		if err != nil {
			return err
		}

		// Others
		err = s.otherSignals(cfg, timestamp, binType, closes, i)
		if err != nil {
			return err
		}

		step++
	}

	return nil
}

func (s *Strategies) macdSave(
	cfg *config.GlobalConfig, timestamp time.Time, size models.BinSize, closes []float64) error {

	macd := s.tradeCalc.CalcMACD(
		closes,
		cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdFastCount,
		talib.EMA,
		cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSlowCount,
		talib.EMA,
		cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSigCount,
		talib.WMA,
	)
	_, err := s.db.SaveSignal(models.Signal{
		BinSize:            size,
		MACDFast:           cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdFastCount,
		MACDSlow:           cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSlowCount,
		MACDSig:            cfg.GlobStrategies.GetCfgByBinSize(size.String()).MacdSigCount,
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
	cfg *config.GlobalConfig, timestamp time.Time, size models.BinSize, closes []float64, step int) error {
	signals, err := s.db.GetSignalsByTS(models.RSI, size, []time.Time{timestamp})
	if err != nil {
		return err
	}
	if len(signals) != 0 {
		switch {
		case signals[0].SignalValue >= float64(cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMaxBorder):
			s.rsiPrev.maxBorderInProc = true
			s.rsiPrev.minBorderInProc = false
			s.log.Infof("max border is overcome up")
		case signals[0].SignalValue <= float64(cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMinBorder):
			s.rsiPrev.minBorderInProc = true
			s.rsiPrev.maxBorderInProc = false
			s.log.Infof("min border is overcome - down")
		default:
			s.rsiPrev.maxBorderInProc = false
			s.rsiPrev.minBorderInProc = false
		}
		s.log.Warnf("rsi signal by ts: %v, and binsize: %v already exist", timestamp, size)
		return nil
	}

	rsiValues := closes[:step]
	rsi := s.tradeCalc.CalcRSI(rsiValues, cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiCount-1)
	_, err = s.db.SaveSignal(models.Signal{
		N:           cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiCount,
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})

	switch {
	case rsi.Value >= float64(cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMaxBorder):
		s.rsiPrev.maxBorderInProc = true
		s.rsiPrev.minBorderInProc = false
		s.log.Infof("max border is overcome up")
	case rsi.Value <= float64(cfg.GlobStrategies.GetCfgByBinSize(size.String()).RsiMinBorder):
		s.rsiPrev.minBorderInProc = true
		s.rsiPrev.maxBorderInProc = false
		s.log.Infof("min border is overcome - down")
	default:
		s.rsiPrev.maxBorderInProc = false
		s.rsiPrev.minBorderInProc = false
	}
	return err
}

func (s *Strategies) otherSignals(
	cfg *config.GlobalConfig, timestamp time.Time, size models.BinSize, closes []float64, step int,
) error {
	if step < cfg.GlobStrategies.GetCfgByBinSize(size.String()).GetCandlesCount {
		return nil
	}
	sigs, err := s.db.GetSignalsByTS(models.BolingerBand, size, []time.Time{timestamp})
	if err != nil {
		return err
	}
	if len(sigs) != 0 {
		s.log.Warnf("other signals by ts: %v, and binsize: %v already exist", timestamp, size)
		return nil
	}

	maStartIndx, maStopIndx := step-cfg.GlobStrategies.GetCfgByBinSize(size.String()).GetCandlesCount, step
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
