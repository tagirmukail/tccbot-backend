package tradeapi

import (
	"net/url"
	"time"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws"

	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const (
	defaultRetryCount            = 5
	defaultIdleTimeout           = 15 * time.Second
	defaultMaxIdleConns          = 10
	TradeBucketedTimestampLayout = "2006-01-02T15:04:05.000Z"
)

type Api interface {
	GetBitmex() BitmexApi
}

type BitmexApi interface {
	EnableTestNet()
	GetWS() *ws.WS

	// Bitmex
	SetDefaultUserAgent(agent string)
	SendRequest(path string, params url.Values, response interface{}) error
	SendAuthenticatedRequest(verb, path string, params, response interface{}) error
	GetUserMargin(currency string) (bitmex.UserMargin, error)
	GetAllUserMargin() ([]bitmex.UserMargin, error)
	GetUserWalletInfo(currency string) (bitmex.WalletInfo, error)
	GetOrders(params *bitmex.OrdersRequest) ([]bitmex.OrderCopied, error)
	CreateOrder(params *bitmex.OrderNewParams) (bitmex.OrderCopied, error)
	AmendOrder(params *bitmex.OrderAmendParams) (bitmex.OrderCopied, error)
	CancelOrders(params *bitmex.OrderCancelParams) ([]bitmex.OrderCopied, error)
	CancelAllOrders(params *bitmex.OrderCancelAllParams) ([]bitmex.OrderCopied, error)
	GetTradeBucketed(params *bitmex.TradeGetBucketedParams) ([]bitmex.TradeBuck, error)
	LeveragePosition(params *bitmex.PositionUpdateLeverageParams) (bitmex.Position, error)
	GetPositions(params bitmex.PositionGetParams) ([]bitmex.Position, error)
	GetInstrument(params bitmex.InstrumentRequestParams) ([]bitmex.Instrument, error)
}

type TradeApi struct {
	bitmex BitmexApi
}

func NewTradeApi(
	bitmexKey,
	bitmexSecret string,
	log *logrus.Logger,
	test bool,
	ws *ws.WS,
) *TradeApi {
	tapi := &TradeApi{
		bitmex: bitmex.New(
			bitmexKey,
			bitmexSecret,
			true,
			defaultRetryCount,
			defaultIdleTimeout,
			defaultMaxIdleConns,
			0,
			0,
			ws,
			log,
		),
	}
	if test {
		tapi.bitmex.EnableTestNet()
	}
	return tapi
}

func (t *TradeApi) GetBitmex() BitmexApi {
	return t.bitmex
}
