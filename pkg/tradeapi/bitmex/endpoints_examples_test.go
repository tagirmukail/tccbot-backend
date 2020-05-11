// +build examples

package bitmex

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	// set environments by this keys for examples on your account
	keyEnv       = "BITMEX_KEY"
	secretKeyEnv = "BITMEX_SECRET_KEY"
)

func getEnvs(t *testing.T) map[string]string {
	bitmexKey := os.Getenv(keyEnv)
	require.NotEmpty(t, bitmexKey, keyEnv)
	bitmexSecret := os.Getenv(secretKeyEnv)
	require.NotEmpty(t, bitmexSecret, secretKeyEnv)
	return map[string]string{
		keyEnv:       bitmexKey,
		secretKeyEnv: bitmexSecret,
	}
}

func TestBitmex_GetAllUserMargin(t *testing.T) {
	asserter := require.New(t)
	envs := getEnvs(t)
	type fields struct {
		retryCount      int
		idleConnTimeout time.Duration
		maxIdleConns    int
		timeout         time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		want    []UserMargin
		wantErr bool
	}{
		{
			name: "get all user margins",
			fields: fields{
				retryCount:      2,
				idleConnTimeout: 15 * time.Second,
				maxIdleConns:    10,
				timeout:         0,
			},
			want:    []UserMargin{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(
				envs[keyEnv],
				envs[secretKeyEnv],
				false,
				tt.fields.retryCount,
				tt.fields.idleConnTimeout,
				tt.fields.maxIdleConns,
				0,
				tt.fields.timeout,
				logrus.New(),
			)
			b.EnableTestNet()

			got, err := b.GetAllUserMargin()
			asserter.NoError(err)
			asserter.NotEmpty(got)
			for _, margin := range got {
				asserter.True(margin.Account > 0)
				asserter.True(margin.Amount > 0)
				asserter.True(margin.AvailableMargin > 0)
				asserter.Equal("XBt", margin.Currency)
				asserter.True(margin.MarginBalance > 0)
				asserter.True(margin.WalletBalance > 0)
				t.Logf("%#v", margin)
			}
		})
	}
}

func TestBitmex_GetTradeBucketed(t *testing.T) {
	t.Run("get trade bucketed", func(t *testing.T) {
		b := New(
			"", "", true, 2, defaultIdleConnTimeout,
			10, 50, 0, logrus.New(),
		)

		b.EnableTestNet()

		//endTime := time.Now().UTC().Format(TradeTimeFormat)
		startTime := time.Now().UTC().Truncate(24 * time.Hour).Format(TradeTimeFormat)

		resp, err := b.GetTradeBucketed(&TradeGetBucketedParams{
			Symbol:    "XBTUSD",
			BinSize:   "5m",
			Count:     100,
			StartTime: startTime,
			//EndTime:   endTime,
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp)
		for _, trade := range resp {
			t.Logf("%#v", trade)
		}
	})
}

func TestBitmex_GetPositions(t *testing.T) {
	asserter := require.New(t)
	envs := getEnvs(t)
	type fields struct {
		retryCount      int
		idleConnTimeout time.Duration
		maxIdleConns    int
		timeout         time.Duration
	}

	type args struct {
		params PositionGetParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Position
		wantErr bool
	}{
		{
			name: "get positions",
			fields: fields{
				retryCount:      2,
				idleConnTimeout: 15 * time.Second,
				maxIdleConns:    10,
				timeout:         0,
			},
			args: args{
				params: PositionGetParams{
					Filter: `{"symbol": "XBTUSD"}`,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(
				envs[keyEnv],
				envs[secretKeyEnv],
				false,
				tt.fields.retryCount,
				tt.fields.idleConnTimeout,
				tt.fields.maxIdleConns,
				0,
				tt.fields.timeout,
				logrus.New(),
			)
			b.EnableTestNet()

			got, err := b.GetPositions(tt.args.params)
			asserter.NoError(err)
			asserter.NotEmpty(got)
			for _, pos := range got {
				t.Logf("%#v", pos)
			}
		})
	}
}

func TestBitmex_CreateOrder(t *testing.T) {
	asserter := require.New(t)
	envs := getEnvs(t)

	type fields struct {
		retryCount      int
		idleConnTimeout time.Duration
		maxIdleConns    int
		timeout         time.Duration
	}
	type args struct {
		params *OrderNewParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    OrderCopied
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				retryCount:      2,
				idleConnTimeout: 15 * time.Second,
				maxIdleConns:    10,
				timeout:         0,
			},
			args: args{
				params: &OrderNewParams{
					Symbol:    "XBTUSD",
					Side:      "Sell",
					OrderType: "Limit",
					OrderQty:  math.Round(10.5434322312),
					Price:     8413.5,
				},
			},
			want:    OrderCopied{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(
				envs[keyEnv],
				envs[secretKeyEnv],
				false,
				tt.fields.retryCount,
				tt.fields.idleConnTimeout,
				tt.fields.maxIdleConns,
				0,
				tt.fields.timeout,
				logrus.New(),
			)
			b.EnableTestNet()

			got, err := b.CreateOrder(tt.args.params)
			asserter.NoError(err)
			asserter.NotEmpty(got)
			t.Logf("%#v", got)
		})
	}
}
