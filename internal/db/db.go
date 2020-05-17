package db

import (
	"time"

	"github.com/golang-migrate/migrate/v4"

	migrate_db "github.com/tagirmukail/tccbot-backend/internal/db/migrate-db"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const defaultRetryCount = 5

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
	GetSignalsByTs(signalTypes []models.SignalType, binSizes []models.BinSize, ts []time.Time) ([]*models.Signal, error)
	// CandlesWithSignals
	//GetCandleWithSignals(id int64) (*models.CandleWithSignals, error)
	//GetLastCandlesWithSignals(theme types.Theme, n int, limit int) ([]*models.CandleWithSignals, error)
	SaveOrder(ord bitmex.OrderCopied) (*models.Order, error)
}

type DB struct {
	log   *logrus.Logger
	retry int
	db    *sqlx.DB
}

func NewDB(
	cfg *config.GlobalConfig, db *sqlx.DB, log *logrus.Logger, command migrate_db.Command, step int,
) (*DB, error) {
	var err error
	manager := &DB{
		log:   log,
		retry: defaultRetryCount,
	}
	if cfg.DBPath == "" {
		cfg.DBPath = ".tccbot_db"
	}
	if db == nil {
		db, err = sqlx.Open("sqlite3", cfg.DB.DBName)
		if err != nil {
			return nil, err
		}
	}
	manager.db = db
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = migrate_db.Migrate(cfg.DBPath, manager.db.DB, command, step)
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return manager, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
