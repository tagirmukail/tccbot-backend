package trademath

import (
	"math/rand"
	"testing"
	"time"

	"github.com/markcheno/go-talib"

	"github.com/stretchr/testify/assert"
)

func TestCalc_CalculateSignals(t *testing.T) {
	assertr := assert.New(t)
	type args struct {
		values []float64
	}
	tests := []struct {
		name string
		args args
		want Signals
	}{
		{
			name: "ok",
			args: args{
				values: []float64{
					1100,
					1115,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
					1100,
					1115,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
				},
			},
			want: Signals{
				SMA: 1112,
				WMA: 1112.5714,
				EMA: 1112,
				BB: struct {
					TL float64
					ML float64
					BL float64
				}{1127.3101, 1112, 1096.69},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalcSignals(tt.args.values, talib.SMA)
			assertr.Equal(tt.want, got)
		})
	}
}

func BenchmarkCalc_CalculateSignals(b *testing.B) {
	type args struct {
		values []float64
	}
	tests := []struct {
		name string
		args args
		want Signals
	}{
		{
			name: "10 values",
			args: args{
				values: []float64{
					1100,
					1115,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
				},
			},
			want: Signals{
				SMA: 1112,
				WMA: 1113.0909,
				EMA: 1113.3636,
				BB: struct {
					TL float64
					ML float64
					BL float64
				}{1125.3447, 1112, 1098.6554},
			},
		},
		{
			name: "ok",
			args: args{
				values: generateValues(1000, 1300, 1000),
			},
			want: Signals{
				SMA: 1112,
				WMA: 1113.0909,
				EMA: 1113.3636,
				BB: struct {
					TL float64
					ML float64
					BL float64
				}{1125.3447, 1112, 1098.6554},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		b.Run(tt.name, func(b *testing.B) {
			c := &Calc{}
			c.CalcSignals(tt.args.values, talib.SMA)
		})
	}
}

func generateValues(n int, max, min float64) []float64 {
	rand.Seed(time.Now().UnixNano())
	var result = make([]float64, 0)
	for i := 0; i < n; i++ {
		result = append(result, min+rand.Float64()*(max-min))
	}
	return result
}

func TestCalc_CalculateRSI(t *testing.T) {
	assertr := assert.New(t)
	type args struct {
		values     []float64
		indication MAIndication
	}
	tests := []struct {
		name string
		args args
		want RSI
	}{
		{
			name: "ok",
			args: args{
				values: []float64{
					1100,
					1115,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
					1125,
					1130,
					1128,
					1120,
					1120,
					1122,
					1125,
					1130,
					1128,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
					1110,
					1105,
				},
				indication: SMAIndication,
			},
			want: RSI{
				Value: 46.6843,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalcRSI(tt.args.values, 14)
			assertr.Equal(tt.want, got, "rsi calculation")
		})
	}
}

func TestCalc_CalculateMACD(t *testing.T) {
	assertr := assert.New(t)
	type args struct {
		values []float64
	}
	tests := []struct {
		name string
		args args
		want MACD
	}{
		{
			name: "ok",
			args: args{
				values: []float64{
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
					1125,
					1130,
					1128,
					1120,
					1138,
					1134,
					1128,
					1124,
					1111,
					1003,
					1110,
					1111,
					1127,
					1128,
					1125,
					1120,
					1100,
					1115,
					1120,
					1118,
					1110,
					1105,
					1102,
					1108,
					1120,
					1122,
					1125,
					1130,
					1128,
					1120,
				},
			},
			want: MACD{
				HistogramValue: 1.282,
				Value:          1.447,
				Sig:            0.166,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalcMACD(tt.args.values, 14, talib.EMA, 26, talib.EMA, 9, talib.WMA)
			assertr.Equal(tt.want, got, "MACD calculation")
		})
	}
}
