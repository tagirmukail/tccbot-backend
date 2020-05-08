package db

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

func (db *DB) SaveSignal(data models.Signal) (int64, error) {
	var (
		createdAt = time.Now().Unix()
		lastID    int64
	)
	data.CreatedAt = createdAt
	data.UpdatedAt = createdAt
	err := db.db.QueryRow(`INSERT INTO signals(
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
                    macd_h_v,
                    created_at,
                    updated_at
	) VALUES (
	          $1, $2,$3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	) RETURNING id`,
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
		data.MACDHistogramValue,
		data.CreatedAt,
		data.UpdatedAt,
	).Scan(&lastID)
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

func (db *DB) GetSignalsByTs(binSize models.BinSize, ts []time.Time) ([]*models.Signal, error) {
	var result = make([]*models.Signal, 0, len(ts)*2)

	if len(ts) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(`SELECT * FROM signals WHERE bin=? AND timestamp IN (?)`, binSize, ts)
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
