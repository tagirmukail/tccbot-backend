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

func (b *Bitmex) GetOrders(params *OrdersRequest) ([]OrderCopied, error) {
	var orders []OrderCopied
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
func (b *Bitmex) AmendOrder(params *OrderAmendParams) (OrderCopied, error) {
	var order OrderCopied
	return order, b.SendAuthenticatedRequest(
		http.MethodPut,
		endpointOrder,
		params,
		&order,
	)
}

func (b *Bitmex) CancelOrders(params *OrderCancelParams) ([]OrderCopied, error) {
	var orders []OrderCopied
	return orders, b.SendAuthenticatedRequest(
		http.MethodDelete,
		endpointOrder,
		params,
		&orders,
	)
}

func (b *Bitmex) CancelAllOrders(params *OrderCancelAllParams) ([]OrderCopied, error) {
	var orders []OrderCopied
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

func (b *Bitmex) LeveragePosition(params *PositionUpdateLeverageParams) (Position, error) {
	var resp Position
	return resp, b.SendAuthenticatedRequest(
		http.MethodPost,
		endpointLeveragePosition,
		params,
		&resp,
	)
}

// GetPositions returns positions
func (b *Bitmex) GetPositions(params PositionGetParams) ([]Position, error) {
	var positions []Position

	return positions, b.SendAuthenticatedRequest(
		http.MethodGet,
		endpointPosition,
		params,
		&positions,
	)
}

func (b *Bitmex) GetInstrument(params InstrumentRequestParams) ([]Instrument, error) {
	var resp []Instrument
	return resp, b.SendRequest(
		endpointInstrument,
		params.toUrlVars(),
		&resp,
	)
}
