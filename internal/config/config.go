package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type GlobalConfig struct {
	ExchangesSettings ExchangesSettings
	Admin             Admin
	Scheduler         Scheduler
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

func parseConfig(cfgFile string) (*GlobalConfig, error) {
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
		Scheduler:        initSchedulers(),
		OrdProcPeriodSec: viper.GetInt("ord_proc_period_sec"),
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("global cfg: %#v\n", cfg)
	fmt.Println("--------------------------------------------")
	return cfg, nil
}
