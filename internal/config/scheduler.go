package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Scheduler struct {
	Position PositionScheduler
}

type PositionScheduler struct {
	Enable         bool
	PriceTrailing  float64
	ProfitCloseBTC float64
	LossCloseBTC   float64
	ProfitPnlDiff  float64
	LossPnlDiff    float64
}

func initSchedulers() Scheduler {
	schdl := Scheduler{
		Position: PositionScheduler{
			Enable:         viper.GetBool("scheduler.position.enable"),
			PriceTrailing:  viper.GetFloat64("scheduler.position.trailing_price"),
			ProfitCloseBTC: viper.GetFloat64("scheduler.position.profit_close_btc"),
			LossCloseBTC:   viper.GetFloat64("scheduler.position.loss_close_btc"),
			ProfitPnlDiff:  viper.GetFloat64("scheduler.position.profit_pnl_diff"),
			LossPnlDiff:    viper.GetFloat64("scheduler.position.loss_pnl_diff"),
		},
	}

	fmt.Println("--------------------------------------------")
	fmt.Printf("scheduler: %#v\n", schdl)
	fmt.Println("--------------------------------------------")
	return schdl
}
