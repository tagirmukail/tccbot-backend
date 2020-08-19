package strategy

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
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
	var result = make([]float64, 0, len(candles))
	for _, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		result = append(result, candle.Close)
	}
	return result
}

func (s *BBRSIStrategy) getCandles(scfg *config.GlobalConfig, binSize models.BinSize) ([]bitmex.TradeBuck, error) {
	cfg := scfg.GlobStrategies.GetCfgByBinSize(binSize.String())
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

func fetchTSFromCandles(candles []bitmex.TradeBuck) ([]time.Time, error) {
	var result = make([]time.Time, 0, len(candles))
	for _, candle := range candles {
		candleTS, err := time.Parse(tradeapi.TradeBucketedTimestampLayout, candle.Timestamp)
		if err != nil {
			return nil, err
		}
		result = append(result, candleTS)
	}
	return result, nil
}

func (s *BBRSIStrategy) fetchLastCandlesForBB(
	scfg *config.GlobalConfig, binSize string, candles []bitmex.TradeBuck,
) []bitmex.TradeBuck {
	lastIndx := len(candles) - scfg.GlobStrategies.GetCfgByBinSize(binSize).BBLastCandlesCount
	if lastIndx < 0 {
		return nil
	}
	var result = candles[lastIndx:]
	return result
}

func getCandles( // nolint:unused,deadcode
	caches candlecache.Caches, binSize models.BinSize, count int,
) ([]bitmex.TradeBuck, error) {
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
	ord, err := orderProc.PlaceOrder(types.Bitmex, side, 0, passive)
	if err != nil {
		log.Warnf("orderProc.PlaceOrder failed: %v", err)
		return err
	}

	log.Infof("%s order placed: %#v", side, ord)
	return nil
}
