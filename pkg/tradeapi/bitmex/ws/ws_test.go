package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tagirmukail/tccbot-backend/internal/types"
)

func Test_buildSubscribeParams(t *testing.T) {
	type args struct {
		symbol types.Symbol
		themes []types.Theme
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ok",
			args: args{
				symbol: "XBTUSD",
				themes: []types.Theme{"tradeBin5m"},
			},
			want: "subscribe=tradeBin5m%3AXBTUSD",
		},
		{
			name: "ok",
			args: args{
				symbol: "XBTUSD",
				themes: []types.Theme{"tradeBin5m", "tradeBin1m", "tradeBin1h"},
			},
			want: "subscribe=tradeBin5m%3AXBTUSD%2CtradeBin1m%3AXBTUSD%2CtradeBin1h%3AXBTUSD",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := buildSubscribeParams(tt.args.symbol, tt.args.themes)
			assert.Equal(t, tt.want, got)
		})
	}
}
