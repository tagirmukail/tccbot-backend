package bitmex

import "net/http"

func (b *Bitmex) GetUserMargin(currency string) (UserMargin, error) {
	var margin UserMargin
	return margin, b.SendAuthenticatedRequest(
		http.MethodGet,
		endpointUserMargin,
		map[string]interface{}{
			"currency": currency,
		},
		&margin,
	)
}

func (b *Bitmex) GetAllUserMargin() ([]UserMargin, error) {
	var margin []UserMargin
	return margin, b.SendAuthenticatedRequest(
		http.MethodGet,
		endpointUserMargin,
		map[string]interface{}{
			"currency": "all",
		},
		&margin,
	)
}

func (b *Bitmex) GetUserWalletInfo(currency string) (WalletInfo, error) {
	var info WalletInfo
	return info, b.SendAuthenticatedRequest(
		http.MethodGet,
		endpointUserWallet,
		map[string]interface{}{
			"currency": currency,
		},
		&info,
	)
}

func (b *Bitmex) GetOrders(params *OrdersRequest) ([]Order, error) {
	var orders []Order
	return orders, b.SendAuthenticatedRequest(
		http.MethodGet,
		endpointOrder,
		params,
		&orders,
	)
}

func (b *Bitmex) CreateOrder(params *OrderNewParams) (OrderCopied, error) {
	var order OrderCopied
	return order, b.SendAuthenticatedRequest(
		http.MethodPost,
		endpointOrder,
		params,
		&order,
	)
}

// AmendOrder amends the quantity or price of an open order
func (b *Bitmex) AmendOrder(params *OrderAmendParams) (Order, error) {
	var order Order
	return order, b.SendAuthenticatedRequest(
		http.MethodPut,
		endpointOrder,
		params,
		&order,
	)
}

func (b *Bitmex) CancelOrders(params *OrderCancelParams) ([]Order, error) {
	var orders []Order
	return orders, b.SendAuthenticatedRequest(
		http.MethodDelete,
		endpointOrder,
		params,
		&orders,
	)
}

func (b *Bitmex) CancelAllOrders(params *OrderCancelAllParams) ([]Order, error) {
	var orders []Order
	return orders, b.SendAuthenticatedRequest(
		http.MethodDelete,
		endpointAllOrders,
		params,
		&orders,
	)
}

func (b *Bitmex) GetTradeBucketed(params *TradeGetBucketedParams) ([]TradeBuck, error) {
	var resp []TradeBuck
	vals, err := params.toUrlVals()
	if err != nil {
		return nil, err
	}
	return resp, b.SendRequest(
		endpointTradeBucketed,
		*vals,
		&resp,
	)
}
