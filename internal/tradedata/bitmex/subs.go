package bitmextradedata

import (
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

type Subscriber struct {
	themes   []types.Theme
	messages chan *data.BitmexData
}

func NewSubscriber(themes []types.Theme) *Subscriber {
	return &Subscriber{
		themes:   themes,
		messages: make(chan *data.BitmexData),
	}
}

func (s *Subscriber) GetMsgChan() chan *data.BitmexData {
	return s.messages
}

func (s *Subscriber) isSubscriberTheme(theme types.Theme) bool {
	for _, th := range s.themes {
		if th == theme {
			return true
		}
	}

	return false
}
