package config

import (
	"errors"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/types"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

type GlobalConfig struct {
	ExchangesSettings ExchangesSettings
	Admin             Admin
	Scheduler         Scheduler
	Accesses          ExchangesAccess
	GlobStrategies    StrategiesGlobConfig
	OrdProcPeriodSec  int
	DBPath            string
	UpdatedAt         time.Time
}

type Admin struct {
	Username    string `json:"username"`
	SecretToken string `json:"secret_token"`
	Token       string `json:"token"`
}

//nolint:funlen,deadcode,unused
func toGlobalConfig(cfgM models.GlobalConfig) GlobalConfig {
	var cfg GlobalConfig

	cfg.DBPath = cfgM.DBPath
	cfg.OrdProcPeriodSec = cfgM.OrderProcPeriodInSec
	cfg.Admin.Token = cfgM.Admin.Token
	cfg.Admin.SecretToken = cfgM.Admin.SecretToken
	cfg.Admin.Username = cfgM.Admin.Username
	if cfgM.ExchangeAccess.Exchange == string(types.Bitmex) {
		if cfgM.ExchangeAccess.Test {
			cfg.Accesses.Bitmex.Testnet.Secret = cfgM.ExchangeAccess.Secret
			cfg.Accesses.Bitmex.Testnet.Key = cfgM.ExchangeAccess.Key
		} else {
			cfg.Accesses.Bitmex.Secret = cfgM.ExchangeAccess.Secret
			cfg.Accesses.Bitmex.Key = cfgM.ExchangeAccess.Key
		}
	}

	if cfgM.ExchangeAPISettings.Exchange == string(types.Bitmex) {
		cfg.ExchangesSettings.Bitmex.Test = cfgM.ExchangeAPISettings.Test
		cfg.ExchangesSettings.Bitmex.PingSec = cfgM.ExchangeAPISettings.PingSec
		cfg.ExchangesSettings.Bitmex.TimeoutSec = cfgM.ExchangeAPISettings.TimeoutSec
		cfg.ExchangesSettings.Bitmex.RetrySec = cfgM.ExchangeAPISettings.RetrySec
		cfg.ExchangesSettings.Bitmex.BufferSize = cfgM.ExchangeAPISettings.BufferSize
		cfg.ExchangesSettings.Bitmex.Currency = cfgM.ExchangeAPISettings.Currency
		cfg.ExchangesSettings.Bitmex.Symbol = cfgM.ExchangeAPISettings.Symbol
		cfg.ExchangesSettings.Bitmex.OrderType = cfgM.ExchangeAPISettings.OrderType
		cfg.ExchangesSettings.Bitmex.MaxAmount = cfgM.ExchangeAPISettings.MaxAmount
		cfg.ExchangesSettings.Bitmex.LimitContractsCount = cfgM.ExchangeAPISettings.LimitContractsCount
		cfg.ExchangesSettings.Bitmex.SellOrderCoef = cfgM.ExchangeAPISettings.SellOrderCoef
		cfg.ExchangesSettings.Bitmex.BuyOrderCoef = cfgM.ExchangeAPISettings.BuyOrderCoef
	}

	if cfgM.Scheduler.Type == "position" {
		cfg.Scheduler.Position.Enable = cfgM.Scheduler.Enable
		cfg.Scheduler.Position.LossPnlDiff = cfgM.Scheduler.LossPnlDiff
		cfg.Scheduler.Position.ProfitPnlDiff = cfgM.Scheduler.ProfitPnlDiff
		cfg.Scheduler.Position.LossCloseBTC = cfgM.Scheduler.LossCloseBTC
		cfg.Scheduler.Position.ProfitCloseBTC = cfgM.Scheduler.ProfitCloseBTC
		cfg.Scheduler.Position.PriceTrailing = cfgM.Scheduler.PriceTrailing
	}

	var strategiesCfg StrategiesConfig
	strategiesCfg.EnableRSIBB = cfgM.StrategiesConfig.EnableRSIBB
	strategiesCfg.RetryProcessCount = cfgM.StrategiesConfig.RetryProcessCount
	strategiesCfg.GetCandlesCount = cfgM.StrategiesConfig.GetCandlesCount
	strategiesCfg.TrendFilterEnable = cfgM.StrategiesConfig.TrendFilterEnable
	strategiesCfg.CandlesFilterEnable = cfgM.StrategiesConfig.CandlesFilterEnable
	strategiesCfg.MaxFilterTrendCount = cfgM.StrategiesConfig.MaxFilterTrendCount
	strategiesCfg.MaxCandlesFilterCount = cfgM.StrategiesConfig.MaxCandlesFilterCount
	strategiesCfg.BBLastCandlesCount = cfgM.StrategiesConfig.BBLastCandlesCount
	strategiesCfg.RsiCount = cfgM.StrategiesConfig.RsiCount
	strategiesCfg.RsiMinBorder = cfgM.StrategiesConfig.RsiMinBorder
	strategiesCfg.RsiMaxBorder = cfgM.StrategiesConfig.RsiMaxBorder
	strategiesCfg.RsiTradeCoef = cfgM.StrategiesConfig.RsiTradeCoef
	strategiesCfg.MacdFastCount = cfgM.StrategiesConfig.MacdFastCount
	strategiesCfg.MacdSlowCount = cfgM.StrategiesConfig.MacdSlowCount
	strategiesCfg.MacdSigCount = cfgM.StrategiesConfig.MacdSigCount
	switch cfgM.StrategiesConfig.Bin {
	case models.Bin1m:
		cfg.GlobStrategies.M1 = &strategiesCfg
	case models.Bin5m:
		cfg.GlobStrategies.M5 = &strategiesCfg
	case models.Bin1h:
		cfg.GlobStrategies.H1 = &strategiesCfg
	case models.Bin1d:
		cfg.GlobStrategies.D1 = &strategiesCfg
	}

	return cfg
}

//nolint:funlen
func toModelConfiguration(cfg GlobalConfig) (models.GlobalConfig, error) {
	var cfgM models.GlobalConfig

	cfgM.DBPath = cfg.DBPath
	cfgM.OrderProcPeriodInSec = cfg.OrdProcPeriodSec
	cfgM.Admin.Token = cfg.Admin.Token
	cfgM.Admin.SecretToken = cfg.Admin.SecretToken
	cfgM.Admin.Username = cfg.Admin.Username
	cfgM.ExchangeAPISettings.Exchange = string(types.Bitmex)
	cfgM.ExchangeAPISettings.Enable = true
	cfgM.ExchangeAPISettings.Test = cfg.ExchangesSettings.Bitmex.Test
	cfgM.ExchangeAPISettings.BuyOrderCoef = cfg.ExchangesSettings.Bitmex.BuyOrderCoef
	cfgM.ExchangeAPISettings.SellOrderCoef = cfg.ExchangesSettings.Bitmex.SellOrderCoef
	cfgM.ExchangeAPISettings.PingSec = cfg.ExchangesSettings.Bitmex.PingSec
	cfgM.ExchangeAPISettings.TimeoutSec = cfg.ExchangesSettings.Bitmex.TimeoutSec
	cfgM.ExchangeAPISettings.RetrySec = cfg.ExchangesSettings.Bitmex.RetrySec
	cfgM.ExchangeAPISettings.BufferSize = cfg.ExchangesSettings.Bitmex.BufferSize
	cfgM.ExchangeAPISettings.Currency = cfg.ExchangesSettings.Bitmex.Currency
	cfgM.ExchangeAPISettings.Symbol = cfg.ExchangesSettings.Bitmex.Symbol
	cfgM.ExchangeAPISettings.OrderType = cfg.ExchangesSettings.Bitmex.OrderType
	cfgM.ExchangeAPISettings.MaxAmount = cfg.ExchangesSettings.Bitmex.MaxAmount
	cfgM.ExchangeAPISettings.LimitContractsCount = cfg.ExchangesSettings.Bitmex.LimitContractsCount
	cfgM.ExchangeAccess.Exchange = string(types.Bitmex)
	if cfg.Accesses.Bitmex.Secret != "" && cfg.Accesses.Bitmex.Key != "" {
		cfgM.ExchangeAccess.Key = cfg.Accesses.Bitmex.Key
		cfgM.ExchangeAccess.Secret = cfg.Accesses.Bitmex.Secret
	} else {
		cfgM.ExchangeAccess.Key = cfg.Accesses.Bitmex.Testnet.Key
		cfgM.ExchangeAccess.Secret = cfg.Accesses.Bitmex.Testnet.Secret
		cfgM.ExchangeAccess.Test = true
	}
	cfgM.Scheduler.Type = "position"
	cfgM.Scheduler.Enable = cfg.Scheduler.Position.Enable
	cfgM.Scheduler.PriceTrailing = cfg.Scheduler.Position.PriceTrailing
	cfgM.Scheduler.ProfitCloseBTC = cfg.Scheduler.Position.ProfitCloseBTC
	cfgM.Scheduler.LossCloseBTC = cfg.Scheduler.Position.LossCloseBTC
	cfgM.Scheduler.ProfitPnlDiff = cfg.Scheduler.Position.ProfitPnlDiff
	cfgM.Scheduler.LossPnlDiff = cfg.Scheduler.Position.LossPnlDiff

	binSizes := cfg.GlobStrategies.GetBinSizes()
	if len(binSizes) == 0 || len(binSizes) > 1 {
		return cfgM, errors.New("wrong strategies, must be only 1 strategy")
	}

	strategy := cfg.GlobStrategies.GetCfgByBinSize(binSizes[0])

	bin, _ := models.ToBinSize(binSizes[0])

	cfgM.StrategiesConfig.Bin = bin
	cfgM.StrategiesConfig.EnableRSIBB = strategy.EnableRSIBB
	cfgM.StrategiesConfig.RetryProcessCount = strategy.RetryProcessCount
	cfgM.StrategiesConfig.GetCandlesCount = strategy.GetCandlesCount
	cfgM.StrategiesConfig.TrendFilterEnable = strategy.TrendFilterEnable
	cfgM.StrategiesConfig.CandlesFilterEnable = strategy.CandlesFilterEnable
	cfgM.StrategiesConfig.MaxFilterTrendCount = strategy.MaxFilterTrendCount
	cfgM.StrategiesConfig.MaxCandlesFilterCount = strategy.MaxCandlesFilterCount
	cfgM.StrategiesConfig.BBLastCandlesCount = strategy.BBLastCandlesCount
	cfgM.StrategiesConfig.RsiCount = strategy.RsiCount
	cfgM.StrategiesConfig.RsiMinBorder = strategy.RsiMinBorder
	cfgM.StrategiesConfig.RsiMaxBorder = strategy.RsiMaxBorder
	cfgM.StrategiesConfig.RsiTradeCoef = strategy.RsiTradeCoef
	cfgM.StrategiesConfig.MacdFastCount = strategy.MacdFastCount
	cfgM.StrategiesConfig.MacdSlowCount = strategy.MacdSlowCount
	cfgM.StrategiesConfig.MacdSigCount = strategy.MacdSigCount

	return cfgM, nil
}
