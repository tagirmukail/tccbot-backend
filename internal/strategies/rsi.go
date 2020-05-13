package strategies

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/types"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) processRsiStrategy(binSize string) error {
	s.log.Infof("start process rsi")
	defer s.log.Infof("finish process rsi")
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		return err
	}

	count := s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiCount * 2
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
	closes := s.fetchCloses(candles)
	if len(closes) < s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiCount {
		return errors.New("candles less than rsi count")
	}

	timestamp, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		return err
	}

	rsi := s.tradeCalc.CalcRSI(closes, s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiCount)
	_, err = s.db.SaveSignal(models.Signal{
		N:           s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiCount,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})
	if err != nil {
		return err
	}

	if rsi.Value >= float64(s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiMaxBorder) {
		s.rsiPrev.maxBorderInProc = true
		s.rsiPrev.minBorderInProc = false
		s.log.Infof("max border is overcome up")
	} else if rsi.Value <= float64(s.cfg.GlobStrategies.GetCfgByBinSize(binSize).RsiMinBorder) {
		s.rsiPrev.minBorderInProc = true
		s.rsiPrev.maxBorderInProc = false
		s.log.Infof("min border is overcome - down")
	} else {
		if s.rsiPrev.maxBorderInProc {
			s.log.Infof("max border is overcome - down - place sell order")
			err = s.placeBitmexOrder(types.SideSell, models.RSI)
			if err != nil {
				return err
			}
			s.rsiPrev.maxBorderInProc = false
		}
		if s.rsiPrev.minBorderInProc {
			s.log.Infof("min border is overcome - up - place buy order")
			err = s.placeBitmexOrder(types.SideBuy, models.RSI)
			if err != nil {
				return err
			}
			s.rsiPrev.minBorderInProc = false
		}
	}

	return nil

}
