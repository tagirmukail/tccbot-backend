package config

import (
	"database/sql"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/tagirmukail/tccbot-backend/internal/db"
)

type Configurator struct {
	cfgFile      string
	db           db.DatabaseManager
	mx           sync.Mutex
	cfg          *GlobalConfig
	updPeriodSec time.Duration
}

func NewConfigurator(cfgFile string, updPeriodSec time.Duration) (*Configurator, error) {
	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	updCfg := &Configurator{
		cfgFile:      cfgFile,
		mx:           sync.Mutex{},
		updPeriodSec: updPeriodSec,
		cfg:          cfg,
	}

	return updCfg, nil
}

func (u *Configurator) SetDB(db db.DatabaseManager) {
	u.db = db
}

func (u *Configurator) GetConfig() (*GlobalConfig, error) {
	u.mx.Lock()
	defer u.mx.Unlock()
	return u.cfg, nil
	//if u.cfg != nil && u.cfg.UpdatedAt.Add(u.updPeriodSec).After(time.Now()) {
	//	copyCfg := *u.cfg
	//	u.mx.Unlock()
	//	return &copyCfg, nil
	//}
	u.mx.Unlock()

	cfgM, err := u.db.GetConfiguration()
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}

		err = u.update()
		if err != nil {
			return nil, err
		}
		cfgM, err = u.db.GetConfiguration()
		if err != nil {
			return nil, err
		}
	}

	cfg := toGlobalConfig(*cfgM)

	return &cfg, nil
}

func (u *Configurator) SetConfig(cfg *GlobalConfig) {
	u.mx.Lock()
	u.cfg = cfg
	u.mx.Unlock()
}

func (u *Configurator) UpdateRun() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(u.updPeriodSec)
	defer tick.Stop()

	for {
		select {
		case <-done:
			return
		case <-tick.C:
			err := u.update()
			if err != nil {
				logrus.Fatal(err)
			}
		}
	}
}

func (u *Configurator) update() error {
	cfg, err := ParseConfig(u.cfgFile)
	if err != nil {
		return errors.Errorf("parse tccbot configuration file %s failed: %v", u.cfgFile, err)
	}

	u.mx.Lock()
	u.cfg = cfg
	u.cfg.UpdatedAt = time.Now()
	u.mx.Unlock()

	cfgM, err := toModelConfiguration(*cfg)
	if err != nil {
		return errors.Errorf("conversation configuration to model failed: file %s, error: %v", u.cfgFile, err)
	}

	err = u.db.SaveConfiguration(cfgM)
	if err != nil {
		return errors.Errorf("save configuration to database failed: file %s, error: %v", u.cfgFile, err)
	}

	return nil
}
