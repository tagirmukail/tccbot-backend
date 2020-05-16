package strategy

import (
	"errors"
	"fmt"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const limitMinOnOrderQty = 100

func (s *BBRSIStrategy) checkCloses(candles []bitmex.TradeBuck) error {
	for _, candle := range candles {
		if candle.Close == 0 {
			return errors.New("candles not full, exist empty close values")
		}
	}
	return nil
}

func (s *BBRSIStrategy) fetchCloses(candles []bitmex.TradeBuck) []float64 {
	var result []float64
	for _, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		result = append(result, candle.Close)
	}
	return result
}

func (s *BBRSIStrategy) getCandles(binSize models.BinSize) ([]bitmex.TradeBuck, error) {
	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(binSize.String())
	var count int
	if cfg.RsiCount > cfg.GetCandlesCount {
		count = cfg.RsiCount * 2
	} else {
		count = cfg.GetCandlesCount * 2
	}
	startTime, err := utils.FromTime(time.Now().UTC(), binSize.String(), count)
	if err != nil {
		return nil, err
	}
	candles := s.caches.GetCache(binSize).GetBucketed(startTime, time.Time{}, count)
	return candles, nil
}

func (s *BBRSIStrategy) fetchTsFromCandles(candles []bitmex.TradeBuck) ([]time.Time, error) {
	var result []time.Time
	for _, candle := range candles {
		candleTs, err := time.Parse(tradeapi.TradeBucketedTimestampLayout, candle.Timestamp)
		if err != nil {
			return nil, err
		}
		result = append(result, candleTs)
	}
	return result, nil
}

func (s *BBRSIStrategy) fetchLastCandlesForBB(binSize string, candles []bitmex.TradeBuck) []bitmex.TradeBuck {
	lastIndx := len(candles) - s.cfg.GlobStrategies.GetCfgByBinSize(binSize).BBLastCandlesCount
	if lastIndx < 0 {
		return nil
	}
	var result = candles[lastIndx:]
	return result
}

func (s *BBRSIStrategy) placeBitmexOrder(side types.Side, passive bool) error {
	balance, err := s.orderProc.GetBalance()
	if err != nil {
		s.log.Warnf("orderProc.GetBalance failed: %v", err)
		return err
	}
	contracts := trademath.ConvertFromBTCToContracts(balance)
	if contracts <= 3 {
		return fmt.Errorf("balance is exhausted, %.3f left", balance)
	}

	amount := utils.RandomRange(limitMinOnOrderQty, s.cfg.ExchangesSettings.Bitmex.MaxAmount)

	ord, err := s.orderProc.PlaceOrder(types.Bitmex, side, amount, passive)
	if err != nil {
		s.log.Warnf("orderProc.PlaceOrder failed: %v", err)
		return err
	}

	s.log.Infof("sell order placed: %#v", ord)
	return nil
}
