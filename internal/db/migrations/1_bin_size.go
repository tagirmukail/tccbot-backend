package migrations

import (
	"github.com/go-pg/migrations/v7"
	"github.com/sirupsen/logrus"
)

func RegisterBinSizeEnum(collections *migrations.Collection, log *logrus.Logger) {
	collections.MustRegisterTx(
		func(db migrations.DB) error {
			log.Info("create bin_size enum")
			_, err := db.Exec(`
				CREATE TYPE bin_size AS ENUM (
					'1m',
					'5m',
					'1h',
					'1d'
				)
			`)
			if err != nil {
				log.Errorf("bin size create error: %v", err)
			}
			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping bin_size enum")
			_, err := db.Exec(`DROP TYPE bin_size`)
			return err
		})
}
