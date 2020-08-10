package orderproc

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const (
	limitBalanceContracts = 200
	limitMinOnOrderQty    = 100
	liquidationPriceLimit = 1600
)

type OrderProcessor struct {
	tickPeriod      time.Duration
	api             tradeapi.API
	log             *logrus.Logger
	cfg             *config.GlobalConfig
	mx              sync.Mutex
	currentPosition *bitmex.Position
}

func New(
	api tradeapi.API, cfg *config.GlobalConfig, log *logrus.Logger,
) *OrderProcessor {
	rand.Seed(time.Now().UnixNano())
	return &OrderProcessor{
		tickPeriod: time.Duration(cfg.OrdProcPeriodSec) * time.Second,
		api:        api,
		cfg:        cfg,
		log:        log,
	}
}

func (o *OrderProcessor) SetPosition(p *bitmex.Position) {
	o.mx.Lock()
	defer o.mx.Unlock()
	if p == nil {
		return
	}

	if o.currentPosition == nil {
		o.currentPosition = &bitmex.Position{}
	}

	var avgPrice = o.currentPosition.AvgCostPrice
	if o.currentPosition.CurrentQty != p.CurrentQty || o.currentPosition.LiquidationPrice != p.LiquidationPrice {
		o.currentPosition = p
	}

	if avgPrice == 0 || p.AvgCostPrice == 0 {
		symbol := o.currentPosition.Symbol
		if symbol == "" {
			symbol = p.Symbol
		}
		position, err := o.getPosition(symbol)
		if err != nil {
			o.log.Errorf("OrderProcessor.getPosition() failed: %v", err)
			return
		}

		o.currentPosition.AvgCostPrice = position.AvgCostPrice
		o.currentPosition.AvgEntryPrice = position.AvgEntryPrice
		o.currentPosition.LastPrice = position.LastPrice
		o.currentPosition.CurrentQty = position.CurrentQty
	}
}

func (o *OrderProcessor) GetPosition() (*bitmex.Position, bool) {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.currentPosition, o.currentPosition != nil
}

func (o *OrderProcessor) PlaceOrder(
	exchange types.Exchange,
	side types.Side,
	amount float64,
	passive bool,
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
		inst, err := o.getInstrument()
		if err != nil {
			return nil, err
		}
		var price float64
		if side == types.SideSell {
			price = inst.AskPrice
		} else {
			price = inst.BidPrice
		}

		if amount == 0 {
			err = o.checkLiquidation(price, side)
			if err != nil {
				return nil, err
			}
			amount, err = o.calcOrderQty(
				availableBalance,
				side,
			)
			if err != nil {
				return nil, err
			}
		}

		params := &bitmex.OrderNewParams{
			Symbol:    o.cfg.ExchangesSettings.Bitmex.Symbol,
			Side:      string(side),
			OrderType: string(o.cfg.ExchangesSettings.Bitmex.OrderType),
			OrderQty:  math.Round(amount),
			Price:     price,
		}
		if passive {
			params.ExecInst = string(types.PassiveOrderExecInstType)
		}
		o.log.Infof("create order params: %#v", params)
		order, err = o.api.GetBitmex().CreateOrder(params)
		if err != nil {
			return nil, err
		}

		return order, nil
	default:
		return nil, fmt.Errorf("unknown exchange: %s", exchange)
	}
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

func (o *OrderProcessor) getInstrument() (bitmex.Instrument, error) {
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

func (o *OrderProcessor) getPosition(symbol string) (bitmex.Position, error) {
	positions, err := o.api.GetBitmex().GetPositions(bitmex.PositionGetParams{
		Columns: "avgCostPrice,avgEntryPrice,currentQty,lastPrice",
	})
	if err != nil {
		return bitmex.Position{}, err
	}

	for _, pos := range positions {
		if pos.Symbol == symbol {
			return pos, nil
		}
	}

	return bitmex.Position{}, nil
}

func (o *OrderProcessor) checkLimitContracts(side types.Side) error {
	currentPosition, ok := o.GetPosition()
	if !ok {
		return nil
	}
	switch side {
	case types.SideSell:
		isLimitedShort := currentPosition.CurrentQty <= -int64(o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		if isLimitedShort {
			return fmt.Errorf("place sell order limitted - qty: %d, limit: %d",
				o.currentPosition.CurrentQty, o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		}
	case types.SideBuy:
		isLimitedLong := currentPosition.CurrentQty >= int64(o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		if isLimitedLong {
			return fmt.Errorf("place buy order limitted - qty: %d, limit: %d",
				o.currentPosition.CurrentQty, o.cfg.ExchangesSettings.Bitmex.LimitContractsCount)
		}
	default:
		break
	}
	return nil
}

// calcOrderQty in contracts
func (o *OrderProcessor) calcOrderQty(balance float64, side types.Side) (qtyContrts float64, err error) {
	position, ok := o.GetPosition()
	if ok {
		if position.CurrentQty > 0 {
			qtyContrts = math.Abs(float64(position.CurrentQty))
			return
		}
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

//
func (o *OrderProcessor) checkLiquidation(price float64, side types.Side) error {
	position, ok := o.GetPosition()
	if !ok {
		return nil
	}
	liquidationDiff := math.Abs(position.LiquidationPrice - price)
	isLiquidationWarn := liquidationDiff < liquidationPriceLimit
	if position.CurrentQty < 0 &&
		side == types.SideSell &&
		isLiquidationWarn {
		return errors.New("liquidation warning triggered; further placing of sell orders suspended")
	}
	if position.CurrentQty > 0 &&
		side == types.SideBuy &&
		isLiquidationWarn {
		return errors.New("liquidation warning triggered; further placing of buy orders suspended")
	}
	return nil
}
