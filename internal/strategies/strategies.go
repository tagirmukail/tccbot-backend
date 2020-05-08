package strategies

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
)

type Strategies struct {
	wgRunner    *sync.WaitGroup
	cfg         *config.GlobalConfig
	tradeApi    *tradeapi.TradeApi
	db          db.DBManager
	log         *logrus.Logger
	tradeCalc   trademath.Calc
	initSignals bool
}

func New(
	wgRunner *sync.WaitGroup,
	cfg *config.GlobalConfig,
	tradeApi *tradeapi.TradeApi,
	db db.DBManager,
	log *logrus.Logger,
	initSignals bool,
) *Strategies {
	return &Strategies{
		wgRunner:    wgRunner,
		cfg:         cfg,
		tradeApi:    tradeApi,
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
	s.wgRunner.Wait()
}

func (s *Strategies) start() {
	defer s.wgRunner.Done()

	wg := &sync.WaitGroup{}
	for _, binSize := range s.cfg.Strategies.BinSizes {
		switch binSize {
		case "5m":
			wg.Add(1)
			go s.process5mCandles(wg)
		case "1h":
			wg.Add(1)
			go s.process1hCandles(wg)
		case "15m":
			//wg.Add(1)
			//go s.process15Candles(wg)
		case "1d":
		// TODO add
		default:
			s.log.Fatalf("unknown bin_size: %s", binSize)
		}

	}
	wg.Wait()
}

func (s *Strategies) process5mCandles(wg *sync.WaitGroup) {
	defer wg.Done()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	// for debug used 1m
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			s.processStrategies("5m", int32(s.cfg.Strategies.GetCandlesCount))
		case <-done:
			s.log.Infof("5m candles process stopped")
			return
		}
	}
}

// TODO fix for 15 min candles
func (s *Strategies) process15Candles(wg *sync.WaitGroup) {
	defer wg.Done()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(15 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			s.processStrategies("15m", int32(s.cfg.Strategies.GetCandlesCount))
		case <-done:
			s.log.Infof("15m candles process stopped")
			return
		}
	}
}

func (s *Strategies) process1hCandles(wg *sync.WaitGroup) {
	defer wg.Done()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(1 * time.Hour)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			s.processStrategies("1h", int32(s.cfg.Strategies.GetCandlesCount))
		case <-done:
			s.log.Infof("1h candles process stopped")
			return
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
}
