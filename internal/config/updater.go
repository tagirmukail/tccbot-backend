package config

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type UpdConfigurator struct {
	mx  sync.RWMutex
	cfg *GlobalConfig
}

func NewUpdConfigurator(cfgFile string) (*UpdConfigurator, error) {
	cfg, err := parseConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	updCfg := &UpdConfigurator{
		mx:  sync.RWMutex{},
		cfg: cfg,
	}

	return updCfg, nil
}

func (u *UpdConfigurator) GetConfig() *GlobalConfig {
	u.mx.RLock()
	u.mx.RUnlock()
	return u.cfg
}

func (u *UpdConfigurator) UpdateRun(periodInSec time.Duration, cfgFile string) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(periodInSec)
	defer tick.Stop()

	for {
		select {
		case <-done:
			return
		case <-tick.C:
			var err error
			cfg, err := parseConfig(cfgFile)
			if err != nil {
				logrus.Fatalf("parse tccbot configuration file %s failed: %v", cfgFile, err)
			}
			u.mx.Lock()
			u.cfg = cfg
			u.mx.Unlock()
		}
	}
}
