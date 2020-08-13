package strategies

import (
	"reflect"
	"testing"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

func TestStrategies_findTwoMax(t *testing.T) {
	type args struct {
		tframe  int
		signals []*models.Signal
	}
	tests := []struct {
		name       string
		args       args
		wantResult struct {
			firstMaxIndx  int
			firstMax      float64
			secondMaxIndx int
			secondMax     float64
		}
		wantErr bool
	}{
		{
			name: "not bearDiverg",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: 4.5,
					},
					{
						MACDHistogramValue: 3.2,
					},
					{
						MACDHistogramValue: 10.8,
					},
					{
						MACDHistogramValue: 8.8,
					},
					{
						MACDHistogramValue: 11.11,
					},
				},
			},
			wantResult: struct {
				firstMaxIndx  int
				firstMax      float64
				secondMaxIndx int
				secondMax     float64
			}{
				firstMaxIndx:  4,
				firstMax:      11.11,
				secondMaxIndx: 2,
				secondMax:     10.8,
			},
			wantErr: false,
		},
		{
			name: "max for bearDiverg",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: 4.5,
					},
					{
						MACDHistogramValue: 12.4,
					},
					{
						MACDHistogramValue: 7.8,
					},
					{
						MACDHistogramValue: 8.8,
					},
					{
						MACDHistogramValue: 3.3,
					},
				},
			},
			wantResult: struct {
				firstMaxIndx  int
				firstMax      float64
				secondMaxIndx int
				secondMax     float64
			}{
				firstMaxIndx:  1,
				firstMax:      12.4,
				secondMaxIndx: 3,
				secondMax:     8.8,
			},
			wantErr: false,
		},
		{
			name: "max for bearDiverg, first max is equal second max",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: 4.5,
					},
					{
						MACDHistogramValue: 12.4,
					},
					{
						MACDHistogramValue: 7.8,
					},
					{
						MACDHistogramValue: 12.4,
					},
					{
						MACDHistogramValue: 3.3,
					},
				},
			},
			wantResult: struct {
				firstMaxIndx  int
				firstMax      float64
				secondMaxIndx int
				secondMax     float64
			}{
				firstMaxIndx:  3,
				firstMax:      12.4,
				secondMaxIndx: 1,
				secondMax:     12.4,
			},
			wantErr: false,
		},
		{
			name: "exist negative histogram values",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: 4.5,
					},
					{
						MACDHistogramValue: 12.4,
					},
					{
						MACDHistogramValue: 7.8,
					},
					{
						MACDHistogramValue: -1.8,
					},
					{
						MACDHistogramValue: 8.8,
					},
					{
						MACDHistogramValue: 3.3,
					},
					{
						MACDHistogramValue: 7.8,
					},
					{
						MACDHistogramValue: 4.8,
					},
				},
			},
			wantResult: struct {
				firstMaxIndx  int
				firstMax      float64
				secondMaxIndx int
				secondMax     float64
			}{
				firstMaxIndx:  4,
				firstMax:      8.8,
				secondMaxIndx: 6,
				secondMax:     7.8,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategies{}
			gotResult, err := s.findTwoMax(tt.args.tframe, tt.args.signals)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTwoMax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("findTwoMax() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestStrategies_findTwoMin(t *testing.T) {
	type args struct {
		tframe  int
		signals []*models.Signal
	}
	tests := []struct {
		name       string
		args       args
		wantResult struct {
			firstMinIndx  int
			firstMin      float64
			secondMinIndx int
			secondMin     float64
		}
		wantErr bool
	}{
		{
			name: "find two min, bullDiverg - ok",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: -4,
					},
					{
						MACDHistogramValue: -3,
					},
					{
						MACDHistogramValue: 1,
					},
					{
						MACDHistogramValue: 2,
					},
					{
						MACDHistogramValue: -1,
					},
					{
						MACDHistogramValue: -4,
					},
					{
						MACDHistogramValue: -3,
					},
					{
						MACDHistogramValue: -2,
					},
				},
			},
			wantResult: struct {
				firstMinIndx  int
				firstMin      float64
				secondMinIndx int
				secondMin     float64
			}{
				firstMinIndx:  5,
				firstMin:      -4,
				secondMinIndx: 6,
				secondMin:     -3,
			},
			wantErr: false,
		},
		{
			name: "find two min, bullDiverg invalid",
			args: args{
				tframe: 3,
				signals: []*models.Signal{
					{
						MACDHistogramValue: -4,
					},
					{
						MACDHistogramValue: -3,
					},
					{
						MACDHistogramValue: 1,
					},
					{
						MACDHistogramValue: 2,
					},
					{
						MACDHistogramValue: -10,
					},
					{
						MACDHistogramValue: -4,
					},
					{
						MACDHistogramValue: -3,
					},
					{
						MACDHistogramValue: -12,
					},
				},
			},
			wantResult: struct {
				firstMinIndx  int
				firstMin      float64
				secondMinIndx int
				secondMin     float64
			}{
				firstMinIndx:  7,
				firstMin:      -12,
				secondMinIndx: 4,
				secondMin:     -10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategies{}
			gotResult, err := s.findTwoMin(tt.args.tframe, tt.args.signals)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTwoMin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("findTwoMin() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
