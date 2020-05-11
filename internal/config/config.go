package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tagirmukail/tccbot-backend/internal/types"
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
	EnableBolingerBand bool
	EnableMACD         bool
	EnableRSI          bool
	BinSizes           []string
	RetryProcessCount  int
	GetCandlesCount    int
	BBLastCandlesCount int
	RsiCount           int
	RsiMinBorder       uint32
	RsiMaxBorder       uint32
	MacdFastCount      int
	MacdSlowCount      int
	MacdSigCount       int
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
	Symbol     string
	OrderType  types.OrderType
	MaxAmount  float64
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
	bitmexSettings := viper.GetStringMap("exchanges_settings.bitmex")
	if len(bitmexSettings) == 0 {
		// default
		bitmex = ApiSettings{
			Test:       true,
			PingSec:    20,
			TimeoutSec: 30,
			RetrySec:   5,
			BufferSize: 10,
			Currency:   "XBt",
			Symbol:     "XBTUSD",
			OrderType:  types.Limit,
			MaxAmount:  10,
		}
	} else {
		bitmex = ApiSettings{
			Test:       viper.GetBool("exchanges_settings.bitmex.test"),
			PingSec:    viper.GetInt("exchanges_settings.bitmex..ping_sec"),
			TimeoutSec: viper.GetInt("exchanges_settings.bitmex.timeout_sec"),
			RetrySec:   viper.GetInt("exchanges_settings.bitmex.retry_sec"),
			BufferSize: viper.GetInt("exchanges_settings.bitmex.buffer_size"),
			Symbol:     viper.GetString("exchanges_settings.bitmex.symbol"),
			Currency:   viper.GetString("exchanges_settings.bitmex.currency"),
			OrderType:  types.OrderType(viper.GetString("exchanges_settings.bitmex.order_type")),
			MaxAmount:  viper.GetFloat64("exchanges_settings.bitmex.max_amount"),
		}
	}

	var strategies StrategiesConfig
	strategiesExist := viper.InConfig("strategies")
	if strategiesExist {
		strategies.EnableBolingerBand = viper.GetBool("strategies.enable_bb")
		strategies.EnableMACD = viper.GetBool("strategies.enable_macd")
		strategies.EnableRSI = viper.GetBool("strategies.enable_rsi")
		strategies.BinSizes = viper.GetStringSlice("strategies.bin_sizes")
		strategies.RetryProcessCount = viper.GetInt("strategies.retry_process_count")
		strategies.GetCandlesCount = viper.GetInt("strategies.get_candles_count")
		strategies.BBLastCandlesCount = viper.GetInt("strategies.bb_last_candles_count")
		strategies.MacdFastCount = viper.GetInt("strategies.macd_fast_count")
		strategies.MacdSlowCount = viper.GetInt("strategies.macd_slow_count")
		strategies.MacdSigCount = viper.GetInt("strategies.macd_sig_count")
		strategies.RsiCount = viper.GetInt("strategies.rsi_count")
		strategies.RsiMinBorder = viper.GetUint32("strategies.rsi_min_border")
		strategies.RsiMaxBorder = viper.GetUint32("strategies.rsi_max_border")
		for _, binSize := range strategies.BinSizes {
			switch binSize {
			case "5m", "15m", "1h", "1d":
				break
			default:
				logrus.Fatalf("failed - unknown bin size: %s", binSize)
			}
		}
		if strategies.MacdFastCount >= strategies.MacdSlowCount {
			logrus.Fatal("macd fast should be less than macd slow")
		}
		if strategies.MacdSigCount > strategies.MacdFastCount ||
			strategies.MacdSigCount > strategies.MacdSlowCount {
			logrus.Fatal("macd sigshould be less than macd slow and macd fast")
		}
		if strategies.RsiCount >= strategies.MacdSlowCount {
			logrus.Fatal("for initialization signals rsi_count must be less than macd_slow_count")
		}
	} else {
		// default
		strategies.EnableBolingerBand = true
		strategies.BinSizes = []string{"5m"}
		strategies.RetryProcessCount = 5
		strategies.GetCandlesCount = 20
		strategies.BBLastCandlesCount = 4
		strategies.RsiCount = 14
		strategies.RsiMinBorder = 30
		strategies.RsiMaxBorder = 70
		strategies.MacdFastCount = 12
		strategies.MacdSlowCount = 26
		strategies.MacdSigCount = 9
	}

	if !strategies.EnableBolingerBand && !strategies.EnableMACD && !strategies.EnableRSI {
		logrus.Fatal("all strategies disabled, enable any strategy")
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
