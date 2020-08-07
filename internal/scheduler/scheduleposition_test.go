package scheduler

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/config"
)

func TestPositionScheduler_checkPlaceOrder(t *testing.T) {
	type fields struct {
		log              *logrus.Logger
		cfg              *config.GlobalConfig
		pnlT             positionPnl
		positionPnlLimit int
	}
	type args struct {
		p *positionPnl
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantPnl positionPnl
		want    bool
	}{
		{
			name: "update o.pnlT - profit",
			fields: fields{
				log: logrus.New(),
				cfg: &config.GlobalConfig{
					Scheduler: config.Scheduler{
						Position: config.PositionScheduler{
							Enable:        true,
							PriceTrailing: 5,
							ProfitPnlDiff: 0.03,
						},
					},
				},
				pnlT: positionPnl{
					pnl: 0.04,
					t:   Profit,
				},
			},
			args: args{
				p: &positionPnl{
					pnl: 0.08,
					t:   Profit,
				},
			},
			wantPnl: positionPnl{
				pnl: 0.08,
				t:   Profit,
			},
			want: false,
		},
		{
			name: "2 update o.pnlT - profit",
			fields: fields{
				log: logrus.New(),
				cfg: &config.GlobalConfig{
					Scheduler: config.Scheduler{
						Position: config.PositionScheduler{
							Enable:        true,
							PriceTrailing: 5,
							ProfitPnlDiff: 0.03,
						},
					},
				},
				pnlT: positionPnl{
					pnl: 0.04,
					t:   Profit,
				},
			},
			args: args{
				p: &positionPnl{
					pnl: 0.06,
					t:   Profit,
				},
			},
			wantPnl: positionPnl{
				pnl: 0.06,
				t:   Profit,
			},
			want: false,
		},
		{
			name: "place order o.pnlT - profit",
			fields: fields{
				log: logrus.New(),
				cfg: &config.GlobalConfig{
					Scheduler: config.Scheduler{
						Position: config.PositionScheduler{
							Enable:        true,
							PriceTrailing: 5,
							ProfitPnlDiff: 0.03,
						},
					},
				},
				pnlT: positionPnl{
					pnl: 0.05,
					t:   Profit,
				},
			},
			args: args{
				p: &positionPnl{
					pnl: 0.019,
					t:   Profit,
				},
			},
			wantPnl: positionPnl{
				pnl: 0.05,
				t:   Profit,
			},
			want: true,
		},
		{
			name: "not place order o.pnlT - profit, p.pnl less than o.pnlT.pnl",
			fields: fields{
				log: logrus.New(),
				cfg: &config.GlobalConfig{
					Scheduler: config.Scheduler{
						Position: config.PositionScheduler{
							Enable:        true,
							PriceTrailing: 5,
							ProfitPnlDiff: 0.03,
						},
					},
				},
				pnlT: positionPnl{
					pnl: 0.05,
					t:   Profit,
				},
			},
			args: args{
				p: &positionPnl{
					pnl: 0.02,
					t:   Profit,
				},
			},
			wantPnl: positionPnl{
				pnl: 0.05,
				t:   Profit,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			o := &PositionScheduler{
				log:              tt.fields.log,
				cfg:              tt.fields.cfg,
				mx:               sync.Mutex{},
				pnlT:             tt.fields.pnlT,
				positionPnlLimit: tt.fields.positionPnlLimit,
			}
			got := o.checkPlaceOrder(tt.args.p)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantPnl, o.pnlT)
		})
	}
}
