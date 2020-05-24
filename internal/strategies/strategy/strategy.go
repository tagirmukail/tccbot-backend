package strategy

import (
	"context"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

type Strategy interface {
	Execute(ctx context.Context, size models.BinSize) error
}
