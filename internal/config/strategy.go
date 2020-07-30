package config

type StrategiesConfig struct {
	EnableMACD        bool
	EnableRSIBB       bool
	RetryProcessCount int
	GetCandlesCount   int

	TrendFilterEnable     bool
	CandlesFilterEnable   bool
	MaxFilterTrendCount   int
	MaxCandlesFilterCount int

	BBLastCandlesCount int

	RsiCount     int
	RsiMinBorder uint32
	RsiMaxBorder uint32
	RsiTradeCoef float64

	MacdFastCount int
	MacdSlowCount int
	MacdSigCount  int
}

func (strategies *StrategiesConfig) AnyStrategyEnabled() bool {
	return strategies.EnableRSIBB
}
