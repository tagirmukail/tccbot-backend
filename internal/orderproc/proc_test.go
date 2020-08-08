package orderproc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func TestOrderProcessor_calcOrderQty(t *testing.T) {
	type fields struct {
		cfg      *config.GlobalConfig
		position *bitmex.Position
	}
	type args struct {
		balance float64
		side    types.Side
	}
	type test struct {
		fields         fields
		args           args
		wantQtyContrts float64
		wantErr        error
	}

	t.Run("pnl > 0 and curency qty > 0", func(t *testing.T) {
		tt := test{
			fields: fields{
				cfg: &config.GlobalConfig{
					ExchangesSettings: config.ExchangesSettings{
						Bitmex: config.ApiSettings{
							SellOrderCoef: 0.1,
							BuyOrderCoef:  0.2,
						},
					},
				},
				position: &bitmex.Position{
					CurrentQty:    345,
					UnrealisedPnl: 163,
				},
			},
			args: args{
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 345,
		}

		o := &OrderProcessor{
			currentPosition: tt.fields.position,
			cfg:             tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})

	t.Run("pnl < 0 and currency qty == 0", func(t *testing.T) {
		tt := test{
			fields: fields{
				cfg: &config.GlobalConfig{
					ExchangesSettings: config.ExchangesSettings{
						Bitmex: config.ApiSettings{
							SellOrderCoef: 0.1,
							BuyOrderCoef:  0.15,
						},
					},
				},
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: -100,
				},
			},
			args: args{
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 135000,
		}

		o := &OrderProcessor{
			currentPosition: tt.fields.position,
			cfg:             tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})

	t.Run("pnl == 0 and currency qty == 0", func(t *testing.T) {
		tt := test{
			fields: fields{
				cfg: &config.GlobalConfig{
					ExchangesSettings: config.ExchangesSettings{
						Bitmex: config.ApiSettings{
							SellOrderCoef: 0.1,
							BuyOrderCoef:  0.15,
						},
					},
				},
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
			},
			args: args{
				balance: 0.0090,
				side:    types.SideSell,
			},
			wantQtyContrts: 90000,
		}

		o := &OrderProcessor{
			cfg:             tt.fields.cfg,
			currentPosition: tt.fields.position,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})

	t.Run("side empty", func(t *testing.T) {
		tt := test{
			fields: fields{
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
				cfg: &config.GlobalConfig{
					ExchangesSettings: config.ExchangesSettings{
						Bitmex: config.ApiSettings{
							SellOrderCoef: 0.1,
							BuyOrderCoef:  0.15,
						},
					},
				},
			},
			args: args{
				balance: 0.0090,
				side:    types.SideEmpty,
			},
			wantQtyContrts: 0,
			wantErr:        fmt.Errorf("unknown side type: %s", types.SideEmpty),
		}

		o := &OrderProcessor{
			currentPosition: tt.fields.position,
			cfg:             tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.balance, tt.args.side)
		require.EqualError(t, err, tt.wantErr.Error())
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})

	t.Run("pnl == 0 and currency qty == 0", func(t *testing.T) {
		tt := test{
			fields: fields{
				cfg: &config.GlobalConfig{
					ExchangesSettings: config.ExchangesSettings{
						Bitmex: config.ApiSettings{
							SellOrderCoef: 0.1,
							BuyOrderCoef:  0.01,
						},
					},
				},
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
			},
			args: args{
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 9000,
		}

		o := &OrderProcessor{
			cfg:             tt.fields.cfg,
			currentPosition: tt.fields.position,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})
}
