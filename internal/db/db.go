package db

import (
	"database/sql"
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
	GetSignalsByTs(signalType models.SignalType, binSize models.BinSize, ts []time.Time) ([]*models.Signal, error)
	// CandlesWithSignals
	//GetCandleWithSignals(id int64) (*models.CandleWithSignals, error)
	//GetLastCandlesWithSignals(theme types.Theme, n int, limit int) ([]*models.CandleWithSignals, error)
	SaveOrder(ord bitmex.OrderCopied) (*models.Order, error)
}

type QueryLogger struct {
	queryer sqlx.Queryer
	execer  sqlx.Execer
	logger  *logrus.Logger
}

func (p *QueryLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	p.logger.Debugf("SQL QUERY:%s;ARGS:%v", query, args)
	return p.queryer.Query(query, args...)
}

func (p *QueryLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	p.logger.Debugf("SQL QUERY:%s;ARGS:%v", query, args)
	return p.queryer.Queryx(query, args...)
}

func (p *QueryLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	p.logger.Debugf("SQL QUERY:%s;ARGS:%v", query, args)
	return p.queryer.QueryRowx(query, args...)
}

func (p *QueryLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	p.logger.Debugf("SQL ExecQUERY:%s;ARGS:%v", query, args)
	return p.execer.Exec(query, args...)
}

type DB struct {
	log   *logrus.Logger
	retry int
	db    *sqlx.DB
	ql    *QueryLogger
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

	manager.ql = &QueryLogger{
		queryer: db,
		execer:  db,
		logger:  log,
	}
	return manager, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
