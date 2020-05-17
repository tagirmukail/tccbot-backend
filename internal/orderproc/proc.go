package orderproc

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type OrderProcessor struct {
	tickPeriod time.Duration
	api        tradeapi.Api
	log        *logrus.Logger
	cfg        *config.GlobalConfig
}

func New(api tradeapi.Api, cfg *config.GlobalConfig, log *logrus.Logger) *OrderProcessor {
	rand.Seed(time.Now().UnixNano())
	return &OrderProcessor{
		tickPeriod: time.Duration(cfg.OrdProcPeriodSec) * time.Second,
		api:        api,
		cfg:        cfg,
		log:        log,
	}
}

func (o *OrderProcessor) Start(wg *sync.WaitGroup) {
	o.log.Infof("order processor started")
	defer func() {
		o.log.Infof("order processor finished")
		wg.Done()
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(o.tickPeriod)
	defer tick.Stop()
	for {
		select {
		case <-done:
			return
		case <-tick.C:
			err := o.procActiveOrders()
			if err != nil {
				o.log.Warnf("procActiveOrders failed: %v", err)
			}
			err = o.processPosition()
			if err != nil {
				o.log.Warnf("processPosition failed: %v", err)
			}
		}
	}

}

func (o *OrderProcessor) processPosition() error {
	o.log.Info("\n===============================\nstart process position")
	defer o.log.Info("finish process position\n===============================\n")

	positions, err := o.api.GetBitmex().GetPositions(bitmex.PositionGetParams{
		Filter: fmt.Sprintf(`{"symbol": "%s"}`, o.cfg.ExchangesSettings.Bitmex.Symbol),
	})
	if err != nil {
		return err
	}
	for _, position := range positions {
		if trademath.ConvertToBTC(position.UnrealisedPnl) >= o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC {
			if position.OpeningQty > 0 {
				_, err := o.PlaceOrder(types.Bitmex, types.SideSell, math.Abs(float64(position.OpeningQty)), true)
				if err != nil {
					return err
				}
			} else if position.OpeningQty < 0 {
				_, err := o.PlaceOrder(types.Bitmex, types.SideBuy, math.Abs(float64(position.OpeningQty)), true)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (o *OrderProcessor) procActiveOrders() error {
	o.log.Infof("\n===============================\nstart process active orders")
	defer o.log.Infof("finish process active orders\n===============================\n")

	filter := fmt.Sprintf(`{"ordStatus":"%s"}`, types.OrdNew)
	orders, err := o.api.GetBitmex().GetOrders(&bitmex.OrdersRequest{
		Symbol: o.cfg.ExchangesSettings.Bitmex.Symbol,
		Filter: filter,
	})
	if err != nil {
		return err
	}

	for _, order := range orders {
		duration := time.Now().UTC().Sub(order.Timestamp)
		o.log.Debugf("order:%v duration: %v", order.OrderID, duration)
		if duration > o.tickPeriod {
			inst, err := o.getPrices()
			if err != nil {
				return err
			}
			var price float64
			switch types.Side(order.Side) {
			case types.SideSell:
				price = inst.AskPrice
			case types.SideBuy:
				price = inst.BidPrice
			}
			ord, err := o.api.GetBitmex().AmendOrder(&bitmex.OrderAmendParams{
				OrderID: order.OrderID,
				Price:   price,
				Text:    "amend order - proc active orders",
			})
			if err != nil {
				o.log.Errorf("failed amend order:id:%s, error: %v", order.OrderID, err)
				continue
			}
			o.log.Debugf("amend order successfully completed: %#v", ord)
		}
	}

	return nil
}

func (o *OrderProcessor) PlaceOrder(
	exchange types.Exchange,
	side types.Side,
	amount float64,
	passive bool,
) (order interface{}, err error) {
	switch exchange {
	case types.Bitmex:
		inst, err := o.getPrices()
		if err != nil {
			return nil, err
		}
		var price float64
		if side == types.SideSell {
			price = inst.AskPrice
		} else {
			price = inst.BidPrice
		}

		params := &bitmex.OrderNewParams{
			Symbol:    o.cfg.ExchangesSettings.Bitmex.Symbol,
			Side:      string(side),
			OrderType: string(o.cfg.ExchangesSettings.Bitmex.OrderType),
			OrderQty:  math.Round(amount),
			Price:     price,
		}
		if passive {
			params.ExecInst = "ParticipateDoNotInitiate"
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
