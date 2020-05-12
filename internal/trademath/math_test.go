package trademath

import (
	"math/rand"
	"testing"
	"time"

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
		want Singals
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
				},
			},
			want: Singals{
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
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalculateSignals(tt.args.values)
			assertr.Equal(got, tt.want)
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
		want Singals
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
			want: Singals{
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
			want: Singals{
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
		b.Run(tt.name, func(b *testing.B) {
			c := &Calc{}
			c.CalculateSignals(tt.args.values)
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
			name: "by sma",
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
				},
				indication: SMAIndication,
			},
			want: RSI{
				Value: 59.503,
			},
		},
		{
			name: "by ema",
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
				},
				indication: EMAIndication,
			},
			want: RSI{
				Value: 49.771,
			},
		},
		{
			name: "by wma",
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
				},
				indication: WMAIndication,
			},
			want: RSI{
				Value: 52.861,
			},
		},
		{
			name: "unknown indication",
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
				},
				indication: 10,
			},
			want: RSI{
				Value: 0,
			},
		},
		{
			name: "values less than 2",
			args: args{
				values: []float64{
					1100,
				},
				indication: SMAIndication,
			},
			want: RSI{
				Value: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalculateRSI(tt.args.values, tt.args.indication)
			assertr.Equal(tt.want, got, "rsi calculation")
		})
	}
}

func TestCalc_CalculateMACD(t *testing.T) {
	assertr := assert.New(t)
	type args struct {
		fastValues     []float64
		slowValues     []float64
		prevMACDValues []float64
		indication     MAIndication
	}
	tests := []struct {
		name string
		args args
		want MACD
	}{
		{
			name: "ok",
			args: args{
				fastValues: []float64{
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
				slowValues: []float64{
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
				prevMACDValues: []float64{
					3.20,
					-2.11,
					3.01,
					1.15,
					-4,
					-1,
					1.2,
					2.3,
				},
				indication: EMAIndication,
			},
			want: MACD{
				HistogramValue: 2.907,
				Value:          3.857,
				Sig:            0.9514,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calc{}
			got := c.CalculateMACD(tt.args.fastValues, tt.args.slowValues, tt.args.prevMACDValues, tt.args.indication)
			assertr.Equal(tt.want, got, "MACD calculation")
		})
	}
}
