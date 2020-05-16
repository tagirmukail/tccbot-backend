package strategies

import (
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/internal/trademath"
)

func (s *Strategies) saveSignals(timestamp time.Time, bin models.BinSize, n int, signals *trademath.Singals) error {
	_, err := s.db.SaveSignal(models.Signal{
		N:          n,
		BinSize:    bin,
		Timestamp:  timestamp,
		SignalType: models.BolingerBand,
		BBTL:       signals.BB.TL,
		BBML:       signals.BB.ML,
		BBBL:       signals.BB.BL,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal bolinger band error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.SMA,
		SignalValue: signals.SMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal sma error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.EMA,
		SignalValue: signals.EMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal ema error: %v", err)
		return err
	}
	_, err = s.db.SaveSignal(models.Signal{
		N:           n,
		BinSize:     bin,
		Timestamp:   timestamp,
		SignalType:  models.WMA,
		SignalValue: signals.WMA,
	})
	if err != nil {
		s.log.Debugf("saveSignals db.SaveSignal wma error: %v", err)
		return err
	}
	return nil
}
