package strategy

import (
	"errors"
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const (
	limitMinOnOrderQty    = 100
	limitBalanceContracts = 200
)

func checkCloses(candles []bitmex.TradeBuck) error {
	for _, candle := range candles {
		if candle.Close == 0 {
			return errors.New("candles not full, exist empty close values")
		}
	}
	return nil
}

func fetchCloses(candles []bitmex.TradeBuck) []float64 {
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

func fetchTsFromCandles(candles []bitmex.TradeBuck) ([]time.Time, error) {
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

func getCandles(caches candlecache.Caches, binSize models.BinSize, count int) ([]bitmex.TradeBuck, error) {
	startTime, err := utils.FromTime(time.Now().UTC(), binSize.String(), count)
	if err != nil {
		return nil, err
	}
	candles := caches.GetCache(binSize).GetBucketed(startTime, time.Time{}, count)
	return candles, nil
}

func placeBitmexOrder(
	orderProc *orderproc.OrderProcessor, side types.Side, passive bool, log *logrus.Logger,
) error {
	ord, err := orderProc.PlaceOrder(types.Bitmex, side, 0, passive, true)
	if err != nil {
		log.Warnf("orderProc.PlaceOrder failed: %v", err)
		return err
	}

	log.Infof("%s order placed: %#v", side, ord)
	return nil
}

func findMaxAndMin(candles []bitmex.TradeBuck) (float64, float64) {
	var (
		max float64
		min = math.MaxFloat64
	)
	closes := fetchCloses(candles)
	for _, closeP := range closes {
		if closeP > max {
			max = closeP
		}
		if closeP < min {
			min = closeP
		}
	}
	return max, min
}
