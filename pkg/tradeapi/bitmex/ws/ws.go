package ws

import (
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/crypto"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/recws"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

const (
	timeReadSleep    = 3 * time.Second
	handshakeTimeout = 5 * time.Second

	bitmexWSUrl        = "wss://www.bitmex.com/realtime"
	bitmexTestnetWSUrl = "wss://testnet.bitmex.com/realtime"
)

type Receiver interface {
}

type WS struct {
	log *logrus.Logger

	ws      *recws.RecConn
	connUrl string

	pingInterval int // ping interval in second
	timeout      int // in second

	theme    []types.Theme
	symbol   types.Symbol
	messages chan *data.BitmexData

	apiKey    string
	apiSecret string
}

func NewWS(
	log *logrus.Logger,
	test bool,
	ping int,
	timeout int,
	retrySec uint32,
	theme []types.Theme,
	symbol types.Symbol,
	apiKey string,
	apiSecret string,
) *WS {
	var bitmexUrl string
	if test {
		bitmexUrl = bitmexTestnetWSUrl
	} else {
		bitmexUrl = bitmexWSUrl
	}

	wsr := &WS{
		log: log,
		ws: &recws.RecConn{
			RecIntvlMin:      time.Duration(retrySec) * time.Second,
			RecIntvlMax:      time.Duration(retrySec) * time.Second,
			KeepAliveTimeout: time.Duration(timeout) * time.Second,
			NonVerbose:       true,
			HandshakeTimeout: handshakeTimeout,
		},
		connUrl:      bitmexUrl,
		pingInterval: ping,
		timeout:      timeout,
		theme:        theme,
		symbol:       symbol,
		messages:     make(chan *data.BitmexData),
		apiKey:       apiKey,
		apiSecret:    apiSecret,
	}

	wsr.connUrl = bitmexUrl
	wsr.ws.SubscribeHandler = wsr.subscribeAuthHandler

	return wsr
}

func (r *WS) GetMessages() chan *data.BitmexData {
	return r.messages
}

// Start start reads bitmex messages
func (r *WS) Start(wgForeign *sync.WaitGroup) {
	defer wgForeign.Done()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	r.ws.Dial(r.connUrl, nil)
	err := r.ws.GetDialError()
	if err != nil {
		r.log.Fatalf("bitmex not connected, error: %v", err)
	}
	defer r.ws.Close()

	r.log.Infof("connected to %s", r.connUrl)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go r.read(wg)
	wg.Add(1)
	go r.ping(wg)
	wg.Wait()
	<-done
}

// read reads messages from ws connection to bitmex and sends this to messages chanel
func (r *WS) read(wg *sync.WaitGroup) {
	defer wg.Done()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	j := jsoniter.ConfigCompatibleWithStandardLibrary
	r.log.Infof("WS.read read trades:%v from bitmex started", r.theme)
	for {
		mType, msg, err := r.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				r.log.Infof("read messages from bitmex ws stopped")
				return
			}
			r.log.Warnf("bitmex WS.read() read message from websocket error: %v", err)
			time.Sleep(timeReadSleep)
			continue
		}

		switch mType {
		case websocket.CloseMessage:
			r.log.Infof("WS.read() %s websocket bitmex closed", r.connUrl)
			close(r.messages)
			return
		case websocket.TextMessage:
			break
		default:
			r.log.Warnln("bitmex WS.read() websocket message type is not text")
			continue
		}

		if string(msg) == "pong" {
			r.log.Infoln("bitmex pong message received")
			continue
		}

		resp := &data.BitmexData{}
		err = j.Unmarshal(msg, resp)
		if err != nil {
			r.log.Warnf("bitmex WS.read() unmarshal websocket message error: %v", err)
			r.log.Warnf("bitmex WS.read() unmarshal websocket message error - data: %v", string(msg))
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

func (r *WS) subscribeHandler() error {
	return r.subscribe()
}

func (r *WS) subscribeAuthHandler() error {
	j := jsoniter.ConfigCompatibleWithStandardLibrary

	timestamp := time.Now().Add(time.Hour * 1).Unix()
	timestampNew := strconv.FormatInt(timestamp, 10)

	hmac := crypto.GetHashMessage(crypto.HashSHA256,
		[]byte("GET/realtime"+timestampNew),
		[]byte(r.apiSecret))

	authMsg := types.NewAuthMsg(r.apiKey, crypto.HexEncodeToString(hmac), timestamp)
	bData, err := j.Marshal(authMsg)
	if err != nil {
		r.log.Errorf("WS.subscribeAuthHandler() marshal error: %v", err)
		return err
	}
	err = r.ws.WriteMessage(websocket.TextMessage, bData)
	if err != nil {
		r.log.Errorf("WS.subscribeAuthHandler() websocket write [msg]:%#v error: %v", authMsg, err)
		return err
	}

	return r.subscribe()
}

// subscribeHandler fires after the connection successfully establish and subscribed on ws messages by theme
func (r *WS) subscribe() error {
	j := jsoniter.ConfigCompatibleWithStandardLibrary

	var themes []types.Theme
	for _, theme := range r.theme {
		if strings.Contains(string(theme), string(types.Trade)) {
			theme = types.NewTemeWithPair(theme, r.symbol)
		}
		themes = append(themes, theme)
	}

	subsMsg := types.NewSubscribeMsg(
		types.SubscribeAct,
		themes,
	)
	data, err := j.Marshal(subsMsg)
	if err != nil {
		r.log.Errorf("WS.subscribe() marshal error: %v", err)
		return err
	}

	err = r.ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		r.log.Errorf("WS.subscribe() websocket write msg error: %v", err)
		r.log.Errorf("WS.subscribe() websocket write msg data: %#v", subsMsg)
		return err
	}

	r.log.Debugf("bitmex - send subscribe message: %s", string(data))

	return nil
}

func (r *WS) ping(wg *sync.WaitGroup) {
	defer wg.Done()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	tick := time.NewTicker(time.Duration(r.pingInterval) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-done:
			err := r.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				r.log.Errorf("write close ws failed:%v", err)
			}
			return
		case <-tick.C:
			r.log.Debug("send ping message bitmex ws")
			err := r.ws.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				r.log.Errorf("ping message send failed: %v", err)
				if websocket.IsCloseError(err) {
					r.log.Fatal(err)
				}
			}
		}
	}

}

func buildSubscribeParams(symbol types.Symbol, themes []types.Theme) string {
	params := url.Values{}
	var subsParams []string
	for _, th := range themes {
		subsParams = append(subsParams, string(types.NewTemeWithPair(th, symbol)))
	}
	params.Add("subscribe", strings.Join(subsParams, ","))
	return params.Encode()
}
