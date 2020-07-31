package bitmex

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/crypto"
)

type Bitmex struct {
	key              string
	secret           string
	url              string
	verbose          bool
	defaultUserAgent string
	retryCount       int
	maxRequestsLimit int32
	requestsCount    int32
	logger           *logrus.Logger
	idleConnTimeout  time.Duration
	maxIdleConns     int
	timeout          time.Duration
	rwLock           sync.RWMutex
	ws               *ws.WS
	authWS           *ws.WS
}

type Request struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        io.Reader
	Response    interface{}
	AuthRequest bool
	Verbose     bool
	Endpoint    EndpointLimit
}

func New(
	key,
	secret string,
	verbose bool,
	retryCount int,
	idleConnTimeout time.Duration,
	maxIdleConns int,
	maxRequestsLimit int32,
	timeout time.Duration,
	ws *ws.WS,
	authWS *ws.WS,
	logger *logrus.Logger,
) *Bitmex {
	if maxRequestsLimit == 0 {
		maxRequestsLimit = maxRequests
	}
	return &Bitmex{
		key:              key,
		secret:           secret,
		url:              bitmexUrl,
		verbose:          verbose,
		retryCount:       retryCount,
		maxRequestsLimit: maxRequestsLimit,
		logger:           logger,
		idleConnTimeout:  idleConnTimeout,
		maxIdleConns:     maxIdleConns,
		timeout:          timeout,
		rwLock:           sync.RWMutex{},
		ws:               ws,
		authWS:           authWS,
	}
}

func (b *Bitmex) EnableTestNet() {
	b.url = testnetUrl
}

func (b *Bitmex) GetWS() *ws.WS {
	return b.ws
}

func (b *Bitmex) GetAuthWS() *ws.WS {
	return b.authWS
}

func (b *Bitmex) SetDefaultUserAgent(agent string) {
	b.defaultUserAgent = agent
}

func (b *Bitmex) getClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: b.idleConnTimeout,
			MaxIdleConns:    b.maxIdleConns,
		},
		Timeout: b.timeout,
	}
}

func (b *Bitmex) validateRequest() error {
	if b.url == "" {
		return errors.New("empty url")
	}
	if b.key == "" {
		return errors.New("empty key")
	}
	if b.secret == "" {
		return errors.New("empty secret")
	}
	return nil
}

func (b *Bitmex) SendRequest(path string, params url.Values, response interface{}) error {
	if b.url == "" {
		return errors.New("bitmex url is empty")
	}
	if path == "" {
		return errors.New("path is empty")
	}
	var uri = b.url + path
	if params != nil {
		encodeParams := params.Encode()
		uri += "?" + encodeParams
	}

	return b.do(&Request{
		Method:      http.MethodGet,
		Path:        uri,
		Response:    &response,
		AuthRequest: false,
		Verbose:     b.verbose,
	})
}

func (b *Bitmex) SendAuthenticatedRequest(
	verb, path string, params, response interface{},
) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	if err := b.validateRequest(); err != nil {
		return err
	}

	timestamp := time.Now().Add(time.Second * 10).UnixNano()
	timestampStr := strconv.FormatInt(timestamp, 10)
	timestampNew := timestampStr[:13]

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["api-expires"] = timestampNew
	headers["api-key"] = b.key

	var data string
	if params != nil {
		bData, err := json.Marshal(params)
		if err != nil {
			return err
		}
		data = string(bData)
		if b.verbose {
			b.logger.Infof("request params: %s", data)
		}
	}
	hmac := crypto.GetHashMessage(crypto.HashSHA256,
		[]byte(verb+"/api/v1"+path+timestampNew+data),
		[]byte(b.secret))
	headers["api-signature"] = crypto.HexEncodeToString(hmac)

	if err := b.do(&Request{
		Method:      verb,
		Path:        b.url + path,
		Headers:     headers,
		Body:        bytes.NewBuffer([]byte(data)),
		Response:    &response,
		AuthRequest: true,
		Verbose:     b.verbose,
		Endpoint:    Auth,
	}); err != nil {
		return err
	}
	return nil
}

func (b *Bitmex) do(item *Request) error {
	b.rwLock.RLock()
	b.rwLock.RUnlock()

	cli := b.getClient()
	if err := b.validateRequestItem(item); err != nil {
		return err
	}
	req, err := http.NewRequest(item.Method, item.Path, item.Body)
	if err != nil {
		return err
	}

	for key, value := range item.Headers {
		req.Header.Add(key, value)
	}
	if b.defaultUserAgent != "" && req.Header.Get(userAgent) != "" {
		req.Header.Add(userAgent, b.defaultUserAgent)
	}

	if b.verbose {
		b.logger.Debugf("request method:%s, path: %s", item.Method, item.Path)
		for k, v := range req.Header {
			b.logger.Debugf("path:%s request header[%s]:%s", item.Path, k, v)
		}
	}

	if atomic.LoadInt32(&b.requestsCount) >= b.maxRequestsLimit {
		return errors.New("max request limit exceeded")
	}

	var resp *http.Response
	for i := 0; i < b.retryCount; i++ {
		atomic.AddInt32(&b.requestsCount, 1)
		resp, err = cli.Do(req)
		atomic.AddInt32(&b.requestsCount, -1)
		if err != nil {
			if b.verbose {
				b.logger.Errorf("path:%s error request, attempt:%d, error:%v", item.Path, i, err)
			}
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	if b.verbose {
		for k, v := range resp.Header {
			b.logger.Infof("response header[%s]:[%v]", k, v[0])
		}
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode > http.StatusAccepted {
		return fmt.Errorf("path:%s unsuccessful HTTP status code: %d  raw response: %s",
			item.Path,
			resp.StatusCode,
			string(content),
		)
	}
	resp.Body.Close()
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err = json.Unmarshal(content, item.Response)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bitmex) validateRequestItem(item *Request) error {
	if item == nil {
		return errors.New("empty request item")
	}
	if item.Path == "" {
		return errors.New("invalid path")
	}
	if item.Response == nil {
		return errors.New("response point must be not nil")
	}
	return nil
}
