package bitmex

import "time"

type EndpointLimit int

// UserMargin margin information
type UserMargin struct {
	Account            int64     `json:"account"`
	Action             string    `json:"action"`
	Amount             int64     `json:"amount"`
	AvailableMargin    int64     `json:"availableMargin"`
	Commission         float64   `json:"commission"`
	ConfirmedDebit     int64     `json:"confirmedDebit"`
	Currency           string    `json:"currency"`
	ExcessMargin       int64     `json:"excessMargin"`
	ExcessMarginPcnt   float64   `json:"excessMarginPcnt"`
	GrossComm          int64     `json:"grossComm"`
	GrossExecCost      int64     `json:"grossExecCost"`
	GrossLastValue     int64     `json:"grossLastValue"`
	GrossMarkValue     int64     `json:"grossMarkValue"`
	GrossOpenCost      int64     `json:"grossOpenCost"`
	GrossOpenPremium   int64     `json:"grossOpenPremium"`
	IndicativeTax      int64     `json:"indicativeTax"`
	InitMargin         int64     `json:"initMargin"`
	MaintMargin        int64     `json:"maintMargin"`
	MarginBalance      int64     `json:"marginBalance"`
	MarginBalancePcnt  float64   `json:"marginBalancePcnt"`
	MarginLeverage     float64   `json:"marginLeverage"`
	MarginUsedPcnt     float64   `json:"marginUsedPcnt"`
	PendingCredit      int64     `json:"pendingCredit"`
	PendingDebit       int64     `json:"pendingDebit"`
	PrevRealisedPnl    int64     `json:"prevRealisedPnl"`
	PrevState          string    `json:"prevState"`
	PrevUnrealisedPnl  int64     `json:"prevUnrealisedPnl"`
	RealisedPnl        int64     `json:"realisedPnl"`
	RiskLimit          int64     `json:"riskLimit"`
	RiskValue          int64     `json:"riskValue"`
	SessionMargin      int64     `json:"sessionMargin"`
	State              string    `json:"state"`
	SyntheticMargin    int64     `json:"syntheticMargin"`
	TargetExcessMargin int64     `json:"targetExcessMargin"`
	TaxableMargin      int64     `json:"taxableMargin"`
	Timestamp          time.Time `json:"timestamp"`
	UnrealisedPnl      int64     `json:"unrealisedPnl"`
	UnrealisedProfit   int64     `json:"unrealisedProfit"`
	VarMargin          int64     `json:"varMargin"`
	WalletBalance      int64     `json:"walletBalance"`
	WithdrawableMargin int64     `json:"withdrawableMargin"`
}

// WalletInfo wallet information
type WalletInfo struct {
	Account          int64     `json:"account"`
	Addr             string    `json:"addr"`
	Amount           int64     `json:"amount"`
	ConfirmedDebit   int64     `json:"confirmedDebit"`
	Currency         string    `json:"currency"`
	DeltaAmount      int64     `json:"deltaAmount"`
	DeltaDeposited   int64     `json:"deltaDeposited"`
	DeltaTransferIn  int64     `json:"deltaTransferIn"`
	DeltaTransferOut int64     `json:"deltaTransferOut"`
	DeltaWithdrawn   int64     `json:"deltaWithdrawn"`
	Deposited        int64     `json:"deposited"`
	PendingCredit    int64     `json:"pendingCredit"`
	PendingDebit     int64     `json:"pendingDebit"`
	PrevAmount       int64     `json:"prevAmount"`
	PrevDeposited    int64     `json:"prevDeposited"`
	PrevTimestamp    time.Time `json:"prevTimestamp"`
	PrevTransferIn   int64     `json:"prevTransferIn"`
	PrevTransferOut  int64     `json:"prevTransferOut"`
	PrevWithdrawn    int64     `json:"prevWithdrawn"`
	Script           string    `json:"script"`
	Timestamp        time.Time `json:"timestamp"`
	TransferIn       int64     `json:"transferIn"`
	TransferOut      int64     `json:"transferOut"`
	WithdrawalLock   []string  `json:"withdrawalLock"`
	Withdrawn        int64     `json:"withdrawn"`
}

// Order Placement, Cancellation, Amending, and History
type Order struct {
	Account               int64     `json:"account"`
	AvgPx                 float64   `json:"avgPx"`
	ClOrdID               string    `json:"clOrdID"`
	ClOrdLinkID           string    `json:"clOrdLinkID"`
	ContingencyType       string    `json:"contingencyType"`
	CumQty                int64     `json:"cumQty"`
	Currency              string    `json:"currency"`
	DisplayQuantity       int64     `json:"displayQty"`
	ExDestination         string    `json:"exDestination"`
	ExecInst              string    `json:"execInst"`
	LeavesQty             int64     `json:"leavesQty"`
	MultiLegReportingType string    `json:"multiLegReportingType"`
	OrdRejReason          string    `json:"ordRejReason"`
	OrdStatus             string    `json:"ordStatus"`
	OrdType               int64     `json:"ordType,string"`
	OrderID               string    `json:"orderID"`
	OrderQty              int64     `json:"orderQty"`
	PegOffsetValue        float64   `json:"pegOffsetValue"`
	PegPriceType          string    `json:"pegPriceType"`
	Price                 float64   `json:"price"`
	SettlCurrency         string    `json:"settlCurrency"`
	Side                  int64     `json:"side,string"`
	SimpleCumQty          float64   `json:"simpleCumQty"`
	SimpleLeavesQty       float64   `json:"simpleLeavesQty"`
	SimpleOrderQty        float64   `json:"simpleOrderQty"`
	StopPx                float64   `json:"stopPx"`
	Symbol                string    `json:"symbol"`
	Text                  string    `json:"text"`
	TimeInForce           string    `json:"timeInForce"`
	Timestamp             time.Time `json:"timestamp"`
	TransactTime          string    `json:"transactTime"`
	Triggered             string    `json:"triggered"`
	WorkingIndicator      bool      `json:"workingIndicator"`
}

// OrderCopied Placement, Cancellation, Amending, and History
type OrderCopied struct {
	Account               int64     `json:"account"`
	AvgPx                 float64   `json:"avgPx"`
	ClOrdID               string    `json:"clOrdID"`
	ClOrdLinkID           string    `json:"clOrdLinkID"`
	ContingencyType       string    `json:"contingencyType"`
	CumQty                int64     `json:"cumQty"`
	Currency              string    `json:"currency"`
	DisplayQuantity       int64     `json:"displayQty"`
	ExDestination         string    `json:"exDestination"`
	ExecInst              string    `json:"execInst"`
	LeavesQty             int64     `json:"leavesQty"`
	MultiLegReportingType string    `json:"multiLegReportingType"`
	OrdRejReason          string    `json:"ordRejReason"`
	OrdStatus             string    `json:"ordStatus"`
	OrdType               string    `json:"ordType"`
	OrderID               string    `json:"orderID"`
	OrderQty              int64     `json:"orderQty"`
	PegOffsetValue        float64   `json:"pegOffsetValue"`
	PegPriceType          string    `json:"pegPriceType"`
	Price                 float64   `json:"price"`
	SettlCurrency         string    `json:"settlCurrency"`
	Side                  string    `json:"side"`
	SimpleCumQty          float64   `json:"simpleCumQty"`
	SimpleLeavesQty       float64   `json:"simpleLeavesQty"`
	SimpleOrderQty        float64   `json:"simpleOrderQty"`
	StopPx                float64   `json:"stopPx"`
	Symbol                string    `json:"symbol"`
	Text                  string    `json:"text"`
	TimeInForce           string    `json:"timeInForce"`
	Timestamp             time.Time `json:"timestamp"`
	TransactTime          string    `json:"transactTime"`
	Triggered             string    `json:"triggered"`
	WorkingIndicator      bool      `json:"workingIndicator"`
}

type TradeBuck struct {
	Symbol          string  `json:"symbol"`
	Timestamp       string  `json:"timestamp"`
	HomeNotional    float64 `json:"homeNotional"`
	ForeignNotional float64 `json:"foreignNotional"`
	Open            float64 `json:"open"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Close           float64 `json:"close"`
	Trades          int     `json:"trades"`
	Volume          int64   `json:"volume"`
	LastSize        int     `json:"lastSize"`
	Turnover        int64   `json:"turnover"`
	Vwap            float64 `json:"vwap"`
}
