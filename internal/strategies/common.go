package strategies

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) fetchCloses(candles []bitmex.TradeBuck) []float64 {
	var result []float64
	for _, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		result = append(result, candle.Close)
	}
	return result
}

func (s *Strategies) fetchTsFromCandles(candles []bitmex.TradeBuck) ([]time.Time, error) {
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

func (s *Strategies) fetchLastCandlesForBB(candles []bitmex.TradeBuck) []bitmex.TradeBuck {
	lastIndx := len(candles) - s.cfg.Strategies.BBLastCandlesCount
	if lastIndx < 0 {
		return nil
	}
	var result = candles[lastIndx:]
	return result
}

func (s *Strategies) retryProcess(
	candles []bitmex.TradeBuck,
	binSize models.BinSize,
	retryFunc func(candles []bitmex.TradeBuck, binSize models.BinSize) error,
) error {
	var err error
	for i := 0; i < s.cfg.Strategies.RetryProcessCount; i++ {
		err = retryFunc(candles, binSize)
		if err == nil {
			break
		}
		s.log.Errorf("retryProcess error: %v", err)
	}
	return err
}

func (s *Strategies) checkCloses(candles []bitmex.TradeBuck) error {
	for _, candle := range candles {
		if candle.Close == 0 {
			return errors.New("candles not full, exist empty close values")
		}
	}
	return nil
}

func (s *Strategies) fetchMacdHistVals(signals []*models.Signal) []float64 {
	var result []float64
	for _, signal := range signals {
		if signal.MACDHistogramValue == 0 {
			continue
		}
		result = append(result, signal.MACDHistogramValue)
	}
	return result
}

func (s *Strategies) macdTimeFrameDefine(size models.BinSize) int {
	switch size {
	case models.Bin5m:
		return 6
	case models.Bin1h:
		return 4
	case models.Bin1d:
		return 3
	default:
		return 0
	}
}
