package strategies

import (
	"errors"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

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

	lastSignals, err := s.db.GetSignalsByTs(binSize, lastCandlesTs)
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
	// todo sell order

}

func (s *Strategies) downTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) {
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
	// todo buy order
}
