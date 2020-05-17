package db

import (
	"errors"
	"strings"
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
	var (
		createdAt    = time.Now().Unix()
		lastID       int64
		existSignals []*models.Signal
	)
	err := db.db.Select(&existSignals,
		`SELECT id FROM signals WHERE signal_t=$1 AND bin=$2 AND timestamp=$3`,
		data.SignalType.String(),
		data.BinSize.String(),
		data.Timestamp,
	)
	if err != nil {
		return 0, err
	}
	if len(existSignals) != 0 {
		return existSignals[0].ID, nil
	}

	data.CreatedAt = createdAt
	data.UpdatedAt = createdAt
	result, err := db.db.Exec(`INSERT INTO signals(
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
	if err != nil {
		return 0, err
	}
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

func (db *DB) GetSignalsByTs(
	signalTypes []models.SignalType, binSizes []models.BinSize, ts []time.Time,
) ([]*models.Signal, error) {
	var result = make([]*models.Signal, 0, len(ts)*2)

	if len(ts) == 0 {
		return result, nil
	}

	var sigQueryBuild = strings.Builder{}
	for i, sigT := range signalTypes {
		sigQueryBuild.WriteString("'")
		sigQueryBuild.WriteString(sigT.String())
		sigQueryBuild.WriteString("'")
		if i != len(signalTypes)-1 {
			sigQueryBuild.WriteString(",")
		}
	}
	var binQueryBuild = strings.Builder{}
	for i, binS := range binSizes {
		binQueryBuild.WriteString("'")
		binQueryBuild.WriteString(binS.String())
		binQueryBuild.WriteString("'")
		if i != len(binSizes)-1 {
			binQueryBuild.WriteString(",")
		}
	}

	query, args, err := sqlx.In(
		`SELECT * FROM signals WHERE signal_t IN (?) AND bin IN (?) AND timestamp IN (?)`,
		sigQueryBuild.String(), binQueryBuild.String(),
		ts)
	if err != nil {
		return nil, err
	}

	query = db.db.Rebind(query)
	err = db.db.Select(&result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}
