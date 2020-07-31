package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type BitmexData struct {
	Table  string               `json:"table"`
	Action string               `json:"action"`
	Data   []BitmexIncomingData `json:"data"`
}

type TradeBinData struct {
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Trades   int     `json:"trades"`
	Volume   int64   `json:"volume"`
	LastSize int     `json:"lastSize"`
	Turnover int64   `json:"turnover"`
	Vwap     float64 `json:"vwap"`
}

type BitmexExchangeData struct {
	Side          types.Side `json:"side"`
	Size          int        `json:"size"`
	Price         float64    `json:"price"`
	TickDirection string     `json:"tickDirection"`
	TrdMatchID    string     `json:"trdMatchID"`
	GrossValue    int64      `json:"grossValue"`
}

type PositionData struct {
	Account              int64     `json:"account"`
	AvgCostPrice         float64   `json:"avgCostPrice"`
	AvgEntryPrice        float64   `json:"avgEntryPrice"`
	BankruptPrice        float64   `json:"bankruptPrice"`
	BreakEvenPrice       float64   `json:"breakEvenPrice"`
	Commission           float64   `json:"commission"`
	CrossMargin          bool      `json:"crossMargin"`
	Currency             string    `json:"currency"`
	CurrentComm          int64     `json:"currentComm"`
	CurrentCost          int64     `json:"currentCost"`
	CurrentQty           int64     `json:"currentQty"`
	CurrentTimestamp     time.Time `json:"currentTimestamp"`
	DeleveragePercentile float64   `json:"deleveragePercentile"`
	ExecBuyCost          int64     `json:"execBuyCost"`
	ExecBuyQty           int64     `json:"execBuyQty"`
	ExecComm             int64     `json:"execComm"`
	ExecCost             int64     `json:"execCost"`
	ExecQty              int64     `json:"execQty"`
	ExecSellCost         int64     `json:"execSellCost"`
	ExecSellQty          int64     `json:"execSellQty"`
	ForeignNotional      float64   `json:"foreignNotional"`
	GrossExecCost        int64     `json:"grossExecCost"`
	GrossOpenCost        int64     `json:"grossOpenCost"`
	GrossOpenPremium     int64     `json:"grossOpenPremium"`
	HomeNotional         float64   `json:"homeNotional"`
	IndicativeTax        int64     `json:"indicativeTax"`
	IndicativeTaxRate    float64   `json:"indicativeTaxRate"`
	InitMargin           int64     `json:"initMargin"`
	InitMarginReq        float64   `json:"initMarginReq"`
	IsOpen               bool      `json:"isOpen"`
	LastPrice            float64   `json:"lastPrice"`
	LastValue            int64     `json:"lastValue"`
	Leverage             float64   `json:"leverage"`
	LiquidationPrice     float64   `json:"liquidationPrice"`
	LongBankrupt         int64     `json:"longBankrupt"`
	MaintMargin          int64     `json:"maintMargin"`
	MaintMarginReq       float64   `json:"maintMarginReq"`
	MarginCallPrice      float64   `json:"marginCallPrice"`
	MarkPrice            float64   `json:"markPrice"`
	MarkValue            int64     `json:"markValue"`
	OpenOrderBuyCost     int64     `json:"openOrderBuyCost"`
	OpenOrderBuyPremium  int64     `json:"openOrderBuyPremium"`
	OpenOrderBuyQty      int64     `json:"openOrderBuyQty"`
	OpenOrderSellCost    int64     `json:"openOrderSellCost"`
	OpenOrderSellPremium int64     `json:"openOrderSellPremium"`
	OpenOrderSellQty     int64     `json:"openOrderSellQty"`
	OpeningComm          int64     `json:"openingComm"`
	OpeningCost          int64     `json:"openingCost"`
	OpeningQty           int64     `json:"openingQty"`
	OpeningTimestamp     time.Time `json:"openingTimestamp"`
	PosAllowance         int64     `json:"posAllowance"`
	PosComm              int64     `json:"posComm"`
	PosCost              int64     `json:"posCost"`
	PosCost2             int64     `json:"posCost2"`
	PosCross             int64     `json:"posCross"`
	PosInit              int64     `json:"posInit"`
	PosLoss              int64     `json:"posLoss"`
	PosMaint             int64     `json:"posMaint"`
	PosMargin            int64     `json:"posMargin"`
	PosState             string    `json:"posState"`
	PrevClosePrice       float64   `json:"prevClosePrice"`
	PrevRealisedPnl      int64     `json:"prevRealisedPnl"`
	PrevUnrealisedPnl    int64     `json:"prevUnrealisedPnl"`
	QuoteCurrency        string    `json:"quoteCurrency"`
	RealisedCost         int64     `json:"realisedCost"`
	RealisedGrossPnl     int64     `json:"realisedGrossPnl"`
	RealisedPnl          int64     `json:"realisedPnl"`
	RealisedTax          int64     `json:"realisedTax"`
	RebalancedPnl        int64     `json:"rebalancedPnl"`
	RiskLimit            int64     `json:"riskLimit"`
	RiskValue            int64     `json:"riskValue"`
	SessionMargin        int64     `json:"sessionMargin"`
	ShortBankrupt        int64     `json:"shortBankrupt"`
	SimpleCost           float64   `json:"simpleCost"`
	SimplePnl            float64   `json:"simplePnl"`
	SimplePnlPcnt        float64   `json:"simplePnlPcnt"`
	SimpleQty            float64   `json:"simpleQty"`
	SimpleValue          float64   `json:"simpleValue"`
	TargetExcessMargin   int64     `json:"targetExcessMargin"`
	TaxBase              int64     `json:"taxBase"`
	TaxableMargin        int64     `json:"taxableMargin"`
	//Timestamp            time.Time `json:"timestamp"`
	Underlying         string  `json:"underlying"`
	UnrealisedCost     int64   `json:"unrealisedCost"`
	UnrealisedGrossPnl int64   `json:"unrealisedGrossPnl"`
	UnrealisedPnl      int64   `json:"unrealisedPnl"`
	UnrealisedPnlPcnt  float64 `json:"unrealisedPnlPcnt"`
	UnrealisedRoePcnt  float64 `json:"unrealisedRoePcnt"`
	UnrealisedTax      int64   `json:"unrealisedTax"`
	VarMargin          int64   `json:"varMargin"`
}

type BitmexIncomingData struct {
	Table           string       `json:"table"`
	Symbol          types.Symbol `json:"symbol"`
	Timestamp       string       `json:"timestamp"`
	HomeNotional    float64      `json:"homeNotional"`
	ForeignNotional float64      `json:"foreignNotional"`

	TradeBinData
	BitmexExchangeData
	PositionData
}

func (b *BitmexData) Validate() error {
	//if b.Table != string(types.TradeBin1m) &&
	//	b.Table != string(types.TradeBin5m) &&
	//	b.Table != string(types.TradeBin1h) &&
	//	b.Table != string(types.TradeBin1d) {
	//	return fmt.Errorf("bad table:%v", b.Table)
	//}
	//
	if b.Action != "update" && b.Action != "insert" {
		return fmt.Errorf("bad action: %v", b.Action)
	}

	if len(b.Data) == 0 {
		return errors.New("empty data")
	}

	return nil
}
