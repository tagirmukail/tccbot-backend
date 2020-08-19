package strategies

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (s *Strategies) fetchCloses(candles []bitmex.TradeBuck) []float64 {
	var result = make([]float64, 0, len(candles))
	for _, candle := range candles {
		if candle.Close == 0 {
			continue
		}
		result = append(result, candle.Close)
	}
	return result
}

func (s *Strategies) fetchTSFromCandles(candles []bitmex.TradeBuck) ([]time.Time, error) { // nolint:unused
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

func (s *Strategies) checkCloses(candles []bitmex.TradeBuck) error {
	for _, candle := range candles {
		if candle.Close == 0 {
			return errors.New("candles not full, exist empty close values")
		}
	}
	return nil
}

func (s *Strategies) fetchMacdVals(signals []*models.Signal) []float64 { // nolint:unused
	var result = make([]float64, 0, len(signals))
	for _, signal := range signals {
		if signal.MACDValue == 0 {
			continue
		}
		result = append(result, signal.MACDValue)
	}
	return result
}

func (s *Strategies) macdTimeFrameDefine(size models.BinSize) int { // nolint:unused
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
