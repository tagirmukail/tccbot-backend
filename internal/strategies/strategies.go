package strategies

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tagirmukail/tccbot-backend/internal/types"

	"github.com/tagirmukail/tccbot-backend/internal/orderproc"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
)

type Strategies struct {
	wgRunner    *sync.WaitGroup
	cfg         *config.GlobalConfig
	tradeApi    tradeapi.Api
	db          db.DBManager
	log         *logrus.Logger
	tradeCalc   trademath.Calc
	orderProc   *orderproc.OrderProcessor
	initSignals bool
	rsiPrev     struct {
		minBorderInProc bool
		maxBorderInProc bool
	}
	restarts struct {
		restart5m bool
		restart1h bool
	}
}

func New(
	wgRunner *sync.WaitGroup,
	cfg *config.GlobalConfig,
	tradeApi tradeapi.Api,
	orderProc *orderproc.OrderProcessor,
	db db.DBManager,
	log *logrus.Logger,
	initSignals bool,
) *Strategies {
	return &Strategies{
		wgRunner:    wgRunner,
		cfg:         cfg,
		tradeApi:    tradeApi,
		orderProc:   orderProc,
		db:          db,
		log:         log,
		tradeCalc:   trademath.Calc{},
		initSignals: initSignals,
	}
}

func (s *Strategies) Start() {
	if s.initSignals {
		err := s.SignalsInit()
		if err != nil {
			s.log.Fatalf("SignalsInit failed: %v", err)
		}
	}
	s.wgRunner.Add(1)
	go s.start()
	s.wgRunner.Add(1)
	go s.orderProc.Start(s.wgRunner)
	s.wgRunner.Wait()
}

func (s *Strategies) start() {
	defer s.wgRunner.Done()

	go s.tradeApi.GetBitmex().GetWS().Start()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
	s.log.Infof("process messages from bitmex started")
	for {
		select {
		case <-done:
			s.log.Infof("process messages stopped")
			return
		case data := <-s.tradeApi.GetBitmex().GetWS().GetMessages():
			if len(data.Data) == 0 {
				continue
			}
			switch data.Table {
			case string(types.TradeBin5m):
				s.processStrategies("5m", int32(s.cfg.Strategies.GetCandlesCount))
			case string(types.TradeBin1h):
				s.processStrategies("1h", int32(s.cfg.Strategies.GetCandlesCount))
			case string(types.TradeBin1d):
				s.processStrategies("1d", int32(s.cfg.Strategies.GetCandlesCount))
			default:
				s.log.Warnf("processStrategies is not supported this trade bin: %v", data.Table)
				continue
			}
		}
	}
}

func (s *Strategies) processStrategies(binSize string, count int32) {
	if s.cfg.Strategies.EnableBolingerBand {
		err := s.processBBStrategy(binSize, count)
		if err != nil {
			s.log.Errorf("processBBStrategy error: %v", err)
		}
	}
	if s.cfg.Strategies.EnableMACD {
		err := s.processMACDStrategy(binSize)
		if err != nil {
			s.log.Errorf("processMACDStrategy error: %v", err)
		}
	}
	if s.cfg.Strategies.EnableRSI {
		err := s.processRsiStrategy(binSize)
		if err != nil {
			s.log.Errorf("processRsiStrategy error: %v", err)
		}
	}
}
