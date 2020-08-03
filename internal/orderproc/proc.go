package orderproc

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	bitmextradedata "github.com/tagirmukail/tccbot-backend/internal/tradedata/bitmex"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"

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
	tickPeriod           time.Duration
	api                  tradeapi.Api
	log                  *logrus.Logger
	cfg                  *config.GlobalConfig
	currentPosition      bitmex.Position
	bitmexDataSubscriber *bitmextradedata.Subscriber
}

func New(
	api tradeapi.Api, cfg *config.GlobalConfig, bitmexDataSubscriber *bitmextradedata.Subscriber, log *logrus.Logger,
) *OrderProcessor {
	rand.Seed(time.Now().UnixNano())
	return &OrderProcessor{
		tickPeriod:           time.Duration(cfg.OrdProcPeriodSec) * time.Second,
		api:                  api,
		cfg:                  cfg,
		log:                  log,
		bitmexDataSubscriber: bitmexDataSubscriber,
	}
}

// TODO: если не реализованный пнл больше ClosePositionMinBTC, то сразу не закрываем позицию, а ждем максимальную прибыль, когда не реализованый пнл уменьшается, то закрываем позицию
// TODO: добавить получение информации о своих ордерах через ws
// TODO: вынести офсет для стоп ордера в конфиг

//func (o *OrderProcessor) proc() {
//	o.log.Info("\n===============================\nstart process position")
//	defer o.log.Info("finish process position\n===============================\n")
//
//	done := make(chan os.Signal, 1)
//	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
//
//	for {
//		select {
//		case <-done:
//			return
//		case tradeData := <-o.bitmexDataSubscriber.GetMsgChan():
//			o.log.Debugf("bitmex auth ws receiver [trade data]:%#v", tradeData)
//			switch tradeData.Table {
//			case string(types.Position):
//				positions := tradeData.Data
//				for _, positionData := range positions {
//					if string(positionData.Symbol) == o.cfg.ExchangesSettings.Bitmex.Symbol {
//						position, err := o.toPosition(&positionData)
//						if err != nil {
//							o.log.Errorf("[bitmex exchange data]:%#v convert to position failed [err]:%v",
//								positionData, err)
//							continue
//						}
//						o.currentPosition = *position
//					}
//
//					unrealisedPnl := trademath.ConvertToBTC(o.currentPosition.UnrealisedPnl)
//					o.log.Debugf("current position [unrealised pnl in btc]: %.9f", unrealisedPnl)
//					isProfit := unrealisedPnl >= o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC  // if position profit
//					isLoss := unrealisedPnl <= -o.cfg.ExchangesSettings.Bitmex.ClosePositionMinBTC/2 // if position lesion
//					if isProfit || isLoss {
//						filter := fmt.Sprintf(`{"open": %t}`, true)
//						orders, err := o.api.GetBitmex().GetOrders(&bitmex.OrdersRequest{
//							Symbol: o.cfg.ExchangesSettings.Bitmex.Symbol,
//							Filter: filter,
//						})
//						if err != nil {
//							o.log.Errorf("o.api.GetBitmex().GetOrders failed: %v", err)
//							continue
//						}
//						if len(orders) > 0 {
//							o.log.Infof("proc %d orders in active, break process current position", len(orders))
//							break
//						}
//
//						var side types.Side
//						if o.currentPosition.CurrentQty > 0 {
//							side = types.SideSell
//						} else if o.currentPosition.CurrentQty < 0 {
//							side = types.SideBuy
//						} else {
//							o.log.Warnf("opening qty is 0")
//							break
//						}
//
//						order, err := o.PlaceOrder(types.Bitmex, side,
//							math.Abs(float64(o.currentPosition.CurrentQty)), false, true)
//						if err != nil {
//							o.log.Errorf("process position failed place %s order: %v", side, err)
//							continue
//						}
//						o.log.Infof("placed [order]:%#v", order)
//					}
//				}
//			default:
//				o.log.Debugf("table %s not processed", tradeData.Table)
//			}
//		}
//	}
//}
//
//func (o *OrderProcessor) procActiveOrders() error {
//	o.log.Infof("\n===============================\nstart process active orders")
//	defer o.log.Infof("finish process active orders\n===============================\n")
//
//	filter := fmt.Sprintf(`{"open": %t}`, true)
//	orders, err := o.api.GetBitmex().GetOrders(&bitmex.OrdersRequest{
//		Symbol: o.cfg.ExchangesSettings.Bitmex.Symbol,
//		Filter: filter,
//	})
//	if err != nil {
//		return err
//	}
//
//	for _, order := range orders {
//		if order.OrdType == string(types.LimitIfTouched) {
//			o.log.Debugf("this order is trailing stop, amend not needed, oid:%v", order.OrderID)
//			continue
//		}
//
//		inst, err := o.getInstrument()
//		if err != nil {
//			return err
//		}
//
//		var needUpdateOrd bool
//		switch types.Side(order.Side) {
//		case types.SideSell:
//			if inst.MarkPrice > order.StopPx {
//				needUpdateOrd = true
//			}
//		case types.SideBuy:
//			if inst.MarkPrice < order.StopPx {
//				needUpdateOrd = true
//			}
//		}
//
//		if !needUpdateOrd {
//			continue
//		}
//
//		cancelOrders, err := o.api.GetBitmex().CancelOrders(&bitmex.OrderCancelParams{
//			OrderID: order.OrderID,
//		})
//		if err != nil {
//			o.log.Errorf("procActiveOrders() cancel current order failed: %v", err)
//			return err
//		}
//		for _, cancelOrd := range cancelOrders {
//			o.log.Debugf("canceled order: %#v", cancelOrd)
//		}
//
//		newOrder, err := o.PlaceOrder(types.Bitmex, types.Side(order.Side),
//			math.Abs(float64(order.OrderQty)), false, true)
//		if err != nil {
//			o.log.Errorf("procActiveOrders() process active orders failed place %s order: %v", order.Side, err)
//			return err
//		}
//		o.log.Debugf("procActiveOrders() places new stop order: %#v", newOrder)
//	}
//
//	return nil
//}

func (o *OrderProcessor) PlaceOrder(
	exchange types.Exchange,
	side types.Side,
	amount float64,
	passive bool,
	isStop bool,
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
			err = o.checkLiquidation(&o.currentPosition, price, side)
			if err != nil {
				return nil, err
			}
			amount, err = o.calcOrderQty(
				&o.currentPosition,
				availableBalance,
				side,
			)
			if err != nil {
				return nil, err
			}
		}

		if !isStop {
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
		}

		order, err = o.placeTrailingStopOrder(
			"",
			side,
			math.Round(amount),
			price,
			passive)
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
	price float64,
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
func (o *OrderProcessor) checkLiquidation(position *bitmex.Position, price float64, side types.Side) error {
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

func (o *OrderProcessor) toPosition(d *data.BitmexIncomingData) (*bitmex.Position, error) {
	ts, err := time.Parse("2006-01-02T15:04:05.999Z", d.Timestamp)
	if err != nil {
		return nil, err
	}

	position := bitmex.Position{
		Account:              d.Account,
		AvgCostPrice:         d.AvgCostPrice,
		AvgEntryPrice:        d.AvgEntryPrice,
		BankruptPrice:        d.BankruptPrice,
		BreakEvenPrice:       d.BreakEvenPrice,
		Commission:           d.Commission,
		CrossMargin:          d.CrossMargin,
		Currency:             d.Currency,
		CurrentComm:          d.CurrentComm,
		CurrentCost:          d.CurrentCost,
		CurrentQty:           d.CurrentQty,
		CurrentTimestamp:     d.CurrentTimestamp,
		DeleveragePercentile: d.DeleveragePercentile,
		ExecBuyCost:          d.ExecBuyCost,
		ExecBuyQty:           d.ExecBuyQty,
		ExecComm:             d.ExecComm,
		ExecCost:             d.ExecCost,
		ExecQty:              d.ExecQty,
		ExecSellCost:         d.ExecSellCost,
		ExecSellQty:          d.ExecSellQty,
		ForeignNotional:      d.ForeignNotional,
		GrossExecCost:        d.GrossExecCost,
		GrossOpenCost:        d.GrossOpenCost,
		GrossOpenPremium:     d.GrossOpenPremium,
		HomeNotional:         d.HomeNotional,
		IndicativeTax:        d.IndicativeTax,
		IndicativeTaxRate:    d.IndicativeTaxRate,
		InitMargin:           d.InitMargin,
		InitMarginReq:        d.InitMarginReq,
		IsOpen:               d.IsOpen,
		LastPrice:            d.LastPrice,
		LastValue:            d.LastValue,
		Leverage:             d.Leverage,
		LiquidationPrice:     d.LiquidationPrice,
		LongBankrupt:         d.LongBankrupt,
		MaintMargin:          d.MaintMargin,
		MaintMarginReq:       d.MaintMarginReq,
		MarginCallPrice:      d.MarginCallPrice,
		MarkPrice:            d.MarkPrice,
		MarkValue:            d.MarkValue,
		OpenOrderBuyCost:     d.OpenOrderBuyCost,
		OpenOrderBuyPremium:  d.OpenOrderBuyPremium,
		OpenOrderBuyQty:      d.OpenOrderBuyQty,
		OpenOrderSellCost:    d.OpenOrderSellCost,
		OpenOrderSellPremium: d.OpenOrderSellPremium,
		OpenOrderSellQty:     d.OpenOrderSellQty,
		OpeningComm:          d.OpeningComm,
		OpeningCost:          d.OpeningCost,
		OpeningQty:           d.OpeningQty,
		OpeningTimestamp:     d.OpeningTimestamp,
		PosAllowance:         d.PosAllowance,
		PosComm:              d.PosComm,
		PosCost:              d.PosCost,
		PosCost2:             d.PosCost2,
		PosCross:             d.PosCross,
		PosInit:              d.PosInit,
		PosLoss:              d.PosLoss,
		PosMaint:             d.PosMaint,
		PosMargin:            d.PosMargin,
		PosState:             d.PosState,
		PrevClosePrice:       d.PrevClosePrice,
		PrevRealisedPnl:      d.PrevRealisedPnl,
		PrevUnrealisedPnl:    d.PrevUnrealisedPnl,
		QuoteCurrency:        d.QuoteCurrency,
		RealisedCost:         d.RealisedCost,
		RealisedGrossPnl:     d.RealisedGrossPnl,
		RealisedPnl:          d.RealisedPnl,
		RealisedTax:          d.RealisedTax,
		RebalancedPnl:        d.RebalancedPnl,
		RiskLimit:            d.RiskLimit,
		RiskValue:            d.RiskValue,
		SessionMargin:        d.SessionMargin,
		ShortBankrupt:        d.ShortBankrupt,
		SimpleCost:           d.SimpleCost,
		SimplePnl:            d.SimplePnl,
		SimplePnlPcnt:        d.SimplePnlPcnt,
		SimpleQty:            d.SimpleQty,
		SimpleValue:          d.SimpleValue,
		Symbol:               string(d.Symbol),
		TargetExcessMargin:   d.TargetExcessMargin,
		TaxBase:              d.TaxBase,
		TaxableMargin:        d.TaxableMargin,
		Timestamp:            ts,
		Underlying:           d.Underlying,
		UnrealisedCost:       d.UnrealisedCost,
		UnrealisedGrossPnl:   d.UnrealisedGrossPnl,
		UnrealisedPnl:        d.UnrealisedPnl,
		UnrealisedPnlPcnt:    d.UnrealisedPnlPcnt,
		UnrealisedRoePcnt:    d.UnrealisedRoePcnt,
		UnrealisedTax:        d.UnrealisedTax,
		VarMargin:            d.VarMargin,
	}

	return &position, nil
}
