package bitmex

const (
	Unset EndpointLimit = iota
	Auth
	UnAuth
)

const (
	bitmexURL  = "https://www.bitmex.com/api/v1"
	testnetURL = "https://testnet.bitmex.com/api/v1"

	userAgent = "User-Agent"

	maxRequests int32 = 50

	TradeTimeFormat = "2006-01-02 15:04"
)

const (
	// endpoints
	endpointUserMargin    = "/user/margin"
	endpointUserWallet    = "/user/wallet"
	endpointOrder         = "/order"
	endpointAllOrders     = "/order/all"
	endpointTradeBucketed = "/trade/bucketed"

	endpointLeveragePosition = "/position/leverage"
	endpointPosition         = "/position"

	endpointInstrument = "/instrument"
)
