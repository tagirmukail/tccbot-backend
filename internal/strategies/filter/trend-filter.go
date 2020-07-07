package filter

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	stratypes "github.com/tagirmukail/tccbot-backend/internal/strategies/types"
	"github.com/tagirmukail/tccbot-backend/internal/types"
)

// TrendFilter - this filter is designed to filter out false signals
type TrendFilter struct {
	prevActions []stratypes.Action
	cfg         *config.GlobalConfig
	log         *logrus.Logger
}

func NewTrendFilter(cfg *config.GlobalConfig, log *logrus.Logger) *TrendFilter {

	return &TrendFilter{
		cfg:         cfg,
		prevActions: make([]stratypes.Action, 0),
		log:         log,
	}
}

func (f *TrendFilter) Apply(ctx context.Context) types.Side {
	ctxData, err := getFromCtx(ctx)
	if err != nil {
		f.log.Debugf("TrendFilter.Apply failed: %v", err)
		return types.SideEmpty
	}

	return f.checkAction(ctxData.action, ctxData.binSize)
}

func (f *TrendFilter) checkAction(action stratypes.Action, size models.BinSize) types.Side {
	f.addInPrevAction(action, size)
	// тренд прерывается, проверяем, если первый тренд экшенов - восходящий тренд,
	// а остальные - это иные экшены, то тогда выставляем на продажу, если первый - нисходящий тренд,
	// а остальные иные экшены, то выставляем на покупку, если остальные экшены(хотя бы один) такие же как и первый,
	// то продолжаем наблюдение, выходим без действия

	// the trend is interrupted, we check if the first trend of actions is an uptrend,
	// and the rest are different actions,
	// then we put up for sale, if the first is a downtrend and the rest are other actions,
	// then we put up for purchase, if the rest of the actions (at least one) same as the first one,
	// then continue to observe, exit without action
	f.log.Debug("TrendFilter.Apply - checkAction check prev actions")
	return f.checkPrevActions(size)
}

func (f *TrendFilter) checkPrevActions(size models.BinSize) types.Side {
	cfg := f.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		f.log.Errorf("cfg by bin size is empty")
		return types.SideEmpty
	}
	if len(f.prevActions) < cfg.MaxFilterTrendCount {
		f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - prev action count less than max_filter_trend_count, exit")
		return types.SideEmpty
	}
	firstAction := f.prevActions[0]
	secondHalf := f.prevActions[1:]

	isUpFirstAction := firstAction == stratypes.UpTrend
	isAnyUpSecondHalf := checkAnyTrend(secondHalf, stratypes.UpTrend)
	if isUpFirstAction && isAnyUpSecondHalf { // in the second half actions exist up, continue to observe
		f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - up trend is continue, exit")
		return types.SideEmpty
	}
	if isUpFirstAction && !isAnyUpSecondHalf { // in the second half not exist up actions, trend ended movement - sell
		f.prevActions = f.prevActions[:0]
		f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - up trend is complete, place sell, exit")
		return types.SideSell
	}

	isDownFirstAction := firstAction == stratypes.DownTrend
	isAnyDownSecondHalf := checkAnyTrend(secondHalf, stratypes.DownTrend)
	if isDownFirstAction && isAnyDownSecondHalf { // in the second half actions exist down, continue to observe
		f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - down trend is continue, exit")
		return types.SideEmpty
	}
	if isDownFirstAction && !isAnyDownSecondHalf { // in the second half not exist down actions, trend ended movement - buy
		f.prevActions = f.prevActions[:0]
		f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - down trend is complete, place buy, exit")
		return types.SideBuy
	}

	f.log.Debug("TrendFilter.Apply - checkAction - checkPrevActions - not up and not down trend, exit")
	return types.SideEmpty
}

func (f *TrendFilter) addInPrevAction(action stratypes.Action, size models.BinSize) {
	cfg := f.cfg.GlobStrategies.GetCfgByBinSize(size.String())
	if cfg == nil {
		f.log.Errorf("cfg by bin size is empty")
		return
	}
	f.prevActions = append(f.prevActions, action)
	if len(f.prevActions) >= cfg.MaxFilterTrendCount {
		f.prevActions = f.prevActions[len(f.prevActions)-cfg.MaxFilterTrendCount:]
	}
}
