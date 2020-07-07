package filter

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	stratypes "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

func Test_checkCandlesTrend_check(t *testing.T) {
	type fields struct {
		currentMaxClosePrice float64
		candlesCurrentCount  uint32
		isUp                 bool
		isDown               bool
		side                 types.Side
		log                  *logrus.Logger
	}
	type args struct {
		ctxData *getFromCtxData
		cfg     *config.StrategiesConfig
	}

	type test struct {
		fields          fields
		args            args
		want            types.Side
		wantIsUp        bool
		wantIsDown      bool
		wantCounterIncr bool
		wantClosePrice  float64
	}

	t.Run("up trend incoming", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 0,
				candlesCurrentCount:  0,
				isDown:               false,
				isUp:                 false,
				log:                  logrus.New(),
				side:                 types.SideEmpty,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.UpTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 300},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:            types.SideEmpty,
			wantIsUp:        true,
			wantCounterIncr: true,
			wantClosePrice:  300,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, tt.wantCounterIncr && c.candlesCurrentCount > 0, "counter not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
	})

	t.Run("up trend active, last candle close price up", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 300,
				candlesCurrentCount:  1,
				isDown:               false,
				isUp:                 true,
				log:                  logrus.New(),
				side:                 types.SideSell,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.NotTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 300}, {Close: 450},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:            types.SideEmpty,
			wantIsUp:        true,
			wantCounterIncr: true,
			wantClosePrice:  450,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, tt.wantCounterIncr && c.candlesCurrentCount > 0, "counter not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
	})

	t.Run("up trend active, max candles filter count overcome", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 300,
				candlesCurrentCount:  3,
				isDown:               false,
				isUp:                 true,
				log:                  logrus.New(),
				side:                 types.SideSell,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.NotTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 300}, {Close: 250},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:           types.SideSell,
			wantIsUp:       false,
			wantClosePrice: 300,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, c.candlesCurrentCount == 0, "counter not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
	})

	t.Run("down trend incoming", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 0,
				candlesCurrentCount:  0,
				isDown:               false,
				isUp:                 false,
				log:                  logrus.New(),
				side:                 types.SideEmpty,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.DownTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 50},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:            types.SideEmpty,
			wantIsDown:      true,
			wantCounterIncr: true,
			wantClosePrice:  50,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, tt.wantCounterIncr && c.candlesCurrentCount > 0, "counter not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
		require.Equal(t, types.SideBuy, c.side)
	})

	t.Run("down trend, incoming down close price", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 50,
				candlesCurrentCount:  2,
				isDown:               true,
				isUp:                 false,
				log:                  logrus.New(),
				side:                 types.SideBuy,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.NotTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 50}, {Close: 30},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:            types.SideEmpty,
			wantIsDown:      true,
			wantCounterIncr: true,
			wantClosePrice:  30,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, tt.wantCounterIncr && c.candlesCurrentCount > 0, "counter not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
		require.Equal(t, types.SideBuy, c.side)
	})

	t.Run("down trend, max filter count overcome", func(t *testing.T) {

		tt := test{
			fields: fields{
				currentMaxClosePrice: 30,
				candlesCurrentCount:  3,
				isDown:               true,
				isUp:                 false,
				log:                  logrus.New(),
				side:                 types.SideBuy,
			},
			args: args{
				ctxData: &getFromCtxData{
					action:  stratypes.NotTrend,
					binSize: models.Bin5m,
					candles: []bitmex.TradeBuck{
						{Close: 100}, {Close: 200}, {Close: 30}, {Close: 35},
					},
				},
				cfg: &config.StrategiesConfig{
					MaxCandlesFilterCount: 4,
				},
			},
			want:           types.SideBuy,
			wantIsDown:     false,
			wantClosePrice: 30,
		}

		c := &checkCandlesTrend{
			currentMaxClosePrice: tt.fields.currentMaxClosePrice,
			candlesCurrentCount:  tt.fields.candlesCurrentCount,
			isUp:                 tt.fields.isUp,
			isDown:               tt.fields.isDown,
			side:                 tt.fields.side,
			log:                  tt.fields.log,
		}
		got := c.check(tt.args.ctxData, tt.args.cfg)
		require.Equal(t, tt.want, got)
		require.True(t, tt.wantIsUp == c.isUp, "up trend failed")
		require.True(t, tt.wantIsDown == c.isDown, "down trend failed")
		require.True(t, c.candlesCurrentCount == 0, "counter should be not incremented")
		require.Equal(t, tt.wantClosePrice, c.currentMaxClosePrice)
		require.Equal(t, types.SideEmpty, c.side)
	})
}
