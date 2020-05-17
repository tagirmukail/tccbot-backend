package migrate_db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Command string

const (
	UP   Command = "up"
	DOWN Command = "down"
	STEP Command = "step"
)

func Migrate(dbPath string, db *sql.DB, command Command, step int) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", dbPath, driver)
	if err != nil {
		return err
	}
	switch command {
	case UP:
		return m.Up()
	case DOWN:
		return m.Down()
	case STEP:
		return m.Steps(step)
	default:
		return errors.New("unknown command")
	}
}
