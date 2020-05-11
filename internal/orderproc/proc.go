package orderproc

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"

	"github.com/tagirmukail/tccbot-backend/internal/trademath"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

// Следить за ордерами на бирже и перевыставлять при необходимости или отменять

type OrderProcessor struct {
	tickPeriod time.Duration
	api        tradeapi.Api
	log        *logrus.Logger
	cfg        *config.GlobalConfig
}

func New(tickPeriodSec uint32, api tradeapi.Api, cfg *config.GlobalConfig, log *logrus.Logger) *OrderProcessor {
	rand.Seed(time.Now().UnixNano())
	return &OrderProcessor{
		tickPeriod: time.Duration(tickPeriodSec) * time.Second,
		api:        api,
		cfg:        cfg,
		log:        log,
	}
}

// todo implement
func (o *OrderProcessor) procTick() {
	tick := time.NewTicker(o.tickPeriod)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			o.procActiveOrders()
		}
	}

}

func (o *OrderProcessor) procActiveOrders() error {
	filter := fmt.Sprintf(`{"ordStatus":"%s"}`, types.OrdNew)
	orders, err := o.api.GetBitmex().GetOrders(&bitmex.OrdersRequest{
		Symbol: o.cfg.ExchangesSettings.Bitmex.Symbol,
		Filter: filter,
	})
	if err != nil {
		return err
	}

	for _, order := range orders {
		if time.Now().UTC().Sub(order.Timestamp) > o.tickPeriod {
			_, err = o.api.GetBitmex().AmendOrder(&bitmex.OrderAmendParams{
				OrigClOrdID:          order.OrderID,
				PegOffsetValue:       0,
				Price:                0,
				SimpleLeavesQuantity: 0,
				SimpleOrderQuantity:  0,
				StopPx:               0,
				Text:                 "",
			})
			if err != nil {
				o.log.Errorf("failed amend order:id:%s, error: %v", order.OrderID, err)
				continue
			}
		}
	}

	return nil
}

func (o *OrderProcessor) PlaceOrder(
	exchange types.Exchange,
	signalType models.SignalType,
	side types.Side,
	amount float64,
	stopPx float64,
	close bool,
) (order interface{}, err error) {
	switch exchange {
	case types.Bitmex:
		inst, err := o.getPrices()
		if err != nil {
			return nil, err
		}
		var price float64
		if side == types.SideSell {
			if signalType == models.RSI {
				price = inst.AskPrice + inst.AskPrice*0.09
			} else {
				price = inst.AskPrice
			}
		} else {
			if signalType == models.RSI {
				price = inst.BidPrice + inst.BidPrice*0.09
			} else {
				price = inst.BidPrice
			}
		}

		params := &bitmex.OrderNewParams{
			Symbol:    o.cfg.ExchangesSettings.Bitmex.Symbol,
			Side:      string(side),
			OrderType: string(o.cfg.ExchangesSettings.Bitmex.OrderType),
			OrderQty:  math.Round(amount),
			Price:     price,
			StopPx:    stopPx,
		}
		if close {
			params.ExecInst = "Close"
		}
		o.log.Infof("create order params: %#v", params)
		order, err := o.api.GetBitmex().CreateOrder(params)
		if err != nil {
			return nil, err
		}
		return order, nil
	default:
		return nil, fmt.Errorf("unknown exchange: %s", exchange)
	}
}

func (o *OrderProcessor) GetBalance() (float64, error) {
	margins, err := o.api.GetBitmex().GetAllUserMargin()
	if err != nil {
		return 0, err
	}
	if len(margins) == 0 {
		return 0, errors.New("user margins not exist")
	}
	for _, margin := range margins {
		if o.cfg.ExchangesSettings.Bitmex.Currency == margin.Currency {
			return trademath.ConvertToBTC(margin.WalletBalance), nil
		}
	}

	return 0, fmt.Errorf("user margin by currency:%s not exist", o.cfg.ExchangesSettings.Bitmex.Currency)
}

func (o *OrderProcessor) getPrices() (bitmex.Instrument, error) {
	var resp bitmex.Instrument
	insts, err := o.api.GetBitmex().GetInstrument(bitmex.InstrumentRequestParams{
		Symbol:  o.cfg.ExchangesSettings.Bitmex.Symbol,
		Columns: "lastPrice,bidPrice,midPrice,askPrice,markPrice",
		Count:   1,
	})
	if err != nil {
		return resp, err
	}
	if len(insts) == 0 {
		return resp, errors.New("instruments not exist")
	}
	return insts[0], nil
}
