package filter

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	stratypes "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

type checkCandlesTrend struct {
	currentMaxClosePrice float64
	candlesCurrentCount  uint32
	isUp                 bool
	isDown               bool
	side                 types.Side
	log                  *logrus.Logger
}

func (c *checkCandlesTrend) check(ctxData *getFromCtxData, cfg *config.StrategiesConfig) types.Side {
	var lastCandle bitmex.TradeBuck
	if len(ctxData.candles) > 1 {
		lastCandle = ctxData.candles[len(ctxData.candles)-1]
	}

	switch ctxData.action {
	case stratypes.UpTrend:
		c.isUp = true
		c.currentMaxClosePrice = lastCandle.Close
		if c.candlesCurrentCount != 0 { // сработал новый тренд, обнуляем счетчик
			c.candlesCurrentCount = 0
		}
		c.candlesCurrentCount++
		c.side = types.SideSell
		return types.SideEmpty
	case stratypes.DownTrend:
		c.isDown = true
		c.currentMaxClosePrice = lastCandle.Close
		if c.candlesCurrentCount != 0 { // сработал новый тренд, обнуляем счетчик
			c.candlesCurrentCount = 0
		}
		c.candlesCurrentCount++
		c.side = types.SideBuy
		return types.SideEmpty
	default:
		break
	}

	if c.candlesCurrentCount > 0 {
		c.candlesCurrentCount++
	}

	if c.isUp && lastCandle.Close > c.currentMaxClosePrice {
		c.currentMaxClosePrice = lastCandle.Close
	}

	if c.isDown && lastCandle.Close < c.currentMaxClosePrice {
		c.currentMaxClosePrice = lastCandle.Close
	}

	if c.candlesCurrentCount >= uint32(cfg.MaxCandlesFilterCount) {
		resultSide := c.side
		c.side = types.SideEmpty
		c.candlesCurrentCount = 0
		c.isDown = false
		c.isUp = false
		return resultSide
	}

	return types.SideEmpty
}

type CandlesFilter struct {
	checkTrend checkCandlesTrend
	actions    []stratypes.Action
	cfg        *config.GlobalConfig
	log        *logrus.Logger
}

func NewCandlesFilter(cfg *config.GlobalConfig, log *logrus.Logger) *CandlesFilter {
	return &CandlesFilter{
		checkTrend: checkCandlesTrend{
			side: types.SideEmpty,
			log:  log,
		},
		actions: make([]stratypes.Action, 0),
		cfg:     cfg,
		log:     log,
	}
}

func (f *CandlesFilter) Apply(ctx context.Context) types.Side {
	ctxData, err := getFromCtx(ctx)
	if err != nil {
		f.log.Debugf("CandlesFilter.Apply failed: %v", err)
		return types.SideEmpty
	}

	return f.apply(ctxData)
}

//  нужно определять по цене, если цена после срабатывания сигналов, продолжила рост, наблюдаем, и все время храним максимальное текущее значение, как только цена упала за определенное количество свечей, считаем, что рост закончился, выставляем ордер на продажу
// храним в переменной максимальное значение цены после срабатывания сигнала
// храним колличество прошедших свечей, после срабатывания сигнала
// как только рост закончился, обнуляем счетчик свечей и переменную с максимальной ценой
func (f *CandlesFilter) apply(ctxData *getFromCtxData) types.Side {
	cfg := f.cfg.GlobStrategies.GetCfgByBinSize(ctxData.binSize.String())
	if cfg == nil {
		f.log.Errorf("cfg by bin size is empty")
		return types.SideEmpty
	}

	// add current action to cache
	f.actions = append(f.actions, ctxData.action)
	if len(f.actions) >= cfg.MaxCandlesFilterCount {
		f.actions = f.actions[len(f.actions)-cfg.MaxCandlesFilterCount:]
	}

	return f.checkTrend.check(ctxData, cfg)
}
