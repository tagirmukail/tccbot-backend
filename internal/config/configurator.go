package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

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

	cfgurtr := &Configurator{
		cfgFile:      cfgFile,
		mx:           sync.Mutex{},
		updPeriodSec: updPeriodSec,
	}

	err := cfgurtr.parseConfig(cfgFile)
	if err != nil {
		logrus.Fatal(err)
	}

	return cfgurtr, nil
}

func (u *Configurator) SetDB(db db.DatabaseManager) {
	u.db = db
}

func (u *Configurator) GetConfig() (*GlobalConfig, error) {
	u.mx.Lock()
	defer u.mx.Unlock()
	return u.cfg, nil
}

func (u *Configurator) Update() error {
	cfgM, err := toModelConfiguration(*u.cfg)
	if err != nil {
		return errors.Errorf("conversation configuration to model failed: file %s, error: %v", u.cfgFile, err)
	}

	err = u.db.SaveConfiguration(cfgM)
	if err != nil {
		return errors.Errorf("save configuration to database failed: file %s, error: %v", u.cfgFile, err)
	}

	return nil
}

func (u *Configurator) parseConfig(cfgFile string) error {
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		u.mx.Lock()
		defer u.mx.Unlock()
		u.cfg = &GlobalConfig{
			GlobStrategies:    initGlobStrategies(),
			ExchangesSettings: initExchangesAPI(),
			Admin: Admin{
				Username:    viper.GetString("admin.username"),
				SecretToken: viper.GetString("admin.secret_token"),
			},
			Accesses:         initExchangesAccesses(),
			DBPath:           initDBPath(),
			Scheduler:        initSchedulers(),
			OrdProcPeriodSec: viper.GetInt("ord_proc_period_sec"),
		}
		err := u.Update()
		if err != nil {
			logrus.Fatal(err)
		}
	})

	u.mx.Lock()
	defer u.mx.Unlock()
	u.cfg = &GlobalConfig{
		GlobStrategies:    initGlobStrategies(),
		ExchangesSettings: initExchangesAPI(),
		Admin: Admin{
			Username:    viper.GetString("admin.username"),
			SecretToken: viper.GetString("admin.secret_token"),
		},
		Accesses:         initExchangesAccesses(),
		DBPath:           initDBPath(),
		Scheduler:        initSchedulers(),
		OrdProcPeriodSec: viper.GetInt("ord_proc_period_sec"),
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("global cfg: %#v\n", u.cfg)
	fmt.Println("--------------------------------------------")

	return nil
}
