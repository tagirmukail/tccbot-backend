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

// TODO добавить проверку на уже инициализированые по timestamp сигналы

func (s *Strategies) SignalsInit() error {
	for _, binSize := range s.cfg.Strategies.BinSizes {
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

	count := s.cfg.Strategies.MacdSlowCount + s.cfg.Strategies.MacdSigCount

	startTime, err := utils.FromTime(time.Now().UTC(), binSize, count)
	if err != nil {
		return err
	}

	candles, err := s.tradeApi.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Currency,
		BinSize:   binSize,
		Count:     int32(count),
		StartTime: startTime.Format(bitmex.TradeTimeFormat),
	})
	err = s.checkCloses(candles)
	if err != nil {
		return err
	}
	err = s.checkCloses(candles)
	if err != nil {
		return err
	}
	closes := s.fetchCloses(candles)
	if len(closes) < s.cfg.Strategies.MacdSlowCount {
		return errors.New("candles less than macd slow count")
	}
	var (
		prevMACDHistVals []float64
	)

	var step int
	for i := len(candles) - s.cfg.Strategies.MacdSigCount; i < len(candles); i++ {
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
		prevMACDHistVals, err = s.macdSave(timestamp, binType, closes, prevMACDHistVals, i)
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
	timestamp time.Time, size models.BinSize, closes []float64, prevMACDHistVals []float64, step int) ([]float64, error) {
	fastStartIndx, fastStopIndx := step-s.cfg.Strategies.MacdFastCount, step
	slowStartIndx, slowStopIndx := step-s.cfg.Strategies.MacdSlowCount, step
	fastValues := closes[fastStartIndx:fastStopIndx]
	slowValues := closes[slowStartIndx:slowStopIndx]
	macd := s.tradeCalc.CalculateMACD(
		fastValues,
		slowValues,
		prevMACDHistVals,
		trademath.EMAIndication,
	)
	_, err := s.db.SaveSignal(models.Signal{
		BinSize:            size,
		MACDFast:           s.cfg.Strategies.MacdFastCount,
		MACDSlow:           s.cfg.Strategies.MacdSlowCount,
		MACDSig:            s.cfg.Strategies.MacdSigCount,
		Timestamp:          timestamp,
		SignalType:         models.MACD,
		SignalValue:        macd.Sig,
		MACDHistogramValue: macd.HistogramValue,
	})
	if err != nil {
		return nil, err
	}
	prevMACDHistVals = append(prevMACDHistVals, macd.HistogramValue)
	return prevMACDHistVals, nil
}

func (s *Strategies) rsiSave(
	timestamp time.Time, size models.BinSize, closes []float64, step int) error {
	rsiStartIndx, rsiStopIndx := step-s.cfg.Strategies.RsiCount, step
	rsiValues := closes[rsiStartIndx:rsiStopIndx]
	rsi := s.tradeCalc.CalculateRSI(rsiValues, trademath.WMAIndication)
	_, err := s.db.SaveSignal(models.Signal{
		N:           len(rsiValues),
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})
	return err
}

func (s *Strategies) otherSignals(timestamp time.Time, size models.BinSize, closes []float64, step int) error {
	if step < s.cfg.Strategies.GetCandlesCount {
		return nil
	}

	maStartIndx, maStopIndx := step-s.cfg.Strategies.GetCandlesCount, step
	values := closes[maStartIndx:maStopIndx]
	signals := s.tradeCalc.CalculateSignals(values)
	_, err := s.db.SaveSignal(models.Signal{
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
