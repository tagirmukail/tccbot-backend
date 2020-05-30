package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type GlobalConfig struct {
	ExchangesSettings ExchangesSettings
	Admin             Admin
	DB                DB
	Accesses          ExchangesAccess
	GlobStrategies    StrategiesGlobConfig
	OrdProcPeriodSec  int
	DBPath            string
}

type DB struct {
	DBName   string
	User     string
	Password string
	Host     string
	Port     uint32
	SSLMode  string
}

type StrategiesGlobConfig struct {
	M1 *StrategiesConfig
	M5 *StrategiesConfig
	H1 *StrategiesConfig
	D1 *StrategiesConfig
}

func (s *StrategiesGlobConfig) GetCfgByBinSize(binSize string) *StrategiesConfig {
	switch binSize {
	case "1m":
		return s.M1
	case "5m":
		return s.M5
	case "1h":
		return s.H1
	case "1d":
		return s.D1
	default:
		return nil
	}
}

func (s *StrategiesGlobConfig) GetThemes() []types.Theme {
	var result []types.Theme
	if s.M1 != nil {
		result = append(result, types.TradeBin1m)
	}
	if s.M5 != nil {
		result = append(result, types.TradeBin5m)
	}
	if s.H1 != nil {
		result = append(result, types.TradeBin1h)
	}
	if s.D1 != nil {
		result = append(result, types.TradeBin1d)
	}
	return result
}

func (s *StrategiesGlobConfig) GetBinSizes() []string {
	var result []string
	if s.M1 != nil {
		result = append(result, "1m")
	}
	if s.M5 != nil {
		result = append(result, "5m")
	}
	if s.H1 != nil {
		result = append(result, "1h")
	}
	if s.D1 != nil {
		result = append(result, "1d")
	}
	return result
}

type StrategiesConfig struct {
	EnableMACD         bool
	EnableRSIBB        bool
	RetryProcessCount  int
	GetCandlesCount    int
	BBLastCandlesCount int
	RsiCount           int
	RsiMinBorder       uint32
	RsiMaxBorder       uint32
	RsiTradeCoef       float64
	MacdFastCount      int
	MacdSlowCount      int
	MacdSigCount       int
}

func (strategies *StrategiesConfig) AnyStrategyEnabled() bool {
	return strategies.EnableRSIBB
}

type ExchangesSettings struct {
	Bitmex ApiSettings
}

type ExchangeSettings struct {
	Enable bool
	Api    ApiSettings
}

type ApiSettings struct {
	Test                bool
	PingSec             int
	TimeoutSec          int
	RetrySec            int
	BufferSize          int
	Currency            string
	Symbol              string
	OrderType           types.OrderType
	MaxAmount           float64
	ClosePositionMinBTC float64
	LimitContractsCount int
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
	Key     string `json:"key"`
	Secret  string `json:"secret"`
	Testnet struct {
		Key    string
		Secret string
	}
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
			Test:                true,
			PingSec:             20,
			TimeoutSec:          30,
			RetrySec:            5,
			BufferSize:          10,
			Currency:            "XBt",
			Symbol:              "XBTUSD",
			OrderType:           types.Limit,
			MaxAmount:           130,
			ClosePositionMinBTC: 0.0005,
			LimitContractsCount: 300,
		}
	} else {
		bitmex = ApiSettings{
			Test:                viper.GetBool("exchanges_settings.bitmex.test"),
			PingSec:             viper.GetInt("exchanges_settings.bitmex.ping_sec"),
			TimeoutSec:          viper.GetInt("exchanges_settings.bitmex.timeout_sec"),
			RetrySec:            viper.GetInt("exchanges_settings.bitmex.retry_sec"),
			BufferSize:          viper.GetInt("exchanges_settings.bitmex.buffer_size"),
			Symbol:              viper.GetString("exchanges_settings.bitmex.symbol"),
			Currency:            viper.GetString("exchanges_settings.bitmex.currency"),
			OrderType:           types.OrderType(viper.GetString("exchanges_settings.bitmex.order_type")),
			MaxAmount:           viper.GetFloat64("exchanges_settings.bitmex.max_amount"),
			ClosePositionMinBTC: viper.GetFloat64("exchanges_settings.bitmex.close_position_min_btc"),
			LimitContractsCount: viper.GetInt("exchanges_settings.bitmex.limit_contracts_cnt"),
		}
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("bitmex settings: %#v\n", bitmex)
	fmt.Println("--------------------------------------------")

	globalStrategies := StrategiesGlobConfig{}
	globStrateg := viper.GetStringMap("strategies_g")
	if len(globStrateg) == 0 {
		// default
		strategies := &StrategiesConfig{}
		strategies.EnableRSIBB = true
		strategies.RetryProcessCount = 5
		strategies.GetCandlesCount = 20
		strategies.BBLastCandlesCount = 4
		strategies.RsiCount = 14
		strategies.RsiMinBorder = 30
		strategies.RsiMaxBorder = 70
		strategies.RsiTradeCoef = 0.0004
		strategies.MacdFastCount = 12
		strategies.MacdSlowCount = 26
		strategies.MacdSigCount = 9
		globalStrategies.M1 = strategies
		fmt.Println("--------------------------------------------")
		fmt.Printf("1m strategies cfg: %#v\n", strategies)
		fmt.Println("--------------------------------------------")
	} else {
		for k := range globStrateg {
			switch k {
			case "1m":
				var strategies = StrategiesConfig{
					EnableMACD:         viper.GetBool("strategies_g.1m.enable_macd"),
					EnableRSIBB:        viper.GetBool("strategies_g.1m.enable_rsi_bb"),
					RetryProcessCount:  viper.GetInt("strategies_g.1m.retry_process_count"),
					GetCandlesCount:    viper.GetInt("strategies_g.1m.get_candles_count"),
					BBLastCandlesCount: viper.GetInt("strategies_g.1m.bb_last_candles_count"),
					MacdFastCount:      viper.GetInt("strategies_g.1m.macd_fast_count"),
					MacdSlowCount:      viper.GetInt("strategies_g.1m.macd_slow_count"),
					MacdSigCount:       viper.GetInt("strategies_g.1m.macd_sig_count"),
					RsiCount:           viper.GetInt("strategies_g.1m.rsi_count"),
					RsiMinBorder:       viper.GetUint32("strategies_g.1m.rsi_min_border"),
					RsiMaxBorder:       viper.GetUint32("strategies_g.1m.rsi_max_border"),
					RsiTradeCoef:       viper.GetFloat64("strategies_g.1m.rsi_trade_coef"),
				}
				globalStrategies.M1 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("1m strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			case "5m":
				var strategies = StrategiesConfig{
					EnableMACD:         viper.GetBool("strategies_g.5m.enable_macd"),
					EnableRSIBB:        viper.GetBool("strategies_g.5m.enable_rsi_bb"),
					RetryProcessCount:  viper.GetInt("strategies_g.5m.retry_process_count"),
					GetCandlesCount:    viper.GetInt("strategies_g.5m.get_candles_count"),
					BBLastCandlesCount: viper.GetInt("strategies_g.5m.bb_last_candles_count"),
					MacdFastCount:      viper.GetInt("strategies_g.5m.macd_fast_count"),
					MacdSlowCount:      viper.GetInt("strategies_g.5m.macd_slow_count"),
					MacdSigCount:       viper.GetInt("strategies_g.5m.macd_sig_count"),
					RsiCount:           viper.GetInt("strategies_g.5m.rsi_count"),
					RsiMinBorder:       viper.GetUint32("strategies_g.5m.rsi_min_border"),
					RsiMaxBorder:       viper.GetUint32("strategies_g.5m.rsi_max_border"),
					RsiTradeCoef:       viper.GetFloat64("strategies_g.5m.rsi_trade_coef"),
				}
				globalStrategies.M5 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("5m strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			case "1h":
				var strategies = StrategiesConfig{
					EnableMACD:         viper.GetBool("strategies_g.1h.enable_macd"),
					EnableRSIBB:        viper.GetBool("strategies_g.1h.enable_rsi_bb"),
					RetryProcessCount:  viper.GetInt("strategies_g.1h.retry_process_count"),
					GetCandlesCount:    viper.GetInt("strategies_g.1h.get_candles_count"),
					BBLastCandlesCount: viper.GetInt("strategies_g.1h.bb_last_candles_count"),
					MacdFastCount:      viper.GetInt("strategies_g.1h.macd_fast_count"),
					MacdSlowCount:      viper.GetInt("strategies_g.1h.macd_slow_count"),
					MacdSigCount:       viper.GetInt("strategies_g.1h.macd_sig_count"),
					RsiCount:           viper.GetInt("strategies_g.1h.rsi_count"),
					RsiMinBorder:       viper.GetUint32("strategies_g.1h.rsi_min_border"),
					RsiMaxBorder:       viper.GetUint32("strategies_g.1h.rsi_max_border"),
					RsiTradeCoef:       viper.GetFloat64("strategies_g.1h.rsi_trade_coef"),
				}
				globalStrategies.H1 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("1h strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			default:
				logrus.Fatal("unknown global strategies bin size key, must be only: 1m,5m, 1h, 1d")
			}
		}
	}

	cfg := &GlobalConfig{
		GlobStrategies: globalStrategies,
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
				Testnet: struct {
					Key    string
					Secret string
				}{
					Key:    viper.GetString("exchanges_access.bitmex.testnet.key"),
					Secret: viper.GetString("exchanges_access.bitmex.testnet.secret"),
				},
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
		DBPath:           viper.GetString("db_path"),
		OrdProcPeriodSec: viper.GetInt("ord_proc_period_sec"),
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("global cfg: %#v\n", cfg)
	fmt.Println("--------------------------------------------")
	return cfg, nil
}
