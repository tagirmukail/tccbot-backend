package tradeapi

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const (
	defaultRetryCount            = 5
	defaultIdleTimeout           = 15 * time.Second
	defaultMaxIdleConns          = 10
	TradeBucketedTimestampLayout = "2006-01-02T15:04:05.000Z"
)

type TradeApi struct {
	bitmex *bitmex.Bitmex
}

func NewTradeApi(
	bitmexKey string,
	bitmexSecret string,
	log *logrus.Logger,
	test bool,
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
			log,
		),
	}
	if test {
		tapi.bitmex.EnableTestNet()
	}
	return tapi
}

func (t *TradeApi) GetBitmex() *bitmex.Bitmex {
	return t.bitmex
}
