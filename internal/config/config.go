package config

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type GlobalConfig struct {
	mx                sync.RWMutex
	ExchangesSettings ExchangesSettings
	Admin             Admin
	Accesses          ExchangesAccess
	GlobStrategies    StrategiesGlobConfig
	OrdProcPeriodSec  int
	DBPath            string
}

type Admin struct {
	Username    string `json:"username"`
	SecretToken string `json:"secret_token"`
	Token       string `json:"token"`
}

func ParseConfig(cfgFile string) (*GlobalConfig, error) {
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &GlobalConfig{
		GlobStrategies:    initGlobStrategies(),
		ExchangesSettings: initExchangesApi(),
		Admin: Admin{
			Username:    viper.GetString("admin.username"),
			SecretToken: viper.GetString("admin.secret_token"),
		},
		Accesses:         initExchangesAccesses(),
		DBPath:           initDBPath(),
		OrdProcPeriodSec: viper.GetInt("ord_proc_period_sec"),
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("global cfg: %#v\n", cfg)
	fmt.Println("--------------------------------------------")
	return cfg, nil
}

func (cfg *GlobalConfig) RealTimeUpdateConfig(periodInSec time.Duration, cfgFile string) {
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
			cfg.mx.Lock()
			cfg, err = ParseConfig(cfgFile)
			cfg.mx.Unlock()
			if err != nil {
				logrus.Fatalf("parse tccbot configuration file %s failed: %v", cfgFile, err)
			}
		}
	}
}

// Get - WARN use only this method for get configuration
func (cfg *GlobalConfig) Get() *GlobalConfig {
	cfg.mx.RLock()
	defer cfg.mx.RUnlock()
	return cfg
}
