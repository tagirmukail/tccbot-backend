package migrate_db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tagirmukail/tccbot-backend/migrations"
)

type Command string

const (
	UP   Command = "up"
	DOWN Command = "down"
	STEP Command = "step"
)

func Migrate(dbPath string, db *sql.DB, command Command, step int) error {
	// driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	// if err != nil {
	// 	return err
	// }

	s := bindata.Resource(migrations.AssetNames(), func(name string) ([]byte, error) {
		return migrations.Asset(name)
	})
	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", d, dbPath)
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
