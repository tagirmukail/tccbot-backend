package candlecache

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

func TestCandleCache_StoreBatch(t *testing.T) {
	type fields struct {
		maxCount int
		symbol   types.Symbol
		log      *logrus.Logger
	}
	type args struct {
		batch []bitmex.TradeBuck
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantStore []bitmex.TradeBuck
	}{
		{
			name: "ok",
			fields: fields{
				maxCount: 5,
				symbol:   types.XBTUSD,
				log:      logrus.New(),
			},
			args: args{
				batch: []bitmex.TradeBuck{
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:10:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:15:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
				},
			},
			wantStore: []bitmex.TradeBuck{
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:35:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:40:00.000Z",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCandleCache(tt.fields.maxCount, tt.fields.symbol, tt.fields.log)
			c.StoreBatch(tt.args.batch)
			require.Equal(t, tt.fields.maxCount, len(c.store), "batch not stored")
			require.Equal(t, tt.wantStore, c.store)
		})
	}
}

func TestCandleCache_Store(t *testing.T) {
	type fields struct {
		maxCount int
		store    []bitmex.TradeBuck
		symbol   types.Symbol
		log      *logrus.Logger
	}
	type args struct {
		candle data.BitmexIncomingData
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantStore []bitmex.TradeBuck
	}{
		{
			name: "ok",
			fields: fields{
				maxCount: 7,
				store: []bitmex.TradeBuck{
					{
						Close:     456,
						Symbol:    string(types.XBTUSD),
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
				},
				symbol: types.XBTUSD,
				log:    nil,
			},
			args: args{
				candle: data.BitmexIncomingData{
					Symbol: types.XBTUSD,
					TradeBinData: data.TradeBinData{
						Close: 456,
					},
					Timestamp: "2020-05-16T10:45:00.000Z",
				},
			},
			wantErr: false,
			wantStore: []bitmex.TradeBuck{
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:35:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:40:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:45:00.000Z",
				},
			},
		},
		{
			name: "error",
			fields: fields{
				maxCount: 5,
				store:    []bitmex.TradeBuck{},
				symbol:   types.XBTUSD,
				log:      nil,
			},
			args: args{
				candle: data.BitmexIncomingData{
					Symbol: types.Symbol("BTC"),
					TradeBinData: data.TradeBinData{
						Close: 456,
					},
					Timestamp: "2020-05-16T10:45:00.000Z",
				},
			},
			wantErr:   true,
			wantStore: []bitmex.TradeBuck{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCandleCache(tt.fields.maxCount, tt.fields.symbol, tt.fields.log)
			c.store = tt.fields.store
			err := c.Store(tt.args.candle)
			require.Equal(t, tt.wantErr, err != nil, "error", err)
			require.Equal(t, tt.wantStore, c.store)
		})
	}
}

func TestCandleCache_GetBucketed(t *testing.T) {
	type fields struct {
		store    []bitmex.TradeBuck
		maxCount int
		symbol   types.Symbol
		log      *logrus.Logger
	}
	type args struct {
		from  time.Time
		to    time.Time
		count int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []bitmex.TradeBuck
	}{
		{
			name: "with all parameters",
			fields: fields{
				store: []bitmex.TradeBuck{
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:10:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:15:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
				},
				maxCount: 5,
				symbol:   types.XBTUSD,
				log:      logrus.New(),
			},
			args: args{
				from:  time.Date(2020, 05, 16, 10, 10, 0, 0, time.UTC),
				to:    time.Date(2020, 05, 16, 11, 0, 0, 0, time.UTC),
				count: 5,
			},
			want: []bitmex.TradeBuck{
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:35:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:40:00.000Z",
				},
			},
		},
		{
			name: "only with from",
			fields: fields{
				store: []bitmex.TradeBuck{
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:10:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:15:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
				},
				maxCount: 5,
				symbol:   types.XBTUSD,
				log:      logrus.New(),
			},
			args: args{
				from: time.Date(2020, 05, 16, 10, 10, 0, 0, time.UTC),
			},
			want: []bitmex.TradeBuck{
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:10:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:15:00.000Z",
				},
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:35:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:40:00.000Z",
				},
			},
		},
		{
			name: "only with to",
			fields: fields{
				store: []bitmex.TradeBuck{
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:10:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:15:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
				},
				maxCount: 5,
				symbol:   types.XBTUSD,
				log:      logrus.New(),
			},
			args: args{
				to: time.Date(2020, 05, 16, 10, 30, 0, 0, time.UTC),
			},
			want: []bitmex.TradeBuck{
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:10:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:15:00.000Z",
				},
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
			},
		},
		{
			name: "with count parameter",
			fields: fields{
				store: []bitmex.TradeBuck{
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:10:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:35:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:15:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:40:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:30:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:20:00.000Z",
					},
					{
						Symbol:    string(types.XBTUSD),
						Close:     456,
						Timestamp: "2020-05-16T10:25:00.000Z",
					},
				},
				maxCount: 5,
				symbol:   types.XBTUSD,
				log:      logrus.New(),
			},
			args: args{
				count: 5,
			},
			want: []bitmex.TradeBuck{
				{
					Close:     456,
					Symbol:    string(types.XBTUSD),
					Timestamp: "2020-05-16T10:20:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:25:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:30:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:35:00.000Z",
				},
				{
					Symbol:    string(types.XBTUSD),
					Close:     456,
					Timestamp: "2020-05-16T10:40:00.000Z",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCandleCache(tt.fields.maxCount, tt.fields.symbol, tt.fields.log)
			c.store = tt.fields.store
			got := c.GetBucketed(tt.args.from, tt.args.to, tt.args.count)
			require.Equal(t, len(tt.want), len(got))
			require.Equal(t, tt.want, got)
		})
	}
}
