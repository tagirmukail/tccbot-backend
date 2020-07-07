package filter

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	types2 "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func TestTrendFilter_Apply(t *testing.T) {
	type fields struct {
		cfg         *config.GlobalConfig
		log         *logrus.Logger
		prevActions []types2.Action
	}
	type args struct {
		ctx context.Context
	}
	t.Run("Buy - down trend accepted", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log: logrus.New(),
				prevActions: []types2.Action{
					types2.DownTrend,
					types2.NotTrend,
					types2.UpTrend,
					types2.NotTrend,
					types2.NotTrend,
				},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.NotTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideBuy,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Empty(t, f.prevActions)
	})

	t.Run("Sell - up trend is accepted", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log: logrus.New(),
				prevActions: []types2.Action{
					types2.UpTrend,
					types2.DownTrend,
					types2.NotTrend,
					types2.NotTrend,
					types2.NotTrend,
				},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.NotTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideSell,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Empty(t, f.prevActions)
	})

	t.Run("Empty - trend is up continue, only add action in to prev actions", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log: logrus.New(),
				prevActions: []types2.Action{
					types2.UpTrend,
					types2.DownTrend,
					types2.NotTrend,
					types2.NotTrend,
				},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.UpTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Equal(t, []types2.Action{
			types2.UpTrend,
			types2.DownTrend,
			types2.NotTrend,
			types2.NotTrend,
			types2.UpTrend,
		}, f.prevActions)
	})

	t.Run("Empty - trend is down continue, only add action in to prev actions", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log: logrus.New(),
				prevActions: []types2.Action{
					types2.DownTrend,
					types2.DownTrend,
					types2.NotTrend,
					types2.DownTrend,
				},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.DownTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Equal(t, []types2.Action{
			types2.DownTrend,
			types2.DownTrend,
			types2.NotTrend,
			types2.DownTrend,
			types2.DownTrend,
		}, f.prevActions)
	})

	t.Run("Empty - not up and not down trend", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log: logrus.New(),
				prevActions: []types2.Action{
					types2.NotTrend,
					types2.NotTrend,
					types2.NotTrend,
					types2.NotTrend,
					types2.NotTrend,
				},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.NotTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Equal(t, []types2.Action{
			types2.NotTrend,
			types2.NotTrend,
			types2.NotTrend,
			types2.NotTrend,
			types2.NotTrend,
			types2.NotTrend,
		}, f.prevActions)
	})

	t.Run("Empty - actions too little", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log:         logrus.New(),
				prevActions: []types2.Action{},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.WithValue(context.Background(), types2.ActionKey, types2.NotTrend),
					types2.CandlesKey, []bitmex.TradeBuck{}), types2.BinSizeKey, models.Bin5m),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Equal(t, []types2.Action{
			types2.NotTrend,
		}, f.prevActions)
	})

	t.Run("Empty - ctx action is nil", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log:         logrus.New(),
				prevActions: []types2.Action{},
			},
			args: args{
				ctx: context.WithValue(context.Background(), "candles", []bitmex.TradeBuck{}),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Empty(t, f.prevActions)
	})

	t.Run("Empty - ctx action type is not Action", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log:         logrus.New(),
				prevActions: []types2.Action{},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), "action", "NotAction"),
					"candles", []bitmex.TradeBuck{}),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Empty(t, f.prevActions)
	})

	t.Run("Empty - ctx action invalid", func(t *testing.T) {
		tt := struct {
			fields fields
			args   args
			want   types.Side
		}{
			fields: fields{
				cfg: &config.GlobalConfig{
					GlobStrategies: config.StrategiesGlobConfig{
						M5: &config.StrategiesConfig{
							MaxFilterTrendCount: 6,
						},
					},
				},
				log:         logrus.New(),
				prevActions: []types2.Action{},
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), "action", types2.Action(6)),
					"candles", []bitmex.TradeBuck{}),
			},
			want: types.SideEmpty,
		}
		f := NewTrendFilter(tt.fields.cfg, tt.fields.log)
		f.prevActions = tt.fields.prevActions
		got := f.Apply(tt.args.ctx)
		require.Equal(t, tt.want, got, "apply side not equal")
		require.Empty(t, f.prevActions)
	})
}
