package config

type StrategiesConfig struct {
	EnableMACD          bool
	EnableRSIBB         bool
	TrendFilterEnable   bool
	CandlesFilterEnable bool
	RetryProcessCount   int
	GetCandlesCount     int

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
