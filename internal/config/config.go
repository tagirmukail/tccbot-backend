package config

import (
	"github.com/spf13/viper"
)

type GlobalConfig struct {
	ExchangesSettings ExchangesSettings
	Admin             Admin
	DB                DB
	Accesses          ExchangesAccess
	Strategies        StrategiesConfig
}

type DB struct {
	DBName   string
	User     string
	Password string
	Host     string
	Port     uint32
	SSLMode  string
}

type StrategiesConfig struct {
	RetryProcessCount  int
	EnableBolingerBand bool
	GetCandlesCount    int
	BBLastCandlesCount int
}

func (strategies *StrategiesConfig) AnyStrategyEnabled() bool {
	return strategies.EnableBolingerBand
}

type ExchangesSettings struct {
	Bitmex ApiSettings
}

type ExchangeSettings struct {
	Enable bool
	Api    ApiSettings
}

type ApiSettings struct {
	Test       bool
	PingSec    int
	TimeoutSec int
	RetrySec   int
	BufferSize int
	Currency   string
}

type BitmexCfg struct {
	Enable   bool
	Scheme   string
	Host     string
	Path     string
	Ping     int
	Timeout  int
	RetrySec uint32
	Test     BitmexTestCfg
}

type BitmexTestCfg struct {
	Host string
	Path string
}

type BinanceCfg struct {
	Enable   bool
	Scheme   string
	Host     string
	Path     string
	Ping     int
	Timeout  int
	RetrySec uint32
}

type Admin struct {
	Username    string `json:"username"`
	SecretToken string `json:"secret_token"`
	Token       string `json:"token"`
}

type ExchangesAccess struct {
	Bitmex Access `json:"bitmex"`
}

type Access struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func ParseConfig(cfgFile string) (*GlobalConfig, error) {
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var bitmex ApiSettings
	bitmexExist := viper.InConfig("exchanges_settings.bitmex")
	if !bitmexExist {
		// default
		bitmex = ApiSettings{
			Test:       true,
			PingSec:    20,
			TimeoutSec: 30,
			RetrySec:   5,
			BufferSize: 10,
			Currency:   "XBTUSD",
		}
	} else {
		bitmex = ApiSettings{
			Test:       viper.GetBool("exchanges_settings.bitmex.test"),
			PingSec:    viper.GetInt("exchanges_settings.bitmex..ping_sec"),
			TimeoutSec: viper.GetInt("exchanges_settings.bitmex.timeout_sec"),
			RetrySec:   viper.GetInt("exchanges_settings.bitmex.retry_sec"),
			BufferSize: viper.GetInt("exchanges_settings.bitmex.buffer_size"),
			Currency:   viper.GetString("exchanges_settings.bitmex.currency"),
		}
	}

	var strategies StrategiesConfig
	strategiesExist := viper.InConfig("strategies")
	if strategiesExist {
		strategies.RetryProcessCount = viper.GetInt("strategies.retry_process_count")
		strategies.GetCandlesCount = viper.GetInt("strategies.get_candles_count")
		strategies.BBLastCandlesCount = viper.GetInt("strategies.bb_last_candles_count")
		strategies.EnableBolingerBand = viper.GetBool("strategies.enable_bb")
	} else {
		// default
		strategies.RetryProcessCount = 5
		strategies.EnableBolingerBand = true
		strategies.GetCandlesCount = 20
		strategies.BBLastCandlesCount = 4
	}

	cfg := &GlobalConfig{
		ExchangesSettings: ExchangesSettings{
			Bitmex: bitmex,
		},
		Admin: Admin{
			Username:    viper.GetString("admin.username"),
			SecretToken: viper.GetString("admin.secret_token"),
		},
		Accesses: ExchangesAccess{
			Bitmex: Access{
				Key:    viper.GetString("exchanges_access.bitmex.key"),
				Secret: viper.GetString("exchanges_access.bitmex.secret"),
			},
		},
		DB: DB{
			DBName:   viper.GetString("db.name"),
			User:     viper.GetString("db.user"),
			Password: viper.GetString("db.password"),
			Host:     viper.GetString("db.host"),
			Port:     viper.GetUint32("db.port"),
			SSLMode:  viper.GetString("db.sslmode"),
		},
		Strategies: strategies,
	}
	return cfg, nil
}
