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

const (
	PassiveOrderType      = "ParticipateDoNotInitiate"
	limitBalanceContracts = 200
	limitMinOnOrderQty    = 100
	liquidationPriceLimit = 1600

	priceOffset    = 10
	trailingOffset = 3
)

type OrderProcessor struct {
	tickPeriod      time.Duration
	api             tradeapi.Api
	log             *logrus.Logger
	cfg             *config.GlobalConfig
	currentPosition bitmex.Position
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
		if position.Symbol == o.cfg.ExchangesSettings.Bitmex.Symbol {
			o.currentPosition = position
		}
		if trademath.ConvertToBTC(position.UnrealisedPnl) >= o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC {
			if position.OpeningQty > 0 {
				_, err := o.PlaceOrder(types.Bitmex, types.SideSell, math.Abs(float64(position.OpeningQty)), true, false)
				if err != nil {
					return err
				}
			} else if position.OpeningQty < 0 {
				_, err := o.PlaceOrder(types.Bitmex, types.SideBuy, math.Abs(float64(position.OpeningQty)), true, false)
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

	filter := fmt.Sprintf(`{"open": %t}`, true)
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
	withStop bool,
) (order interface{}, err error) {
	switch exchange {
	case types.Bitmex:
		_, availableBalance, err := o.GetBalance()
		if err != nil {
			return nil, err
		}
		contracts := trademath.ConvertFromBTCToContracts(availableBalance)
		if contracts <= limitBalanceContracts {
			return nil, fmt.Errorf("balance is exhausted, %.3f left", availableBalance)
		}
		err = o.checkLimitContracts(side)
		if err != nil {
			return nil, err
		}
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

		var position *bitmex.Position
		if amount == 0 {
			position, err = o.getPosition()
			if err != nil {
				return nil, err
			}
			err = o.checkLiquidation(position, price, side)
			if err != nil {
				return nil, err
			}
			amount, err = o.calcOrderQty(
				position,
				availableBalance,
				side,
			)
			if err != nil {
				return nil, err
			}
		}
		if position == nil {
			position = &o.currentPosition
		}

		params := &bitmex.OrderNewParams{
			Symbol:    o.cfg.ExchangesSettings.Bitmex.Symbol,
			Side:      string(side),
			OrderType: string(o.cfg.ExchangesSettings.Bitmex.OrderType),
			OrderQty:  math.Round(amount),
			Price:     price,
		}
		if passive {
			params.ExecInst = PassiveOrderType
		}
		o.log.Infof("create order params: %#v", params)
		order, err := o.api.GetBitmex().CreateOrder(params)
		if err != nil {
			return nil, err
		}

		if withStop {
			err = o.placeStopOrder(params, types.LimitIfTouched, types.TrailingStopPeg, position, inst)
			if err != nil {
				return nil, err
			}
		}

		return order, nil
	default:
		return nil, fmt.Errorf("unknown exchange: %s", exchange)
	}
}

// placeStopOrder - place stop order
// TODO: needed manual and unit or integration testing
func (o *OrderProcessor) placeStopOrder(
	params *bitmex.OrderNewParams,
	ordType types.OrderType,
	priceType types.PriceType,
	_ *bitmex.Position,
	instrument bitmex.Instrument,
) error {
	var (
		stopPrice float64
		stopSide  types.Side
		offset    float64
	)
	if types.Side(params.Side) == types.SideBuy {
		stopSide = types.SideSell
		stopPrice = instrument.AskPrice - priceOffset
		offset = -1 * trailingOffset
	} else if types.Side(params.Side) == types.SideSell {
		stopSide = types.SideBuy
		stopPrice = instrument.BidPrice + priceOffset
		offset = trailingOffset
	} else {
		return fmt.Errorf("unknown side: %v", params.Side)
	}

	stopParams := &bitmex.OrderNewParams{
		Symbol:         params.Symbol,
		StopPx:         stopPrice,
		Price:          stopPrice,
		OrderType:      string(ordType),
		Side:           string(stopSide),
		PegPriceType:   string(priceType),
		PegOffsetValue: offset,
		OrderQty:       params.OrderQty,
	}

	o.log.Infof("create trailing stop order params: %#v", stopParams)
	order, err := o.api.GetBitmex().CreateOrder(stopParams)
	if err != nil {
		return err
	}
	o.log.Debugf("trailing stop order placed: %#v", order)

	return nil
}

func (o *OrderProcessor) GetBalance() (walletBalance, availableBalance float64, err error) {
	margins, err := o.api.GetBitmex().GetAllUserMargin()
	if err != nil {
		return 0, 0, err
	}
	if len(margins) == 0 {
		return 0, 0, errors.New("user margins not exist")
	}
	for _, margin := range margins {
		if o.cfg.ExchangesSettings.Bitmex.Currency == margin.Currency {
			walletBalance = trademath.ConvertToBTC(margin.WalletBalance)
			availableBalance = trademath.ConvertToBTC(margin.AvailableMargin)
			return walletBalance, availableBalance, nil
		}
	}

	return 0, 0, fmt.Errorf("user margin by currency:%s not exist", o.cfg.ExchangesSettings.Bitmex.Currency)
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

func (o *OrderProcessor) checkLimitContracts(side types.Side) error {
	switch side {
	case types.SideSell:
		isLimitedShort := o.currentPosition.CurrentQty <= -int64(o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		if isLimitedShort {
			return fmt.Errorf("place sell order limitted - qty: %d, limit: %d",
				o.currentPosition.CurrentQty, o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		}
	case types.SideBuy:
		isLimitedLong := o.currentPosition.CurrentQty >= int64(o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		if isLimitedLong {
			return fmt.Errorf("place buy order limitted - qty: %d, limit: %d",
				o.currentPosition.CurrentQty, o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		}
	default:
		break
	}
	return nil
}

func (o *OrderProcessor) getPosition() (*bitmex.Position, error) {
	positions, err := o.api.GetBitmex().GetPositions(bitmex.PositionGetParams{
		Filter: fmt.Sprintf(`{"symbol": "%s"}`, o.cfg.ExchangesSettings.Bitmex.Symbol),
	})
	if err != nil {
		return nil, err
	}
	for _, position := range positions {
		if position.Symbol == o.cfg.ExchangesSettings.Bitmex.Symbol {
			o.currentPosition = position
			return &position, nil
		}
	}

	return nil, errors.New("position not exist")
}

// calcOrderQty in contracts
func (o *OrderProcessor) calcOrderQty(position *bitmex.Position, balance float64, side types.Side) (qtyContrts float64, err error) {
	positionPnlToBTC := trademath.ConvertToBTC(position.UnrealisedPnl)
	if positionPnlToBTC > 0 {
		qtyContrts = math.Abs(float64(position.OpeningQty))
		return
	}

	var qtyBtc float64
	switch side {
	case types.SideBuy:
		qtyBtc = balance * o.cfg.ExchangesSettings.Bitmex.BuyOrderCoef
	case types.SideSell:
		qtyBtc = balance * o.cfg.ExchangesSettings.Bitmex.SellOrderCoef
	default:
		err = fmt.Errorf("unknown side type: %s", side)
		return
	}

	qtyContrts = trademath.ConvertFromBTCToContracts(qtyBtc)

	if qtyContrts < float64(limitMinOnOrderQty) {
		qtyContrts = float64(limitMinOnOrderQty)
	}

	qtyContrts = math.Round(qtyContrts)
	return
}

func (o *OrderProcessor) checkLiquidation(position *bitmex.Position, price float64, side types.Side) error {
	liquidationDiff := math.Abs(position.LiquidationPrice - price)
	isLiquidationWarn := liquidationDiff < liquidationPriceLimit
	if position.OpeningQty < 0 &&
		side == types.SideSell &&
		isLiquidationWarn {
		return errors.New("liquidation warning triggered; further placing of sell orders suspended")
	}
	if position.OpeningQty > 0 &&
		side == types.SideBuy &&
		isLiquidationWarn {
		return errors.New("liquidation warning triggered; further placing of buy orders suspended")
	}
	return nil
}
