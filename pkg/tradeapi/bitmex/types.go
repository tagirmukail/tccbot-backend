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

// Position Summary of Open and Closed Positions
type Position struct {
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
	Symbol               string    `json:"symbol"`
	TargetExcessMargin   int64     `json:"targetExcessMargin"`
	TaxBase              int64     `json:"taxBase"`
	TaxableMargin        int64     `json:"taxableMargin"`
	Timestamp            time.Time `json:"timestamp"`
	Underlying           string    `json:"underlying"`
	UnrealisedCost       int64     `json:"unrealisedCost"`
	UnrealisedGrossPnl   int64     `json:"unrealisedGrossPnl"`
	UnrealisedPnl        int64     `json:"unrealisedPnl"`
	UnrealisedPnlPcnt    float64   `json:"unrealisedPnlPcnt"`
	UnrealisedRoePcnt    float64   `json:"unrealisedRoePcnt"`
	UnrealisedTax        int64     `json:"unrealisedTax"`
	VarMargin            int64     `json:"varMargin"`
}

// Instrument Tradeable Contracts, Indices, and History
type Instrument struct {
	AskPrice                       float64   `json:"askPrice"`
	BankruptLimitDownPrice         float64   `json:"bankruptLimitDownPrice"`
	BankruptLimitUpPrice           float64   `json:"bankruptLimitUpPrice"`
	BidPrice                       float64   `json:"bidPrice"`
	BuyLeg                         string    `json:"buyLeg"`
	CalcInterval                   string    `json:"calcInterval"`
	Capped                         bool      `json:"capped"`
	ClosingTimestamp               time.Time `json:"closingTimestamp"`
	Deleverage                     bool      `json:"deleverage"`
	Expiry                         string    `json:"expiry"`
	FairBasis                      float64   `json:"fairBasis"`
	FairBasisRate                  float64   `json:"fairBasisRate"`
	FairMethod                     string    `json:"fairMethod"`
	FairPrice                      float64   `json:"fairPrice"`
	Front                          string    `json:"front"`
	FundingBaseSymbol              string    `json:"fundingBaseSymbol"`
	FundingInterval                string    `json:"fundingInterval"`
	FundingPremiumSymbol           string    `json:"fundingPremiumSymbol"`
	FundingQuoteSymbol             string    `json:"fundingQuoteSymbol"`
	FundingRate                    float64   `json:"fundingRate"`
	FundingTimestamp               time.Time `json:"fundingTimestamp"`
	HasLiquidity                   bool      `json:"hasLiquidity"`
	HighPrice                      float64   `json:"highPrice"`
	ImpactAskPrice                 float64   `json:"impactAskPrice"`
	ImpactBidPrice                 float64   `json:"impactBidPrice"`
	ImpactMidPrice                 float64   `json:"impactMidPrice"`
	IndicativeFundingRate          float64   `json:"indicativeFundingRate"`
	IndicativeSettlePrice          float64   `json:"indicativeSettlePrice"`
	IndicativeTaxRate              float64   `json:"indicativeTaxRate"`
	InitMargin                     float64   `json:"initMargin"`
	InsuranceFee                   float64   `json:"insuranceFee"`
	InverseLeg                     string    `json:"inverseLeg"`
	IsInverse                      bool      `json:"isInverse"`
	IsQuanto                       bool      `json:"isQuanto"`
	LastChangePcnt                 float64   `json:"lastChangePcnt"`
	LastPrice                      float64   `json:"lastPrice"`
	LastPriceProtected             float64   `json:"lastPriceProtected"`
	LastTickDirection              string    `json:"lastTickDirection"`
	Limit                          float64   `json:"limit"`
	LimitDownPrice                 float64   `json:"limitDownPrice"`
	LimitUpPrice                   float64   `json:"limitUpPrice"`
	Listing                        string    `json:"listing"`
	LotSize                        int64     `json:"lotSize"`
	LowPrice                       float64   `json:"lowPrice"`
	MaintMargin                    float64   `json:"maintMargin"`
	MakerFee                       float64   `json:"makerFee"`
	MarkMethod                     string    `json:"markMethod"`
	MarkPrice                      float64   `json:"markPrice"`
	MaxOrderQty                    int64     `json:"maxOrderQty"`
	MaxPrice                       float64   `json:"maxPrice"`
	MidPrice                       float64   `json:"midPrice"`
	Multiplier                     int64     `json:"multiplier"`
	OpenInterest                   int64     `json:"openInterest"`
	OpenValue                      int64     `json:"openValue"`
	OpeningTimestamp               time.Time `json:"openingTimestamp"`
	OptionMultiplier               float64   `json:"optionMultiplier"`
	OptionStrikePcnt               float64   `json:"optionStrikePcnt"`
	OptionStrikePrice              float64   `json:"optionStrikePrice"`
	OptionStrikeRound              float64   `json:"optionStrikeRound"`
	OptionUnderlyingPrice          float64   `json:"optionUnderlyingPrice"`
	PositionCurrency               string    `json:"positionCurrency"`
	PrevClosePrice                 float64   `json:"prevClosePrice"`
	PrevPrice24h                   float64   `json:"prevPrice24h"`
	PrevTotalTurnover              int64     `json:"prevTotalTurnover"`
	PrevTotalVolume                int64     `json:"prevTotalVolume"`
	PublishInterval                string    `json:"publishInterval"`
	PublishTime                    string    `json:"publishTime"`
	QuoteCurrency                  string    `json:"quoteCurrency"`
	QuoteToSettleMultiplier        int64     `json:"quoteToSettleMultiplier"`
	RebalanceInterval              string    `json:"rebalanceInterval"`
	RebalanceTimestamp             time.Time `json:"rebalanceTimestamp"`
	Reference                      string    `json:"reference"`
	ReferenceSymbol                string    `json:"referenceSymbol"`
	RelistInterval                 string    `json:"relistInterval"`
	RiskLimit                      int64     `json:"riskLimit"`
	RiskStep                       int64     `json:"riskStep"`
	RootSymbol                     string    `json:"rootSymbol"`
	SellLeg                        string    `json:"sellLeg"`
	SessionInterval                string    `json:"sessionInterval"`
	SettlCurrency                  string    `json:"settlCurrency"`
	Settle                         string    `json:"settle"`
	SettledPrice                   float64   `json:"settledPrice"`
	SettlementFee                  float64   `json:"settlementFee"`
	State                          string    `json:"state"`
	Symbol                         string    `json:"symbol"`
	TakerFee                       float64   `json:"takerFee"`
	Taxed                          bool      `json:"taxed"`
	TickSize                       float64   `json:"tickSize"`
	Timestamp                      time.Time `json:"timestamp"`
	TotalTurnover                  int64     `json:"totalTurnover"`
	TotalVolume                    int64     `json:"totalVolume"`
	Turnover                       int64     `json:"turnover"`
	Turnover24h                    int64     `json:"turnover24h"`
	Typ                            string    `json:"typ"`
	Underlying                     string    `json:"underlying"`
	UnderlyingSymbol               string    `json:"underlyingSymbol"`
	UnderlyingToPositionMultiplier int64     `json:"underlyingToPositionMultiplier"`
	UnderlyingToSettleMultiplier   int64     `json:"underlyingToSettleMultiplier"`
	Volume                         float64   `json:"volume"`
	Volume24h                      float64   `json:"volume24h"`
	Vwap                           float64   `json:"vwap"`
}
