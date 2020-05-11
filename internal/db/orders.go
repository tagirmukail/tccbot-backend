package db

import (
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func (db *DB) SaveOrder(ord bitmex.OrderCopied) (*models.Order, error) {
	order := models.ToOrderModel(ord)
	err := db.db.QueryRow(`INSERT INTO orders(
                   account,
					avgPx,
					clOrdID,
					clOrdLinkID,
					contingencyType,
					cumQty,
					currency,
					displayQty,
					exDestination,
					execInst,
					leavesQty,
					multiLegReportingType,
					ordRejReason,
					ordStatus,
					ordType,
					orderID,
					orderQty,
					pegOffsetValue,
					pegPriceType,
					price,
					settlCurrency,
					side,
					simpleCumQty,
					simpleLeavesQty,
					simpleOrderQty,
					stopPx,
					symbol,
					text,
					timeInForce,
					timestamp,
					transactTime,
					triggered,
					workingIndicator,
    				created_at
				) VALUES (
					$1, $2,$3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
					$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
				    $27, $28, $29, $30, $31, $32, $33, $34, $35
				) RETURNING id`,
		ord.Account,
		ord.AvgPx,
		ord.ClOrdID,
		ord.ClOrdLinkID,
		ord.ContingencyType,
		ord.CumQty,
		ord.Currency,
		ord.DisplayQuantity,
		ord.ExDestination,
		ord.ExecInst,
		ord.LeavesQty,
		ord.MultiLegReportingType,
		ord.OrdRejReason,
		ord.OrdStatus,
		ord.OrdType,
		ord.OrderID,
		ord.OrderQty,
		ord.PegOffsetValue,
		ord.PegPriceType,
		ord.Price,
		ord.SettlCurrency,
		ord.Side,
		ord.SimpleCumQty,
		ord.SimpleLeavesQty,
		ord.SimpleOrderQty,
		ord.StopPx,
		ord.Symbol,
		ord.Text,
		ord.TimeInForce,
		ord.Timestamp,
		ord.TransactTime,
		ord.Triggered,
		ord.WorkingIndicator,
		time.Now().UTC(),
	).Scan(&order.ID)
	return order, err
}
