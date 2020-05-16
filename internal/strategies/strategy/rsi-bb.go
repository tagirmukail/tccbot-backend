package strategy

import (
	"context"
	"errors"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type BBRSIStrategy struct {
	cfg        *config.GlobalConfig
	api        tradeapi.Api
	math       trademath.Calc
	orderProc  *orderproc.OrderProcessor
	log        *logrus.Logger
	db         db.DBManager
	bbInAction AcceptAction
}

func NewBBRSIStrategy(
	cfg *config.GlobalConfig,
	api tradeapi.Api,
	orderProc *orderproc.OrderProcessor,
	db db.DBManager,
	log *logrus.Logger,
) *BBRSIStrategy {
	return &BBRSIStrategy{
		cfg:        cfg,
		api:        api,
		orderProc:  orderProc,
		db:         db,
		log:        log,
		bbInAction: NotAccepted,
		math:       trademath.Calc{},
	}
}

func (s *BBRSIStrategy) Execute(_ context.Context, size models.BinSize) error {
	s.log.Infof("start execute bb rsi strategy")
	defer s.log.Infof("finish execute bb rsi strategy")
	candles, err := s.getCandles(size.String())
	if err != nil {
		return err
	}

	acceptAction, err := s.processRsi(candles, size)
	if err != nil {
		return err
	}
	if acceptAction == NotAccepted {
		return nil
	}

	acceptAction, err = s.processBB(candles, size)
	if err != nil {
		return err
	}

	if acceptAction == NotAccepted {
		switch s.bbInAction {
		case UpAccepted:
			return s.placeBitmexOrder(types.SideSell, true)
		case DownAccepted:
			return s.placeBitmexOrder(types.SideBuy, true)
		default:
			return nil
		}
	}
	s.bbInAction = NotAccepted
	return nil
}

func (s *BBRSIStrategy) processRsi(candles []bitmex.TradeBuck, size models.BinSize) (AcceptAction, error) {
	s.log.Infof("start process rsi signal")
	defer s.log.Infof("finish process rsi signal")

	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		return NotAccepted, errors.New("cfg by bin size is empty")
	}

	err := s.checkCloses(candles)
	if err != nil {
		return NotAccepted, err
	}
	closes := s.fetchCloses(candles)
	if len(closes) < cfg.RsiCount {
		return NotAccepted, errors.New("candles less than rsi count")
	}

	timestamp, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		return NotAccepted, err
	}

	rsi := s.math.CalcRSI(closes, cfg.RsiCount)
	_, err = s.db.SaveSignal(models.Signal{
		N:           cfg.RsiCount,
		BinSize:     size,
		Timestamp:   timestamp,
		SignalType:  models.RSI,
		SignalValue: rsi.Value,
	})
	if err != nil {
		return NotAccepted, err
	}

	if rsi.Value >= float64(cfg.RsiMaxBorder) {
		s.log.Infof("max border is overcome up")
		return UpAccepted, nil
	} else if rsi.Value <= float64(cfg.RsiMinBorder) {
		s.log.Infof("min border is overcome - down")
		return DownAccepted, nil
	}

	return NotAccepted, nil
}

func (s *BBRSIStrategy) processBB(candles []bitmex.TradeBuck, size models.BinSize) (AcceptAction, error) {
	cfg := s.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		return NotAccepted, errors.New("cfg by bin size is empty")
	}
	if len(candles) < cfg.BBLastCandlesCount {
		s.log.Debug("processBBStrategyCandles there are fewer candles than necessary for the signal bolinger band")
		return NotAccepted, errors.New("there are fewer candles than necessary for the signal bolinger band")
	}
	closes := s.fetchCloses(candles)
	if len(closes) == 0 {
		s.log.Debug("processBBStrategyCandles closes is empty")
		return NotAccepted, errors.New("all candles invalid")
	}
	lastCandleTs, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		s.log.Debugf("processBBStrategyCandles last candle timestamp parse error: %v", err)
		return NotAccepted, err
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
		return NotAccepted, err
	}
	lastCandles := s.fetchLastCandlesForBB(size.String(), candles)
	if len(lastCandles) == 0 {
		err := errors.New("processBB last candles fo BB signal is empty")
		s.log.Debug(err)
		return NotAccepted, err
	}

	lastCandlesTs, err := s.fetchTsFromCandles(lastCandles)
	if err != nil {
		return NotAccepted, err
	}

	lastSignals, err := s.db.GetSignalsByTs([]models.SignalType{models.BolingerBand}, []models.BinSize{size}, lastCandlesTs)
	if err != nil {
		return NotAccepted, err
	}

	return s.processTrend(size, lastCandles, lastSignals), nil
}

func (s *BBRSIStrategy) processTrend(binSize models.BinSize, candles []bitmex.TradeBuck, signals []*models.Signal) AcceptAction {
	candlesCount, signalsCount := len(candles), len(signals)
	if candlesCount != signalsCount || candlesCount == 0 || signalsCount == 0 {
		s.log.Debugf("processTrend - count candles:%d, but signals:%d", candlesCount, signalsCount)
		return NotAccepted
	}

	s.log.Infof("start process bolinger band trend")
	firstCandle, firstSignal := candles[0], signals[0]
	if firstCandle.Close > firstSignal.BBTL {
		// uptrend detection
		if s.upTrend(binSize, candles[1:], signals[1:]) {
			s.bbInAction = UpAccepted
			return UpAccepted
		}
	}
	if firstCandle.Close < firstSignal.BBBL {
		// downtrend detection
		if s.downTrend(binSize, candles[1:], signals[1:]) {
			s.bbInAction = DownAccepted
			return DownAccepted
		}
	}
	s.log.Infof("finish process bolinger band trend")
	return NotAccepted
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
