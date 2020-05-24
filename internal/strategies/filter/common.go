package filter

import "github.com/tagirmukail/tccbot-backend/internal/strategies/types"

func checkTrend(trendActs []types.Action, trendAct types.Action) bool {
	for _, act := range trendActs {
		if act != trendAct {
			return false
		}
	}
	return true
}

func checkAnyTrend(trendActs []types.Action, trendAct types.Action) bool {
	for _, act := range trendActs {
		if act == trendAct {
			return true
		}
	}
	return false
}
