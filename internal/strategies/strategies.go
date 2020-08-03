package strategies

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	bitmextradedata "github.com/tagirmukail/tccbot-backend/internal/tradedata/bitmex"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	"github.com/tagirmukail/tccbot-backend/internal/strategies/strategy"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
)

type Strategies struct {
	wgRunner              *sync.WaitGroup
	cfg                   *config.GlobalConfig
	tradeApi              tradeapi.Api
	db                    db.DBManager
	log                   *logrus.Logger
	tradeCalc             trademath.Calc
	orderProc             *orderproc.OrderProcessor
	bitmexDataSender      *bitmextradedata.Sender
	bitmexTradeSubscriber *bitmextradedata.Subscriber
	initSignals           bool
	rsiPrev               struct {
		minBorderInProc bool
		maxBorderInProc bool
	}
	candlesCaches candlecache.Caches

	bbRsi strategy.Strategy
}

// TODO перенести все параметры в отдельную структуру
func New(
	wgRunner *sync.WaitGroup,
	cfg *config.GlobalConfig,
	tradeApi tradeapi.Api,
	orderProc *orderproc.OrderProcessor,
	bitmexDataSender *bitmextradedata.Sender,
	bitmexTradeSubscriber *bitmextradedata.Subscriber,
	db db.DBManager,
	log *logrus.Logger,
	initSignals bool,
	bbStrategy strategy.Strategy,
	candlesCaches candlecache.Caches,
) *Strategies {
	return &Strategies{
		wgRunner:              wgRunner,
		cfg:                   cfg,
		tradeApi:              tradeApi,
		orderProc:             orderProc,
		bitmexDataSender:      bitmexDataSender,
		bitmexTradeSubscriber: bitmexTradeSubscriber,
		db:                    db,
		log:                   log,
		tradeCalc:             trademath.Calc{},
		initSignals:           initSignals,
		bbRsi:                 bbStrategy,
		candlesCaches:         candlesCaches,
	}
}

func (s *Strategies) Start() {
	err := s.SignalsInit()
	if err != nil {
		s.log.Fatalf("SignalsInit failed: %v", err)
	}
	s.wgRunner.Add(1)
	go s.start()
	//s.wgRunner.Add(1)
	//go s.orderProc.Start(s.wgRunner)
	s.wgRunner.Add(1)
	go s.bitmexDataSender.SendToSubscribers(s.wgRunner)
	s.wgRunner.Wait()
}

func (s *Strategies) start() {
	defer s.wgRunner.Done()

	s.wgRunner.Add(1)
	go s.tradeApi.GetBitmex().GetWS().Start(s.wgRunner)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	s.log.Infof("process messages from bitmex started")
	for {
		select {
		case <-done:
			s.log.Infof("process messages stopped")
			return
		case data := <-s.bitmexTradeSubscriber.GetMsgChan():
			if len(data.Data) == 0 {
				s.log.Debug("empty data from ws")
				continue
			}
			switch data.Table {
			case string(types.TradeBin1m):
				if len(data.Data) != 0 {
					for _, candle := range data.Data {
						err := s.candlesCaches.GetCache(models.Bin1m).Store(candle)
						if err != nil {
							s.log.Warnf("store candle in cache failed: %v", err)
						}
					}
				}
				s.processStrategies("1m")
			case string(types.TradeBin5m):
				if len(data.Data) != 0 {
					for _, candle := range data.Data {
						err := s.candlesCaches.GetCache(models.Bin5m).Store(candle)
						if err != nil {
							s.log.Warnf("store candle in cache failed: %v", err)
						}
					}
				}
				s.processStrategies("5m")
			case string(types.TradeBin1h):
				if len(data.Data) != 0 {
					for _, candle := range data.Data {
						err := s.candlesCaches.GetCache(models.Bin1h).Store(candle)
						if err != nil {
							s.log.Warnf("store candle in cache failed: %v", err)
						}
					}
				}
				s.processStrategies("1h")
			case string(types.TradeBin1d):
				if len(data.Data) != 0 {
					for _, candle := range data.Data {
						err := s.candlesCaches.GetCache(models.Bin1d).Store(candle)
						if err != nil {
							s.log.Warnf("store candle in cache failed: %v", err)
						}
					}
				}
				s.processStrategies("1d")
			default:
				s.log.Warnf("processStrategies is not supported this trade bin: %v", data.Table)
				continue
			}
		}
	}
}

func (s *Strategies) processStrategies(binSize string) {
	bin, err := models.ToBinSize(binSize)
	if err != nil {
		s.log.Warnf("to bin size error: %v", err)
		return
	}
	strategiesConfig := s.cfg.GlobStrategies.GetCfgByBinSize(binSize)
	if strategiesConfig == nil {
		s.log.Warnf("strategies not installed for bin_size:%s", binSize)
		return
	}
	if !strategiesConfig.AnyStrategyEnabled() {
		s.log.Warnf("all strategies disabled for bin_size:%s", binSize)
		return
	}

	s.log.Infof("\n-------------------------------------\nstart strategies - bin size: %s", binSize)
	defer s.log.Infof("finished strategies - bin size: %s\n-------------------------------------", binSize)

	var currentStrategy strategy.Strategy

	if strategiesConfig.EnableRSIBB {
		currentStrategy = s.bbRsi
	}
	// todo add macd with rsi

	if currentStrategy != nil {
		err = currentStrategy.Execute(context.Background(), bin)
		if err != nil {
			s.log.Errorf("execute strategy failed: %v", err)
		}
	}
}
