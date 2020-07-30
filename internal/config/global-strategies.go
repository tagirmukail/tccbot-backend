package config

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type StrategiesGlobConfig struct {
	M1 *StrategiesConfig
	M5 *StrategiesConfig
	H1 *StrategiesConfig
	D1 *StrategiesConfig
}

func initGlobStrategies() StrategiesGlobConfig {
	globalStrategies := StrategiesGlobConfig{}
	globStrateg := viper.GetStringMap("strategies_g")
	if len(globStrateg) == 0 {
		// default
		strategies := &StrategiesConfig{}
		strategies.EnableRSIBB = true
		strategies.RetryProcessCount = 5
		strategies.GetCandlesCount = 20
		strategies.CandlesFilterEnable = true
		strategies.TrendFilterEnable = false
		strategies.MaxFilterTrendCount = 4
		strategies.MaxCandlesFilterCount = 4
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
			k = strings.ToLower(k)
			var strategies = StrategiesConfig{
				EnableMACD:        viper.GetBool(sprintFstrategy("strategies_g.%s.enable_macd", k)),
				EnableRSIBB:       viper.GetBool(sprintFstrategy("strategies_g.%s.enable_rsi_bb", k)),
				RetryProcessCount: viper.GetInt(sprintFstrategy("strategies_g.%s.retry_process_count", k)),
				GetCandlesCount:   viper.GetInt(sprintFstrategy("strategies_g.%s.get_candles_count", k)),

				CandlesFilterEnable:   viper.GetBool(sprintFstrategy("strategies_g.%s.candles_filter_enable", k)),
				TrendFilterEnable:     viper.GetBool(sprintFstrategy("strategies_g.%s.trend_filter_enable", k)),
				MaxFilterTrendCount:   viper.GetInt(sprintFstrategy("strategies_g.%s.max_filter_trend_count", k)),
				MaxCandlesFilterCount: viper.GetInt(sprintFstrategy("strategies_g.%s.max_candles_filter_count", k)),

				BBLastCandlesCount: viper.GetInt(sprintFstrategy("strategies_g.%s.bb_last_candles_count", k)),
				MacdFastCount:      viper.GetInt(sprintFstrategy("strategies_g.%s.macd_fast_count", k)),
				MacdSlowCount:      viper.GetInt(sprintFstrategy("strategies_g.%s.macd_slow_count", k)),
				MacdSigCount:       viper.GetInt(sprintFstrategy("strategies_g.%s.macd_sig_count", k)),
				RsiCount:           viper.GetInt(sprintFstrategy("strategies_g.%s.rsi_count", k)),
				RsiMinBorder:       viper.GetUint32(sprintFstrategy("strategies_g.%s.rsi_min_border", k)),
				RsiMaxBorder:       viper.GetUint32(sprintFstrategy("strategies_g.%s.rsi_max_border", k)),
				RsiTradeCoef:       viper.GetFloat64(sprintFstrategy("strategies_g.%s.rsi_trade_coef", k)),
			}

			switch k {
			case "1m":
				globalStrategies.M1 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("1m strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			case "5m":
				globalStrategies.M5 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("5m strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			case "1h":
				globalStrategies.H1 = &strategies
				fmt.Println("--------------------------------------------")
				fmt.Printf("1h strategies cfg: %#v\n", strategies)
				fmt.Println("--------------------------------------------")
			default:
				logrus.Fatal("unknown global strategies bin size key, must be only: 1m,5m, 1h, 1d")
			}
		}
	}

	return globalStrategies
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

func sprintFstrategy(config, name string) string {
	return fmt.Sprintf(config, name)
}
