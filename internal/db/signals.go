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
                    bin,
                    timestamp,
                    n,
                    sma,
                    wma,
                    ema,
                    bbtl,
                    bbml,
                    bbbl,
                    created_at,
                    updated_at
	) VALUES (
	          $1, $2,$3, $4, $5, $6, $7, $8, $9, $10, $11
	) RETURNING id`,
		data.BinSize.String(),
		data.Timestamp,
		data.N,
		data.SMA,
		data.WMA,
		data.EMA,
		data.BBTL,
		data.BBML,
		data.BBBL,
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

func (db *DB) GetSignalsByBinSize(binSize models.BinSize) ([]*models.Signal, error) {
	return nil, nil
}

func (db *DB) GetLastSignals(binSize models.BinSize, count int) ([]*models.Signal, error) {
	return nil, nil
}

func (db *DB) GetLastSignalsByTs(binSize models.BinSize, ts []time.Time) ([]*models.Signal, error) {
	var result = make([]*models.Signal, 0, len(ts)*2)

	if len(ts) == 0 {
		return result, nil
	}
	// TODO проблема с декодом bin_size в структуру сигнала
	query, args, err := sqlx.In(`SELECT * FROM signals WHERE timestamp IN (?)`, ts)
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
