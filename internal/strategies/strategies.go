package strategies

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/utils"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type Strategies struct {
	wgRunner  *sync.WaitGroup
	cfg       *config.GlobalConfig
	tradeApi  *tradeapi.TradeApi
	db        db.DBManager
	log       *logrus.Logger
	tradeCalc trademath.Calc
}

func New(
	wgRunner *sync.WaitGroup,
	cfg *config.GlobalConfig,
	tradeApi *tradeapi.TradeApi,
	db db.DBManager,
	log *logrus.Logger,
) *Strategies {
	return &Strategies{
		wgRunner: wgRunner,
		cfg:      cfg,
		tradeApi: tradeApi,
		db:       db,
		log:      log,
	}
}

func (s *Strategies) Start() {
	s.wgRunner.Add(1)
	go s.start()
	s.wgRunner.Wait()
}

func (s *Strategies) start() {
	defer s.wgRunner.Done()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.process5mCandles(wg)
	//wg.Add(1)
	//go s.process15Candles(wg)
	//wg.Add(1)
	//go s.process1hCandles(wg)
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
			err := s.processTrade("5m", int32(s.cfg.Strategies.GetCandlesCount))
			if err != nil {
				s.log.Errorf("processTrade error: %v", err)
				return
			}
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
			err := s.processTrade("15", int32(s.cfg.Strategies.GetCandlesCount))
			if err != nil {
				s.log.Errorf("processTrade error: %v", err)
				return
			}
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

	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := s.processTrade("1h", 30)
			if err != nil {
				s.log.Errorf("processTrade error: %v", err)
				return
			}
		case <-done:
			s.log.Infof("1h candles process stopped")
			return
		}
	}
}

func (s *Strategies) processTrade(binSize string, count int32) error {
	var (
		err     error
		signals *trademath.Singals
		candles []bitmex.TradeBuck
	)
	bin := models.ToBinSize(binSize)
	if bin == 0 {
		return fmt.Errorf("bin size: %v unknown", binSize)
	}

	for i := 0; i < 5; i++ {
		candles, signals, err = s.processCandles(binSize, count)
		if err != nil {
			continue
		}
		break
	}

	err = s.processSignals(candles, signals, bin)
	if err != nil {
		return err
	}

	return nil
}

func (s *Strategies) processCandles(binSize string, count int32) ([]bitmex.TradeBuck, *trademath.Singals, error) {
	bin := models.ToBinSize(binSize)
	if bin == 0 {
		return nil, nil, fmt.Errorf("processCandles bin size: %v unknown", binSize)
	}
	from, err := utils.FromTime(time.Now().UTC(), binSize, int(count))
	if err != nil {
		return nil, nil, err
	}
	candles, err := s.tradeApi.GetBitmex().GetTradeBucketed(&bitmex.TradeGetBucketedParams{
		Symbol:    s.cfg.ExchangesSettings.Bitmex.Currency,
		BinSize:   binSize,
		Count:     count,
		StartTime: from.Format(bitmex.TradeTimeFormat),
	})
	if err != nil {
		s.log.Errorf("processCandles tradeApi.GetBitmex().GetTradeBucketed error: %v", err)
		return nil, nil, err
	}
	if len(candles) < s.cfg.Strategies.BBLastCandlesCount {
		s.log.Debug("processCandles there are fewer candles than necessary for the signal bolinger band")
		return nil, nil, errors.New("there are fewer candles than necessary for the signal bolinger band")
	}
	closes := s.fetchCloses(candles)
	if len(closes) == 0 {
		s.log.Debug("processCandles closes is empty")
		return nil, nil, errors.New("all candles invalid")
	}
	lastCandleTs, err := time.Parse(
		tradeapi.TradeBucketedTimestampLayout,
		candles[len(candles)-1].Timestamp,
	)
	if err != nil {
		s.log.Debugf("processCandles last candle timestamp parse error: %v", err)
		return nil, nil, err
	}
	signals := s.tradeCalc.CalculateSignals(closes)
	_, err = s.db.SaveSignal(models.Signal{
		N:         int(count),
		BinSize:   bin,
		Timestamp: lastCandleTs,
		SMA:       signals.SMA,
		WMA:       signals.WMA,
		EMA:       signals.EMA,
		BBTL:      signals.BB.TL,
		BBML:      signals.BB.ML,
		BBBL:      signals.BB.BL,
	})
	if err != nil {
		s.log.Debugf("processCandles db.SaveSignal error: %v", err)
		return nil, nil, err
	}

	s.log.Debugf("processed signals - close: %v for bin size:%s - signals: %#v", candles[len(candles)-1].Close, binSize, signals)
	return candles, &signals, err
}

func (s *Strategies) processSignals(
	candles []bitmex.TradeBuck, signals *trademath.Singals, binSize models.BinSize,
) error {
	s.log.Infof("start process signals for bin_size:%v", binSize)
	if s.cfg.Strategies.EnableBolingerBand {
		s.log.Infof("start process bolinger band signals for bin_size:%v", binSize)
		err := s.retryProcess(candles, binSize, s.processBB)
		if err != nil {
			return err
		}
	}

	return nil
}
