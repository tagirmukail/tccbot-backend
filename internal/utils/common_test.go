package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestFromTime(t *testing.T) {
	type args struct {
		now     time.Time
		binSize string
		count   int
	}
	tests := []struct {
		name     string
		args     args
		wantFrom time.Time
		wantErr  bool
	}{
		{
			name: "1m",
			args: args{
				now:     time.Date(2010, 10, 10, 10, 10, 10, 10, time.UTC),
				binSize: "1m",
				count:   10,
			},
			wantFrom: time.Date(2010, 10, 10, 10, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "5m",
			args: args{
				now:     time.Date(2010, 10, 10, 10, 0, 10, 10, time.UTC),
				binSize: "5m",
				count:   10,
			},
			wantFrom: time.Date(2010, 10, 10, 9, 10, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "5m - 20 times",
			args: args{
				now:     time.Date(2010, 10, 10, 10, 0, 10, 10, time.UTC),
				binSize: "5m",
				count:   20,
			},
			wantFrom: time.Date(2010, 10, 10, 8, 20, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "1h",
			args: args{
				now:     time.Date(2010, 10, 10, 10, 0, 10, 10, time.UTC),
				binSize: "1h",
				count:   10,
			},
			wantFrom: time.Date(2010, 10, 10, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "1d",
			args: args{
				now:     time.Date(2010, 10, 10, 10, 0, 10, 10, time.UTC),
				binSize: "1d",
				count:   10,
			},
			wantFrom: time.Date(2010, 10, 0, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, err := FromTime(tt.args.now, tt.args.binSize, tt.args.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFrom, tt.wantFrom) {
				t.Errorf("FromTime() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
			}
		})
	}
}
