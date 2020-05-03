package migrations

import (
	"github.com/go-pg/migrations/v7"
	"github.com/sirupsen/logrus"
)

func RegisterSignalTable(collections *migrations.Collection, log *logrus.Logger) {
	collections.MustRegisterTx(
		func(db migrations.DB) error {
			log.Info("create signals table")
			_, err := db.Exec(`
				CREATE TABLE signals(
					id         serial PRIMARY KEY,
					bin        bin_size,
					n          int,
					timestamp  timestamp   NOT NULL,
					sma        float4,
					wma        float4,
					ema        float4,
					bbtl 	   float4,
					bbml	   float4,
					bbbl 	   float4,
					created_at bigint NOT NULL,
    				updated_at bigint NOT NULL
				)
			`)
			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping signals table")
			_, err := db.Exec(`DROP TABLE signals`)
			return err
		})
}
