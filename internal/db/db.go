package db

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

type DBManager interface {
	Close() error

	// Candles
	//SaveCandle(candle *models.Candle) (int64, error)
	//GetCandlesByTimestamp(from, to time.Time) ([]*models.Candle, error)
	//GetPreviousCandlesByTheme(theme types.Theme, from, to time.Time, limit int) ([]*models.Candle, error)

	// Signals
	SaveSignal(data models.Signal) (int64, error)
	//GetSignalsByBinSize(binSize models.BinSize) ([]*models.Signal, error)
	//GetLastSignals(binSize models.BinSize, count int) ([]*models.Signal, error)
	GetSignalsByTs(binSize models.BinSize, ts []time.Time) ([]*models.Signal, error)
	// CandlesWithSignals
	//GetCandleWithSignals(id int64) (*models.CandleWithSignals, error)
	//GetLastCandlesWithSignals(theme types.Theme, n int, limit int) ([]*models.CandleWithSignals, error)
}

type DB struct {
	log *logrus.Logger
	db  *sqlx.DB
}

func New(cfg *config.GlobalConfig, db *sqlx.DB, log *logrus.Logger, command Command) (*DB, error) {
	var err error
	manager := &DB{
		log: log,
	}
	if db == nil {
		db, err = sqlx.Open(
			"pgx",
			fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%v sslmode=%s",
				cfg.DB.DBName, cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.SSLMode),
		)
		if err != nil {
			return nil, err
		}
	}

	manager.db = db

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = migration(cfg, log, command)
	if err != nil {
		return nil, err
	}

	log.Info("database initialized")

	return manager, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
