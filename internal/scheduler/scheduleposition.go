package scheduler

import (
	"errors"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	betrayed "github.com/tagirmukail/tccbot-backend/internal/tradedata/bitmex"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

const (
	LimitPositionPnls = 2
	trailingPrice     = 5
)

type PnlType uint8

const (
	Neutral PnlType = iota
	Profit
	Loss
)

type positionPnl struct {
	pnl float64
	t   PnlType
}

type PositionScheduler struct {
	api                  tradeapi.Api
	orderProc            *orderproc.OrderProcessor
	log                  *logrus.Logger
	cfg                  *config.GlobalConfig
	bitmexDataSubscriber *betrayed.Subscriber
	mx                   sync.Mutex
	positionPnl          []*positionPnl
	positionPnlLimit     int
}

// TODO добавить в конфигурацию настройку для включения определенного шедулера
func NewPositionScheduler(
	cfg *config.GlobalConfig,
	positionPnlLimit int,
	api tradeapi.Api,
	orderProc *orderproc.OrderProcessor,
	bitmexDataSubscriber *betrayed.Subscriber,
	log *logrus.Logger,
) *PositionScheduler {
	return &PositionScheduler{
		orderProc:            orderProc,
		api:                  api,
		log:                  log,
		cfg:                  cfg,
		mx:                   sync.Mutex{},
		bitmexDataSubscriber: bitmexDataSubscriber,
		positionPnl:          []*positionPnl{},
		positionPnlLimit:     positionPnlLimit,
	}
}

func (o *PositionScheduler) Start(wg *sync.WaitGroup) {
	o.log.Infof("position scheduler started")
	defer func() {
		o.log.Infof("position scheduler finished")
		wg.Done()
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(time.Duration(o.cfg.OrdProcPeriodSec) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-done:
			return
		case tradeData := <-o.bitmexDataSubscriber.GetMsgChan():
			switch tradeData.Table {
			case string(types.Position):
				o.processPosition(tradeData.Data)
			}
		case <-tick.C:
			err := o.procActiveOrders()
			if err != nil {
				o.log.Errorf("o.procActiveOrders() failed: %v", err)
			}
		}
	}
}

func (o *PositionScheduler) Stop() error {
	return nil
}

func (o *PositionScheduler) processPosition(positions []data.BitmexIncomingData) {
	for _, positionData := range positions {
		var (
			position = &bitmex.Position{}
			err      error
		)
		if string(positionData.Symbol) == o.cfg.ExchangesSettings.Bitmex.Symbol {
			position, err = FromBitmexIncDataToPosition(&positionData)
			if err != nil {
				o.log.Errorf("[bitmex exchange data]:%#v convert to position failed [err]:%v",
					positionData, err)
				continue
			}
		}

		unrealisedPnl := trademath.ConvertToBTC(position.UnrealisedPnl)
		o.log.Debugf("current position [unrealised pnl in btc]: %.9f", unrealisedPnl)
		var pnlType = Neutral
		if unrealisedPnl >= o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC {
			pnlType = Profit
		} else if unrealisedPnl <= -o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC/3 {
			pnlType = Loss
		}

		o.addToPositionPnl(&positionPnl{
			pnl: unrealisedPnl,
			t:   pnlType,
		})

		o.checkProfitPnlList(*position)
	}
}

func (o *PositionScheduler) addToPositionPnl(p *positionPnl) {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.positionPnl = append(o.positionPnl, p)
	if len(o.positionPnl) > o.positionPnlLimit {
		o.positionPnl = o.positionPnl[len(o.positionPnl)-o.positionPnlLimit:]
	}
}

func (o *PositionScheduler) clearPositionPnl() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.positionPnl = make([]*positionPnl, 0, o.positionPnlLimit)
}

func (o *PositionScheduler) checkProfitPnlList(position bitmex.Position) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.positionPnl) < o.positionPnlLimit {
		return
	}

	switch {
	case o.positionPnl[0].t == Loss &&
		(o.positionPnl[1].t == Profit || o.positionPnl[1].t == Neutral):
		// we are waiting to check the position
		o.log.Debugf("we are waiting to check the position, [pnl1]: %#v, [pnl2]: %#v",
			o.positionPnl[0], o.positionPnl[1])
		return
	case o.positionPnl[0].t == Loss && o.positionPnl[1].t == Loss:
		ord, err := o.placeClosePositionOrder(position)
		if err != nil {
			o.log.Errorln("checkProfitPnlList() placeClosePositionOrder() stop loss failed: %v", err)
			return
		}
		o.log.Debugf("checkProfitPnlList() placed stop loss order: %#v", ord)
		o.clearPositionPnl()
		return
	case o.positionPnl[0].t == Neutral &&
		(o.positionPnl[1].t == Profit || o.positionPnl[1].t == Loss):
		// position has improved or started to deteriorate, we continue to monitor
		o.log.Debugf("we are waiting to check the position, [pnl1]: %#v, [pnl2]: %#v",
			o.positionPnl[0], o.positionPnl[1])
		return
	case o.positionPnl[0].t == Profit && o.positionPnl[1].t == Profit:
		if o.positionPnl[0].pnl > o.positionPnl[1].pnl {
			ord, err := o.placeClosePositionOrder(position)
			if err != nil {
				o.log.Errorln("checkProfitPnlList() placeClosePositionOrder() profit order failed: %v", err)
				return
			}
			o.log.Debugf("checkProfitPnlList() placed profit order: %#v", ord)
			o.clearPositionPnl()
			return
		}
		// we are waiting to check the position
		o.log.Debugf("we are waiting to check the position, [pnl1]: %#v, [pnl2]: %#v",
			o.positionPnl[0], o.positionPnl[1])
		return
	case o.positionPnl[0].t == Profit && o.positionPnl[1].t == Loss:
		ord, err := o.placeClosePositionOrder(position)
		if err != nil {
			o.log.Errorln("checkProfitPnlList() placeClosePositionOrder() stop loss failed: %v", err)
			return
		}
		o.log.Debugf("checkProfitPnlList() placed stop loss order: %#v", ord)
		o.clearPositionPnl()
		return
	case o.positionPnl[0].t == Profit && o.positionPnl[1].t == Neutral:
		// we are waiting to check the position
		o.log.Debugf("we are waiting to check the position, [pnl1]: %#v, [pnl2]: %#v",
			o.positionPnl[0], o.positionPnl[1])
		return
	}
}

func (o *PositionScheduler) placeClosePositionOrder(position bitmex.Position) (interface{}, error) {
	var side types.Side
	if position.CurrentQty > 0 {
		side = types.SideSell
	} else if position.CurrentQty < 0 {
		side = types.SideBuy
	} else {
		return nil, errors.New("qty is 0")
	}
	ord, err := o.orderProc.PlaceOrder(
		types.Bitmex, side, math.Abs(float64(position.CurrentQty)), true, false)
	if err != nil {
		return nil, err
	}
	return ord, nil
}

func (o *PositionScheduler) procActiveOrders() error {
	orders, err := getActiveOrders(o.api, o.cfg.ExchangesSettings.Bitmex.Symbol)
	if err != nil {
		return err
	}
	for _, order := range orders {
		inst, err := getInstrument(o.api, o.cfg.ExchangesSettings.Bitmex.Symbol)
		if err != nil {
			return err
		}
		var price float64
		switch types.Side(order.Side) {
		case types.SideSell:
			diff := inst.BidPrice - order.Price
			if math.Abs(diff) > o.cfg.Scheduler.Position.PriceTrailing {
				price = inst.BidPrice
			}
		case types.SideBuy:
			diff := inst.AskPrice - order.Price
			if math.Abs(diff) > o.cfg.Scheduler.Position.PriceTrailing {
				price = inst.AskPrice
			}
		}

		if price == 0 {
			o.log.Debugf("[order]: %v not need change [price]: %v", order.OrderID, order.Price)
			return nil
		}

		ord, err := o.api.GetBitmex().AmendOrder(&bitmex.OrderAmendParams{
			OrderID: order.OrderID,
			Price:   price,
			Text:    "amend order - proc active orders",
		})
		if err != nil {
			return err
		}

		o.log.Debugf("[order]: %v price changed to %v", ord.OrderID, price)
	}

	return nil
}
