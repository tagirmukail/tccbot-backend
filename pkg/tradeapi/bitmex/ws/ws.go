package ws

import (
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/recws-org/recws"
	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/types"
)

const (
	timeReadSleep = 3 * time.Second
)

type Receiver interface {
}

type WS struct {
	log *logrus.Logger

	ws       *recws.RecConn
	isWorked bool
	connUrl  url.URL

	pingInterval int // ping interval in second
	timeout      int // in second

	theme    []types.Theme
	symbol   types.Symbol
	messages chan *BitmexData
}

func NewWS(
	log *logrus.Logger,
	wsUrl url.URL,
	ping int,
	timeout int,
	retrySec uint32,
	theme []types.Theme,
	symbol types.Symbol,
) *WS {
	wsr := &WS{
		log: log,
		ws: &recws.RecConn{
			RecIntvlMin:      time.Duration(retrySec) * time.Second,
			RecIntvlMax:      time.Duration(retrySec) * time.Second,
			KeepAliveTimeout: time.Duration(timeout) * time.Second,
			NonVerbose:       true,
		},
		connUrl:      wsUrl,
		pingInterval: ping,
		timeout:      timeout,
		theme:        theme,
		symbol:       symbol,
		messages:     make(chan *BitmexData),
	}
	wsr.ws.SubscribeHandler = wsr.subscribeHandler

	return wsr
}

func (r *WS) GetMessages() chan *BitmexData {
	return r.messages
}

// Start start reads bitmex messages
func (r *WS) Start() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	r.ws.Dial(r.connUrl.String(), nil)
	err := r.ws.GetDialError()
	if err != nil {
		r.log.Fatalf("binance not connected, error: %v", err)
	}
	defer r.ws.Close()

	r.log.Infof("connected to %s", r.connUrl.String())

	go r.read()

	r.isWorked = true

	<-done
}

// read reads messages from ws connection to bitmex and sends this to messages chanel
func (r *WS) read() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	j := jsoniter.ConfigCompatibleWithStandardLibrary
	r.log.Infof("WS.read read trades:%v from bitmex started", r.theme)
	for {
		mType, data, err := r.ws.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			r.log.Warnf("bitmex WS.read() read message from websocket error: %v", err)
			time.Sleep(timeReadSleep)
			continue
		}

		switch mType {
		case websocket.CloseMessage:
			r.log.Infof("WS.read() %s websocket bitmex closed", r.connUrl.String())
			close(r.messages)
			return
		case websocket.PingMessage:
			err = r.ws.WriteMessage(websocket.PongMessage, []byte("pong"))
			if err != nil {
				r.log.Infof("WS.read() %s websocket send pong message error: %v", r.connUrl.String(), err)
			}
			r.log.Infof("send pong message")
			continue
		case websocket.PongMessage:
			err = r.ws.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				r.log.Infof("WS.read() %s websocket send pong message error: %v", r.connUrl.String(), err)
			}
			r.log.Infof("send ping message")
			continue
		case websocket.TextMessage:
			break
		default:
			r.log.Warnln("bitmex WS.read() websocket message type is not text")
			continue
		}

		if string(data) == "pong" {
			r.log.Infoln("bitmex ping message received")
			continue
		}

		resp := &BitmexData{}
		err = j.Unmarshal(data, resp)
		if err != nil {
			r.log.Warnf("bitmex WS.read() unmarshal websocket message error: %v", err)
			r.log.Warnf("bitmex WS.read() unmarshal websocket message error - data: %v", string(data))
			continue
		}

		err = resp.Validate()
		if err != nil {
			r.log.Warnf("bitmex WS.read() validate websocket message error: %v", err)
			continue
		}

		select {
		case <-done:
			r.log.Infof("Stopping processing messages from bitmex")
			close(r.messages)
			return
		case r.messages <- resp:
			//r.log.Debugf("bitmex sends message data: %s", string(data))
		default:
			continue
		}
	}

}

// subscribeHandler fires after the connection successfully establish and subscribed on ws messages by theme
func (r *WS) subscribeHandler() error {
	j := jsoniter.ConfigCompatibleWithStandardLibrary

	var themes []types.Theme
	for _, theme := range r.theme {
		themes = append(themes, types.NewTemeWithPair(theme, r.symbol))
	}
	subsMsg := types.NewSubscribeMsg(
		types.SubscribeAct,
		themes,
	)
	data, err := j.Marshal(subsMsg)
	if err != nil {
		r.log.Errorf("WS.Start() marshal error: %v", err)
		return err
	}

	if !r.ws.IsConnected() {
		r.ws.Dial(r.connUrl.String(), nil)
	}

	err = r.ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		r.log.Errorf("WS.Start() websocket write msg error: %v", err)
		r.log.Errorf("WS.Start() websocket write msg data: %#v", subsMsg)
		return err
	}

	r.log.Debugf("bitmex - send subscribe message: %s", string(data))

	return nil
}
