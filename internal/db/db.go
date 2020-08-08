package db

import (
	"database/sql"
	nurl "net/url"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // comment
	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	migrateDb "github.com/tagirmukail/tccbot-backend/internal/db/migrate-db"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex"
)

const defaultRetryCount = 5

type DatabaseManager interface {
	Close() error

	// Candles
	//SaveCandle(candle *models.Candle) (int64, error)
	//GetCandlesByTimestamp(from, to time.Time) ([]*models.Candle, error)
	//GetPreviousCandlesByTheme(theme types.Theme, from, to time.Time, limit int) ([]*models.Candle, error)

	// Signals
	SaveSignal(data models.Signal) (int64, error)
	//GetSignalsByBinSize(binSize models.BinSize) ([]*models.Signal, error)
	//GetLastSignals(binSize models.BinSize, count int) ([]*models.Signal, error)
	GetSignalsByTS(signalType models.SignalType, binSize models.BinSize, ts []time.Time) ([]*models.Signal, error)
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
	cfg *config.GlobalConfig, db *sqlx.DB, log *logrus.Logger, command migrateDb.Command, step int,
) (*DB, error) {
	var err error
	manager := &DB{
		log:   log,
		retry: defaultRetryCount,
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "sqlite3://tccbot_db?x-migrations-table=schema_migrations"
	}

	var dbfile string
	if db == nil {
		purl, err := nurl.Parse(cfg.DBPath)
		if err != nil {
			return nil, err
		}
		dbfile = strings.Replace(migrate.FilterCustomQuery(purl).String(), "sqlite3://", "", 1)
		db, err = sqlx.Open("sqlite3", dbfile)
		if err != nil {
			return nil, err
		}
	}

	manager.db = db
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = migrateDb.Migrate(cfg.DBPath, command, step)
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
