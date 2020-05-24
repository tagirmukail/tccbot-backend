package filter

import (
	"context"

	"github.com/tagirmukail/tccbot-backend/internal/types"
)

const maxPrev = 4

type Filter interface {
	Apply(ctx context.Context) types.Side
}
