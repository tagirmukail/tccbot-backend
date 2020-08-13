package db

import (
	"errors"
	"runtime/debug"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

func (db *DB) SaveSignal(data models.Signal) (lastID int64, err error) {
	for i := 0; i < db.retry; i++ {
		lastID, err = db.saveSignal(data)
		if err == sqlite3.ErrBusy {
			db.log.Warnf("db is busy, wait to next")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

func (db *DB) saveSignal(data models.Signal) (int64, error) {
	defer func() {
		r := recover()
		if r != nil {
			db.log.Errorf("saveSignal panic: %v", r)
			debug.PrintStack()
		}
	}()
	var (
		createdAt    = time.Now().Unix()
		lastID       int64
		existSignals []*models.Signal
	)
	err := sqlx.Select(db.ql, &existSignals,
		`SELECT id FROM signals WHERE signal_t=$1 AND bin=$2 AND timestamp=$3`,
		data.SignalType.String(),
		data.BinSize.String(),
		data.Timestamp,
	)
	if err != nil {
		return 0, err
	}
	for i, sig := range existSignals {
		db.log.Debugf("saveSignal SQL SELECT RESULT[%d]:[%#v]", i, sig)
	}
	if len(existSignals) != 0 {
		return existSignals[0].ID, nil
	}

	data.CreatedAt = createdAt
	data.UpdatedAt = createdAt
	result := sqlx.MustExec(db.ql, `INSERT INTO signals(
                    n,
                    macd_fast,
                    macd_slow,
                    macd_sig,
                    bin,
                    timestamp,
                    signal_t,
                    signal_v,
                    bbtl,
                    bbml,
                    bbbl,
                    macd_v,
                    macd_h_v,
                    created_at,
                    updated_at
	) VALUES (
	          $1, $2,$3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	)`,
		data.N,
		data.MACDFast,
		data.MACDSlow,
		data.MACDSig,
		data.BinSize.String(),
		data.Timestamp,
		data.SignalType.String(),
		data.SignalValue,
		data.BBTL,
		data.BBML,
		data.BBBL,
		data.MACDValue,
		data.MACDHistogramValue,
		data.CreatedAt,
		data.UpdatedAt,
	)
	lastID, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if lastID == 0 {
		return 0, errors.New("signal not inserted")
	}
	return lastID, nil
}

//func (db *DB) GetSignalsByBinSize(binSize models.BinSize) ([]*models.Signal, error) {
//	return nil, nil
//}
//
//func (db *DB) GetLastSignals(binSize models.BinSize, count int) ([]*models.Signal, error) {
//	return nil, nil
//}

func (db *DB) GetSignalsByTS(
	signalType models.SignalType, binSize models.BinSize, ts []time.Time,
) ([]*models.Signal, error) {
	var result = make([]*models.Signal, 0, len(ts)*2)

	if len(ts) == 0 {
		return result, nil
	}

	var tsArgs = make([]string, 0, len(ts))
	for _, t := range ts {
		tsArgs = append(tsArgs, t.Format("2006-01-02 15:04:05+00:00"))
	}

	query, args, err := sqlx.In(
		`SELECT * FROM signals WHERE signal_t=? AND bin=? AND timestamp IN (?)`,
		signalType.String(), binSize.String(), tsArgs)
	if err != nil {
		return nil, err
	}

	query = db.db.Rebind(query)
	err = sqlx.Select(db.ql, &result, query, args...)
	if err != nil {
		return nil, err
	}
	for i, sig := range result {
		db.log.Debugf("GetSignalsByTS SQL RESULT[%d]:[%#v]", i, sig)
	}
	return result, nil
}
