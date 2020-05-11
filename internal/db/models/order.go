package models

import (
	"time"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type Order struct {
	ID                    int64     `db:"id"`
	Account               int64     `json:"account"`
	AvgPx                 float64   `json:"avgPx"`
	ClOrdID               string    `json:"clOrdID"`
	ClOrdLinkID           string    `json:"clOrdLinkID"`
	ContingencyType       string    `json:"contingencyType"`
	CumQty                int64     `json:"cumQty"`
	Currency              string    `json:"currency"`
	DisplayQuantity       int64     `json:"displayQty"`
	ExDestination         string    `json:"exDestination"`
	ExecInst              string    `json:"execInst"`
	LeavesQty             int64     `json:"leavesQty"`
	MultiLegReportingType string    `json:"multiLegReportingType"`
	OrdRejReason          string    `json:"ordRejReason"`
	OrdStatus             string    `json:"ordStatus"`
	OrdType               string    `json:"ordType"`
	OrderID               string    `json:"orderID"`
	OrderQty              int64     `json:"orderQty"`
	PegOffsetValue        float64   `json:"pegOffsetValue"`
	PegPriceType          string    `json:"pegPriceType"`
	Price                 float64   `json:"price"`
	SettlCurrency         string    `json:"settlCurrency"`
	Side                  string    `json:"side"`
	SimpleCumQty          float64   `json:"simpleCumQty"`
	SimpleLeavesQty       float64   `json:"simpleLeavesQty"`
	SimpleOrderQty        float64   `json:"simpleOrderQty"`
	StopPx                float64   `json:"stopPx"`
	Symbol                string    `json:"symbol"`
	Text                  string    `json:"text"`
	TimeInForce           string    `json:"timeInForce"`
	Timestamp             time.Time `json:"timestamp"`
	TransactTime          string    `json:"transactTime"`
	Triggered             string    `json:"triggered"`
	WorkingIndicator      bool      `json:"workingIndicator"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

func ToOrderModel(orderC bitmex.OrderCopied) *Order {
	return &Order{
		Account:               orderC.Account,
		AvgPx:                 orderC.AvgPx,
		ClOrdID:               orderC.ClOrdID,
		ClOrdLinkID:           orderC.ClOrdLinkID,
		ContingencyType:       orderC.ContingencyType,
		CumQty:                orderC.CumQty,
		Currency:              orderC.Currency,
		DisplayQuantity:       orderC.DisplayQuantity,
		ExDestination:         orderC.ExDestination,
		ExecInst:              orderC.ExecInst,
		LeavesQty:             orderC.LeavesQty,
		MultiLegReportingType: orderC.MultiLegReportingType,
		OrdRejReason:          orderC.OrdRejReason,
		OrdStatus:             orderC.OrdStatus,
		OrdType:               orderC.OrdType,
		OrderID:               orderC.OrderID,
		OrderQty:              orderC.OrderQty,
		PegOffsetValue:        orderC.PegOffsetValue,
		PegPriceType:          orderC.PegPriceType,
		Price:                 orderC.Price,
		SettlCurrency:         orderC.SettlCurrency,
		Side:                  orderC.Side,
		SimpleCumQty:          orderC.SimpleCumQty,
		SimpleLeavesQty:       orderC.SimpleLeavesQty,
		SimpleOrderQty:        orderC.SimpleOrderQty,
		StopPx:                orderC.StopPx,
		Symbol:                orderC.Symbol,
		Text:                  orderC.Text,
		TimeInForce:           orderC.TimeInForce,
		Timestamp:             orderC.Timestamp,
		TransactTime:          orderC.TransactTime,
		Triggered:             orderC.Triggered,
		WorkingIndicator:      orderC.WorkingIndicator,
	}
}
