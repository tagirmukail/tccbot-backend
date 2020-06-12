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
		cfg *config.GlobalConfig
	}
	type args struct {
		position *bitmex.Position
		balance  float64
		side     types.Side
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
			},
			args: args{
				position: &bitmex.Position{
					OpeningQty:    345,
					UnrealisedPnl: 163,
				},
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 345,
		}

		o := &OrderProcessor{
			cfg: tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.position, tt.args.balance, tt.args.side)
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
			},
			args: args{
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: -100,
				},
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 135000,
		}

		o := &OrderProcessor{
			cfg: tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.position, tt.args.balance, tt.args.side)
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
			},
			args: args{
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
				balance: 0.0090,
				side:    types.SideSell,
			},
			wantQtyContrts: 90000,
		}

		o := &OrderProcessor{
			cfg: tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.position, tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})

	t.Run("side empty", func(t *testing.T) {
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
			},
			args: args{
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
				balance: 0.0090,
				side:    types.SideEmpty,
			},
			wantQtyContrts: 0,
			wantErr:        fmt.Errorf("unknown side type: %s", types.SideEmpty),
		}

		o := &OrderProcessor{
			cfg: tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.position, tt.args.balance, tt.args.side)
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
			},
			args: args{
				position: &bitmex.Position{
					OpeningQty:    0,
					UnrealisedPnl: 0,
				},
				balance: 0.0090,
				side:    types.SideBuy,
			},
			wantQtyContrts: 9000,
		}

		o := &OrderProcessor{
			cfg: tt.fields.cfg,
		}
		gotQtyContrts, err := o.calcOrderQty(tt.args.position, tt.args.balance, tt.args.side)
		require.NoError(t, err)
		require.Equal(t, tt.wantQtyContrts, gotQtyContrts)
	})
}
