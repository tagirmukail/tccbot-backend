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
	LimitPositionPnls      = 2
	expirePositionDuration = 5 * time.Minute
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
	api                  tradeapi.API
	orderProc            *orderproc.OrderProcessor
	log                  *logrus.Logger
	cfg                  *config.GlobalConfig
	bitmexDataSubscriber *betrayed.Subscriber
	pnlT                 positionPnl
	positionPnlLimit     int
}

// TODO смотреть изменения ордеров по ws, проверять активные, исполненые, отклоненые в procActiveOrders()
// TODO добавить в конфигурацию настройку для включения определенного шедулера
func NewPositionScheduler(
	cfg *config.GlobalConfig,
	positionPnlLimit int,
	api tradeapi.API,
	orderProc *orderproc.OrderProcessor,
	bitmexDataSubscriber *betrayed.Subscriber,
	log *logrus.Logger,
) *PositionScheduler {
	return &PositionScheduler{
		orderProc:            orderProc,
		api:                  api,
		log:                  log,
		cfg:                  cfg,
		bitmexDataSubscriber: bitmexDataSubscriber,
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

	activeOrdersTick := time.NewTicker(time.Duration(o.cfg.OrdProcPeriodSec) * time.Second)
	positionCleanTick := time.NewTicker(expirePositionDuration)
	defer func() {
		activeOrdersTick.Stop()
		positionCleanTick.Stop()
	}()
	for {
		select {
		case <-done:
			return
		case tradeData := <-o.bitmexDataSubscriber.GetMsgChan():
			o.log.Debugf("PositionScheduler.Start process data table: %#v", tradeData.Table)
			if tradeData.Table == string(types.Position) {
				o.processPosition(tradeData.Data)
			}
		case <-activeOrdersTick.C:
			err := o.procActiveOrders()
			if err != nil {
				o.log.Errorf("o.procActiveOrders() failed: %v", err)
			}
		case <-positionCleanTick.C:
			o.log.Debugf("clean position started")
			currentPosition, ok := o.orderProc.GetPosition()
			if !ok {
				o.log.Debugf("position already clear")
				continue
			}
			expTime := currentPosition.Timestamp.UTC().Add(expirePositionDuration)
			now := time.Now().UTC()
			if now.After(expTime) {
				o.log.Debugf("position cleaned now")
				o.orderProc.SetPosition(nil)
				continue
			}

			o.log.Debugf("clean position expire time not now")
		}
	}
}

func (o *PositionScheduler) Stop() error {
	return nil
}

func (o *PositionScheduler) processPosition(positions []data.BitmexIncomingData) {
	orders, err := getActiveOrders(o.api, o.cfg.ExchangesSettings.Bitmex.Symbol)
	if err != nil {
		o.log.Errorf("get active orders failed: %v", err)
		return
	}

	if len(orders) > 0 {
		o.log.Infoln("order already placed, wait")
		return
	}

	for _, positionData := range positions {
		o.log.Debugf("PositionScheduler.Start process data : %#v", positionData)
		var (
			position = &bitmex.Position{}
			err      error
		)
		if string(positionData.Symbol) == o.cfg.ExchangesSettings.Bitmex.Symbol {
			position, err = FromBitmexIncDataToPosition(positionData)
			if err != nil {
				o.log.Errorf("[bitmex exchange data]:%#v convert to position failed [err]:%v",
					positionData, err)
				continue
			}
		} else {
			continue
		}

		if position.UnrealisedPnl == 0 || positionData.CurrentQty == 0 {
			continue
		}

		o.orderProc.SetPosition(position)

		unrealisedPnl := trademath.ConvertToBTC(position.UnrealisedPnl)
		o.log.Debugf("current position [unrealised pnl in btc]: %.9f", unrealisedPnl)
		var pnlType = Neutral
		if unrealisedPnl >= o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC {
			pnlType = Profit
		} else if unrealisedPnl <= -o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC/3 {
			pnlType = Loss
		}

		o.processPnl(&positionPnl{
			pnl: unrealisedPnl,
			t:   pnlType,
		}, *position)
	}
}

func (o *PositionScheduler) processPnl(p *positionPnl, position bitmex.Position) {
	if !o.checkPlaceOrder(p) {
		return
	}

	ord, err := o.placeClosePositionOrder(position)
	if err != nil {
		o.log.Errorf("processPnl() placeClosePositionOrder() place %v order failed: %v", p.t, err)
		return
	}
	o.log.Debugf("processPnl() placed type:%v order: %#v", p.t, ord)
	o.clearPositionPnl()
}

// TODO unit tests
func (o *PositionScheduler) checkPlaceOrder(p *positionPnl) bool {
	var placeOrder bool
	switch {
	case o.pnlT.t == Profit && p.t == Loss:
		placeOrder = true
	case o.pnlT.pnl < p.pnl:
		o.log.Debugf("[o.pnlT.pnl < p.pnl] we are waiting to check the position, "+
			"[pnlT]: %#v, [p]: %#v",
			o.pnlT, p)
		o.pnlT = *p
	case o.pnlT.t == Profit && p.pnl+o.cfg.Scheduler.Position.ProfitPnlDiff <= o.pnlT.pnl:
		placeOrder = true
	case o.pnlT.t == Loss && p.pnl < o.pnlT.pnl-(o.cfg.Scheduler.Position.ProfitPnlDiff/3):
		placeOrder = true
	case o.pnlT.t == Loss && p.pnl > o.pnlT.pnl: // m. b. not needed
		o.log.Debugf("[o.pnlT.t == Loss && p.pnl > o.pnlT.pnl] we are waiting to check the position, ["+
			"pnlT]: %#v, [p]: %#v",
			o.pnlT, p)
		o.pnlT = *p
	case o.pnlT.t == Neutral && p.t == Loss:
		o.log.Debugf("[o.pnlT.t == Neutral && p.t == Loss] we are waiting to check the position, "+
			"[pnlT]: %#v, [p]: %#v",
			o.pnlT, p)
		o.pnlT = *p
	default:
		o.log.Debugf("[default] we are waiting to check the position, [pnlT]: %#v, [p]: %#v",
			o.pnlT, p)
	}

	return placeOrder
}

func (o *PositionScheduler) clearPositionPnl() {
	o.pnlT = positionPnl{}
}

func (o *PositionScheduler) placeClosePositionOrder(position bitmex.Position) (interface{}, error) {
	var side types.Side
	switch {
	case position.CurrentQty > 0:
		side = types.SideSell
	case position.CurrentQty < 0:
		side = types.SideBuy
	default:
		return nil, errors.New("qty is 0")
	}
	ord, err := o.orderProc.PlaceOrder(
		types.Bitmex, side, math.Abs(float64(position.CurrentQty)), true)
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
				price = inst.BidPrice + 2
			}
		case types.SideBuy:
			diff := inst.AskPrice - order.Price
			if math.Abs(diff) > o.cfg.Scheduler.Position.PriceTrailing {
				price = inst.AskPrice - 2
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
