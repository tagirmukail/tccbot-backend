package filter

import (
	"context"
	"errors"
	"fmt"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	stratypes "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const maxPrev = 4

type Filter interface {
	Apply(ctx context.Context) types.Side
}

type getFromCtxData struct {
	action  stratypes.Action
	binSize models.BinSize
	candles []bitmex.TradeBuck
}

func getFromCtx(ctx context.Context) (*getFromCtxData, error) {
	act := ctx.Value(stratypes.ActionKey)
	if act == nil {
		return nil, errors.New("getFromCtx - context action is <nil>")
	}
	action, ok := act.(stratypes.Action)
	if !ok {
		return nil, errors.New("getFromCtx - context action type is not <Action>")
	}
	if err := action.Validate(); err != nil {
		return nil, fmt.Errorf("action validate error:%v", err)
	}

	binS := ctx.Value(stratypes.BinSizeKey)
	if act == nil {
		return nil, errors.New("getFromCtx - context bin size is <nil>")
	}
	binSize, ok := binS.(models.BinSize)
	if !ok {
		return nil, errors.New("getFromCtx - context key type is not <CtxKey>")
	}

	iCandles := ctx.Value(stratypes.CandlesKey)
	if iCandles == nil {
		return nil, errors.New("getFromCtx - context candles is <nil>")
	}

	candles, ok := iCandles.([]bitmex.TradeBuck)
	if !ok {
		return nil, errors.New("getFromCtx - context candles type is not <[]bitmex.TradeBuck>")
	}

	return &getFromCtxData{
		action:  action,
		binSize: binSize,
		candles: candles,
	}, nil
}
