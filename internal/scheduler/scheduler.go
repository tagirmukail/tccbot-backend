package scheduler

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

type Scheduler interface {
	Start(wg *sync.WaitGroup)
	Stop() error
}

func getActiveOrders(api tradeapi.Api, symbol string) ([]bitmex.OrderCopied, error) {
	filter := fmt.Sprintf(`{"open": %t}`, true)
	return api.GetBitmex().GetOrders(&bitmex.OrdersRequest{
		Symbol: symbol,
		Filter: filter,
	})
}

func getInstrument(api tradeapi.Api, symbol string) (bitmex.Instrument, error) {
	var resp bitmex.Instrument
	insts, err := api.GetBitmex().GetInstrument(bitmex.InstrumentRequestParams{
		Symbol:  symbol,
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

func FromBitmexIncDataToPosition(d *data.BitmexIncomingData) (*bitmex.Position, error) {
	ts, err := time.Parse("2006-01-02T15:04:05.999Z", d.Timestamp)
	if err != nil {
		return nil, err
	}

	var unrealPnl int64
	if d.UnrealisedPnl != nil {
		unrealPnl = *d.UnrealisedPnl
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
		UnrealisedPnl:        unrealPnl,
		UnrealisedPnlPcnt:    d.UnrealisedPnlPcnt,
		UnrealisedRoePcnt:    d.UnrealisedRoePcnt,
		UnrealisedTax:        d.UnrealisedTax,
		VarMargin:            d.VarMargin,
	}

	return &position, nil
}
