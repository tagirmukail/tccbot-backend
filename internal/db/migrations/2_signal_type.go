package migrations

import (
	"github.com/go-pg/migrations/v7"
	"github.com/sirupsen/logrus"
)

func RegisterSignalTypeEnum(collections *migrations.Collection, log *logrus.Logger) {
	collections.MustRegisterTx(
		func(db migrations.DB) error {
			log.Info("create signal_type enum")
			_, err := db.Exec(`
				CREATE TYPE signal_type AS ENUM (
					'Empty',
					'SMA',
					'EMA',
					'WMA',
					'RSI',
					'BolingerBand',
					'MACD'
				)
			`)
			if err != nil {
				log.Errorf("signal_type create error: %v", err)
			}
			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping signal_type enum")
			_, err := db.Exec(`DROP TYPE signal_type`)
			return err
		})
}
