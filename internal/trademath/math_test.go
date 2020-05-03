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
