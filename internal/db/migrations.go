package db

import (
	"fmt"
	"strings"

	"github.com/go-pg/migrations/v7"
	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"
	"github.com/tagirmukail/tccbot-backend/internal/config"
	dbmigrations "github.com/tagirmukail/tccbot-backend/internal/db/migrations"
)

type Command string

const (
	INIT        Command = "init"
	UP          Command = "up"
	DOWN        Command = "down"
	RESET       Command = "reset"
	VERSION     Command = "version"
	SET_VERSION Command = "set_version"

	gopgMigrationErrStr = "gopg_migrations"
)

func migration(cfg *config.GlobalConfig, log *logrus.Logger, command Command) error {
	db := pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port),
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Database: cfg.DB.DBName,
	})
	defer db.Close()

	_, err := migrations.Version(db)
	if err != nil {
		if strings.Contains(err.Error(), gopgMigrationErrStr) {
			log.Warn(err)
			_, _, errRun := migrations.Run(db, "init")
			if errRun != nil {
				return errRun
			}

			return nil
		}
	}

	cols := migrations.NewCollection()
	registerAllTables(cols, log, cfg)

	oldV, newV, err := cols.Run(db, string(command))
	if err != nil {
		if strings.Contains(err.Error(), gopgMigrationErrStr) {
			_, _, errRun := migrations.Run(db, "init")
			if errRun != nil {
				return errRun
			}

			return err
		}
	}

	if oldV != newV {
		log.Infof("database migrated from %d to %d version", oldV, newV)
	}
	log.Infof("database version: %d", newV)

	return nil
}

func registerAllTables(collection *migrations.Collection, log *logrus.Logger, cfg *config.GlobalConfig) {
	dbmigrations.RegisterBinSizeEnum(collection, log)
	dbmigrations.RegisterSignalTypeEnum(collection, log)
	dbmigrations.RegisterSignalTable(collection, log)
}
