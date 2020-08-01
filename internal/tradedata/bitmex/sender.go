package bitmextradedata

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

type Sender struct {
	messages    chan *data.BitmexData
	subscribers []*Subscriber
	log         *logrus.Logger
}

func New(messages chan *data.BitmexData, log *logrus.Logger, subscribers ...*Subscriber) *Sender {
	return &Sender{
		messages:    messages,
		subscribers: subscribers,
		log:         log,
	}
}

func (s *Sender) SendToSubscribers(wg *sync.WaitGroup) {
	s.log.Infof("bitmex trade data sender started")
	defer func() {
		s.log.Infof("bitmex trade data sender finished")
		wg.Done()
	}()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-done:
			return
		case msg := <-s.messages:
			for _, subs := range s.subscribers {
				if subs.isSubscriberTheme(types.Theme(msg.Table)) {
					subs.messages <- msg
				}
			}
		}
	}

}
