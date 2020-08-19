package models

import (
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type GlobalConfig struct {
	ID                   int                 `db:"id" json:"id"`
	DBPath               string              `db:"db_path" json:"db_path"`
	OrderProcPeriodInSec int                 `db:"order_proc_period_in_sec" json:"order_proc_period_in_sec"`
	CreatedAt            time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at" db:"updated_at"`
	AdminID              int                 `json:"admin_id" db:"admin_id"`
	Admin                Admin               `json:"admin"`
	Scheduler            Scheduler           `json:"scheduler"`
	ExchangeAPISettings  ExchangeAPISettings `json:"api_settings"`
	ExchangeAccess       ExchangeAccess      `json:"exchange_access"`
	StrategiesConfig     StrategiesConfig    `json:"strategies_config"`
}

type Admin struct {
	ID          int    `db:"id" json:"id"`
	Exchange    string `json:"exchange" db:"exchange"`
	Username    string `json:"username" db:"username"`
	SecretToken string `json:"secret_token" db:"secret_token"`
	Token       string `json:"token" db:"token"`
	GlobalID    int    `json:"global_id" db:"global_id"`
}

type Scheduler struct {
	ID             int     `db:"id" json:"id"`
	Type           string  `json:"type" db:"type"`
	Enable         bool    `json:"enable" db:"enable"`
	PriceTrailing  float64 `json:"price_trailing" db:"price_trailing"`
	ProfitCloseBTC float64 `json:"profit_close_btc" db:"profit_close_btc"`
	LossCloseBTC   float64 `json:"loss_close_btc" db:"loss_close_btc"`
	ProfitPnlDiff  float64 `json:"profit_pnl_diff" db:"profit_pnl_diff"`
	LossPnlDiff    float64 `json:"loss_pnl_diff" db:"loss_pnl_diff"`
	GlobalID       int     `json:"global_id" db:"global_id"`
}

type ExchangeAccess struct {
	ID       int    `db:"id" json:"id"`
	Exchange string `json:"exchange" db:"exchange"`
	Test     bool   `json:"test" db:"test"`
	Key      string `json:"key" db:"key"`
	Secret   string `json:"secret" db:"secret"`
	GlobalID int    `json:"global_id" db:"global_id"`
}

type ExchangeAPISettings struct {
	ID                  int             `db:"id" json:"id"`
	Exchange            string          `json:"exchange" db:"exchange"`
	Enable              bool            `json:"enable" db:"enable"`
	Test                bool            `json:"test" db:"test"`
	PingSec             int             `json:"ping_sec" db:"ping_sec"`
	TimeoutSec          int             `json:"timeout_sec" db:"timeout_sec"`
	RetrySec            int             `json:"retry_sec" db:"retry_sec"`
	BufferSize          int             `json:"buffer_size" db:"buffer_size"`
	Currency            string          `json:"currency" db:"currency"`
	Symbol              string          `json:"symbol" db:"symbol"`
	OrderType           types.OrderType `json:"order_type" db:"order_type"`
	MaxAmount           float64         `json:"max_amount" db:"max_amount"`
	LimitContractsCount int             `json:"limit_contracts_cnt" db:"limit_contracts_cnt"`
	SellOrderCoef       float64         `json:"sell_order_coef" db:"sell_order_coef"`
	BuyOrderCoef        float64         `json:"buy_order_coef" db:"buy_order_coef"`
	GlobalID            int             `json:"global_id" db:"global_id"`
}

type StrategiesConfig struct {
	EnableRSIBB         bool    `json:"enable_rsi_bb" db:"enable_rsi_bb"`
	TrendFilterEnable   bool    `json:"trend_filter_enable" db:"trend_filter_enable"`
	CandlesFilterEnable bool    `json:"candles_filter_enable" db:"candles_filter_enable"`
	Bin                 BinSize `json:"bin" db:"bin"`
	RsiMinBorder        uint32  `json:"rsi_min_border" db:"rsi_min_border"`
	RsiMaxBorder        uint32  `json:"rsi_max_border" db:"rsi_max_border"`
	ID                  int     `db:"id" json:"id"`
	RetryProcessCount   int     `json:"retry_process_count" db:"retry_process_count"`
	GetCandlesCount     int     `json:"get_candles_count" db:"get_candles_count"`

	MaxFilterTrendCount   int `json:"max_filter_trend_count" db:"max_filter_trend_count"`
	MaxCandlesFilterCount int `json:"max_candles_filter_count" db:"max_candles_filter_count"`

	BBLastCandlesCount int `json:"bb_last_candles_count" db:"bb_last_candles_count"`

	RsiCount int `json:"rsi_count" db:"rsi_count"`

	MacdFastCount int `json:"macd_fast_count" db:"macd_fast_count"`
	MacdSlowCount int `json:"macd_slow_count" db:"macd_slow_count"`
	MacdSigCount  int `json:"macd_sig_count" db:"macd_sig_count"`

	GlobalID int `json:"global_id" db:"global_id"`

	RsiTradeCoef float64 `json:"rsi_trade_coef" db:"rsi_trade_coef"`
}
