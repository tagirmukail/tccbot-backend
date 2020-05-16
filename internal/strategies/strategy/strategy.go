package strategy

import (
	"context"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

type AcceptAction uint8

const (
	NotAccepted AcceptAction = iota
	UpAccepted
	DownAccepted
)

type Strategy interface {
	Execute(ctx context.Context, size models.BinSize) error
}
