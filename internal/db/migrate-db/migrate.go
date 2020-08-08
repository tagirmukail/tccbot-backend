package migratedb

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // migrate
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/mattn/go-sqlite3" // migrate

	"github.com/tagirmukail/tccbot-backend/migrations"
)

type Command string

const (
	UP   Command = "up"
	DOWN Command = "down"
	STEP Command = "step"
)

func Migrate(dbPath string, command Command, step int) error {
	s := bindata.Resource(migrations.AssetNames(), migrations.Asset)
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
