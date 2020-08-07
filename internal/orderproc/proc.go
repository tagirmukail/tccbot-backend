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

	trailingOffset = 0.5
)

type OrderProcessor struct {
	tickPeriod      time.Duration
	api             tradeapi.Api
	log             *logrus.Logger
	cfg             *config.GlobalConfig
	mx              sync.Mutex
	currentPosition *bitmex.Position
}

func New(
	api tradeapi.Api, cfg *config.GlobalConfig, log *logrus.Logger,
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
	o.currentPosition = p
}

func (o *OrderProcessor) getPosition() (*bitmex.Position, bool) {
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

// placeTrailingStopOrder - place stop order
// example:{"ordType":"Stop","pegOffsetValue":-0.5,"pegPriceType":"TrailingStopPeg","orderQty":100,"side":"Sell","execInst":"LastPrice","symbol":"XBTUSD","text":"Submission from testnet.bitmex.com"}
func (o *OrderProcessor) placeTrailingStopOrder(
	clOrdID string,
	side types.Side,
	orderQty float64,
	passive bool,
) (order interface{}, err error) {
	var offset float64
	if side == types.SideBuy {
		offset = trailingOffset
	} else if side == types.SideSell {
		offset = -trailingOffset
	} else {
		return nil, fmt.Errorf("unknown side: %v", side)
	}

	stopParams := &bitmex.OrderNewParams{
		ClientOrderID:  clOrdID,
		Symbol:         o.cfg.ExchangesSettings.Bitmex.Symbol,
		OrderType:      string(types.Stop),
		Side:           string(side),
		PegPriceType:   string(types.TrailingStopPeg),
		PegOffsetValue: offset,
		OrderQty:       orderQty,
		ExecInst:       string(types.MarkPriceExecInstType),
	}
	if passive {
		stopParams.ExecInst = string(types.PassiveOrderExecInstType)
	}

	o.log.Infof("create trailing stop order params: %#v", stopParams)
	order, err = o.api.GetBitmex().CreateOrder(stopParams)
	if err != nil {
		return nil, err
	}
	o.log.Debugf("trailing stop order placed: %#v", order)

	return order, nil
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

func (o *OrderProcessor) checkLimitContracts(side types.Side) error {
	currentPosition, ok := o.getPosition()
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
	position, ok := o.getPosition()
	if !ok {
		return
	}
	positionPnlToBTC := trademath.ConvertToBTC(position.UnrealisedPnl)
	if positionPnlToBTC > 0 {
		if position.CurrentQty > 0 {
			qtyContrts = math.Abs(float64(position.CurrentQty))
			return
		}
		qtyContrts = float64(limitMinOnOrderQty)
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

//
func (o *OrderProcessor) checkLiquidation(price float64, side types.Side) error {
	position, ok := o.getPosition()
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
