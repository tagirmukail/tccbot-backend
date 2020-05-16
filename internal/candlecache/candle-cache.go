package candlecache

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws/data"
)

type Caches interface {
	GetCache(size models.BinSize) Cache
}

type Cache interface {
	StoreBatch(batch []bitmex.TradeBuck)
	Store(candle data.BitmexIncomingData) error
	GetBucketed(from, to time.Time, count int) []bitmex.TradeBuck
	Count() int
}

type BinToCache struct {
	sync.Mutex
	caches map[string]Cache
}

func NewBinToCache(binSizes []string, maxCount int, symbol types.Symbol, log *logrus.Logger) *BinToCache {
	binToCache := &BinToCache{
		Mutex:  sync.Mutex{},
		caches: make(map[string]Cache),
	}
	for _, bin := range binSizes {
		binToCache.caches[bin] = NewCandleCache(maxCount, symbol, log)
	}

	return binToCache
}

func (bc *BinToCache) GetCache(size models.BinSize) Cache {
	return bc.caches[size.String()]
}

type CandleCache struct {
	sync.Mutex
	store    []bitmex.TradeBuck
	maxCount int
	symbol   types.Symbol
	log      *logrus.Logger
}

func NewCandleCache(maxCount int, symbol types.Symbol, log *logrus.Logger) *CandleCache {
	return &CandleCache{
		Mutex:    sync.Mutex{},
		store:    make([]bitmex.TradeBuck, 0, maxCount),
		maxCount: maxCount,
		symbol:   symbol,
		log:      log,
	}
}

func (c *CandleCache) StoreBatch(batch []bitmex.TradeBuck) {
	c.Lock()
	for _, elem := range batch {
		if elem.Symbol != string(c.symbol) {
			continue
		}
		c.store = append(c.store, elem)
	}
	c.sort()
	if len(c.store) > c.maxCount {
		c.store = c.store[len(c.store)-c.maxCount:]
	}
	c.Unlock()
}
func (c *CandleCache) Store(candle data.BitmexIncomingData) error {
	c.Lock()
	defer c.Unlock()
	if candle.Symbol != c.symbol {
		return errors.New("candle not saved")
	}
	c.store = append(c.store, bitmex.TradeBuck{
		Symbol:          string(candle.Symbol),
		Timestamp:       candle.Timestamp,
		HomeNotional:    candle.HomeNotional,
		ForeignNotional: candle.ForeignNotional,
		Open:            candle.Open,
		High:            candle.High,
		Low:             candle.Low,
		Close:           candle.Close,
		Trades:          candle.Trades,
		Volume:          candle.Volume,
		LastSize:        candle.LastSize,
		Turnover:        candle.Turnover,
		Vwap:            candle.Vwap,
	})
	c.sort()
	if len(c.store) > c.maxCount {
		c.store = c.store[len(c.store)-c.maxCount:]
	}
	return nil
}

func (c *CandleCache) GetBucketed(
	from,
	to time.Time,
	count int,
) []bitmex.TradeBuck {
	var result []bitmex.TradeBuck
	if count > c.maxCount {
		count = c.maxCount
	}
	fromToNotZero := !from.IsZero() && !to.IsZero()
	fromZero := from.IsZero()
	toZero := to.IsZero()

	c.Lock()
	defer c.Unlock()

	c.sort()
	for _, candle := range c.store {
		candleTs, err := time.Parse(
			tradeapi.TradeBucketedTimestampLayout,
			candle.Timestamp,
		)
		if err != nil {
			c.log.Errorf("candle timestamp fail: %v", err)
			return nil
		}
		if fromToNotZero {
			if (candleTs.After(from) && candleTs.Before(to)) || candleTs.Equal(from) || candleTs.Equal(to) {
				result = append(result, candle)
			}
			continue
		} else {
			if (!fromZero && candleTs.After(from)) || (!fromZero && candleTs.Equal(from)) {
				result = append(result, candle)
				continue
			} else if (!toZero && candleTs.Before(to)) || (!toZero && candleTs.Equal(to)) {
				result = append(result, candle)
				continue
			}
		}
	}
	if count != 0 && len(result) > count {
		result = result[len(result)-count:]
	} else if len(result) == 0 && count != 0 {
		result = c.store[len(c.store)-count:]
	}

	return result
}

func (c *CandleCache) Count() int {
	c.Lock()
	defer c.Unlock()
	return len(c.store)
}

func (c *CandleCache) sort() {
	sort.SliceStable(c.store, func(i, j int) bool {
		timestamp1, err := time.Parse(
			tradeapi.TradeBucketedTimestampLayout,
			c.store[i].Timestamp,
		)
		if err != nil {
			return false
		}
		timestamp2, err := time.Parse(
			tradeapi.TradeBucketedTimestampLayout,
			c.store[j].Timestamp,
		)
		if err != nil {
			return false
		}

		return timestamp1.Before(timestamp2)
	})
	return
}
