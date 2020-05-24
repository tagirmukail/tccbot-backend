package strategy

import (
	"context"
	"errors"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"
	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	"github.com/tagirmukail/tccbot-backend/internal/strategies/filter"
	stratypes "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type BBRSIStrategy struct {
	cfg       *config.GlobalConfig
	api       tradeapi.Api
	math      trademath.Calc
	orderProc *orderproc.OrderProcessor
	log       *logrus.Logger
	db        db.DBManager
	caches    candlecache.Caches
	filters   []filter.Filter
}

func NewBBRSIStrategy(
	cfg *config.GlobalConfig,
	api tradeapi.Api,
	orderProc *orderproc.OrderProcessor,
	db db.DBManager,
	caches candlecache.Caches,
	log *logrus.Logger,
	filters ...filter.Filter,
) *BBRSIStrategy {
	return &BBRSIStrategy{
		cfg:       cfg,
		api:       api,
		orderProc: orderProc,
		db:        db,
		log:       log,
		math:      trademath.Calc{},
		caches:    caches,
		filters:   filters,
	}
}

func (s *BBRSIStrategy) Execute(_ context.Context, size models.BinSize) error {
	s.log.Infof("start execute bb rsi strategy")
	defer s.log.Infof("finish execute bb rsi strategy")

	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		return errors.New("cfg by bin size is empty")
	}

	candles, err := s.getCandles(size)
	if err != nil {
		return err
	}

	rsi, err := s.processRsi(candles, size)
	if err != nil {
		return err
	}

	_, err = s.processBB(candles, size)
	if err != nil {
		return err
	}

	var (
		lastCandles   []bitmex.TradeBuck
		lastCandlesTs []time.Time
		lastSignals   []*models.Signal
		action        stratypes.Action
	)
	if rsi.Value >= float64(cfg.RsiMaxBorder) || rsi.Value <= float64(cfg.RsiMinBorder) {
		lastCandles = s.fetchLastCandlesForBB(size.String(), candles)
		if len(lastCandles) == 0 {
			err := errors.New("processBB last candles fo BB signal is empty")
			s.log.Debug(err)
			return err
		}

		lastCandlesTs, err = fetchTsFromCandles(lastCandles)
		if err != nil {
			return err
		}

		lastSignals, err = s.db.GetSignalsByTs(models.BolingerBand, size, lastCandlesTs)
		if err != nil {
			return err
		}

		action = s.processTrend(size, lastCandles, lastSignals)
	}

	applySide := s.ApplyFilters(action, candles)
	switch applySide {
	case types.SideSell:
		return placeBitmexOrder(s.cfg, s.orderProc, types.SideSell, true, s.log)
	case types.SideBuy:
		return placeBitmexOrder(s.cfg, s.orderProc, types.SideBuy, true, s.log)
	default:
		return nil
	}
}

func (s *BBRSIStrategy) ApplyFilters(action stratypes.Action, candles []bitmex.TradeBuck) types.Side {
	if len(s.filters) == 0 {
		s.log.Debugf("filters is empty")
		return types.SideEmpty
	}
	ctx := context.WithValue(context.Background(), "action", action)
	ctx = context.WithValue(ctx, "candles", candles)
	applySide := types.SideEmpty
	for _, f := range s.filters {
		applySide = f.Apply(ctx)
		if applySide == types.SideEmpty {
			return applySide
		}
	}
	return applySide
}

func (s *BBRSIStrategy) processRsi(candles []bitmex.TradeBuck, size models.BinSize) (rsi trademath.RSI, err error) {
	s.log.Infof("start process rsi signal")
	defer s.log.Infof("finish process rsi signal")

	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		return rsi, errors.New("cfg by bin size is empty")
	}

	err = checkCloses(candles)
	if err != nil {
		return rsi, err
	}
	closes := fetchCloses(candles)
	if len(closes) < cfg.RsiCount {
		return rsi, errors.New("candles less than rsi count")
	}

	timestamp, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		return rsi, err
	}

	rsi = s.math.CalcRSI(closes, cfg.RsiCount)
	_, err = s.db.SaveSignal(models.Signal{
		N:           cfg.RsiCount,
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})
	if err != nil {
		return rsi, err
	}
	return rsi, nil
}

func (s *BBRSIStrategy) processBB(
	candles []bitmex.TradeBuck, size models.BinSize,
) (bb trademath.BB, err error) {
	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		return bb, errors.New("cfg by bin size is empty")
	}
	if len(candles) < cfg.BBLastCandlesCount {
		s.log.Debug("processBBStrategyCandles there are fewer candles than necessary for the signal bolinger band")
		return bb, errors.New("there are fewer candles than necessary for the signal bolinger band")
	}
	closes := fetchCloses(candles)
	if len(closes) == 0 {
		s.log.Debug("processBBStrategyCandles closes is empty")
		return bb, errors.New("all candles invalid")
	}
	lastCandleTs, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		s.log.Debugf("processBBStrategyCandles last candle timestamp parse error: %v", err)
		return bb, err
	}
	tl, ml, bl := s.math.CalcBB(closes, talib.EMA)
	_, err = s.db.SaveSignal(models.Signal{
		N:          cfg.GetCandlesCount,
		BinSize:    size,
		Timestamp:  lastCandleTs,
		SignalType: models.BolingerBand,
		BBTL:       tl,
		BBML:       ml,
		BBBL:       bl,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal bolinger band error: %v", err)
		return bb, err
	}
	bb.TL = tl
	bb.ML = ml
	bb.BL = bl
	return bb, nil
}

func (s *BBRSIStrategy) processTrend(
	binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal,
) stratypes.Action {
	candlesCount, signalsCount := len(candles), len(signals)
	if candlesCount != signalsCount || candlesCount == 0 || signalsCount == 0 {
		s.log.Debugf("processTrend - count candles:%d, but signals:%d", candlesCount, signalsCount)
		return stratypes.NotTrend
	}

	s.log.Infof("start process bolinger band trend")
	firstCandle, firstSignal := candles[0], signals[0]
	if firstCandle.Close > firstSignal.BBTL {
		// uptrend detection
		if s.upTrend(binSize, candles[1:], signals[1:]) {
			return stratypes.UpTrend
		}
	}
	if firstCandle.Close < firstSignal.BBBL {
		// downtrend detection
		if s.downTrend(binSize, candles[1:], signals[1:]) {
			return stratypes.DownTrend
		}
	}
	s.log.Infof("finish process bolinger band trend")
	return stratypes.NotTrend
}

func (s *BBRSIStrategy) upTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) bool {
	s.log.Infof("start process bolinger band up trend")

	if len(candles) == 0 || len(signals) == 0 {
		s.log.Debugf("upTrend empty candles or signals")
		return false
	}
	for i, candle := range candles {
		if candle.Close <= signals[i].BBTL {
			// uptrend broken
			s.log.Debugf("uptrend broken, #bin_size:%v, #close:%v, #bbtl:%v", binSize, candle.Close, signals[i].BBTL)
			return false
		}
	}

	s.log.Infof("uptrend successfully completed for bin_size:%v and close:%v", binSize, candles[len(candles)-1].Close)
	return true
}

func (s *BBRSIStrategy) downTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) bool {
	s.log.Infof("start process bolinger band down trend")

	if len(candles) == 0 || len(signals) == 0 {
		s.log.Debugf("downTrend empty candles or signals")
		return false
	}
	for i, candle := range candles {
		if candle.Close >= signals[i].BBBL {
			// downtrend broken
			s.log.Debugf("downtrend broken, #bin_size:%v, #close:%v, #bbtl:%v", binSize, candle.Close, signals[i].BBTL)
			return false
		}
	}
	s.log.Infof("downtrend successfully completed for bin_size:%v and close:%v", binSize, candles[len(candles)-1].Close)
	return true
}
