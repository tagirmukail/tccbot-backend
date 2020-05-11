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
					signal_t   signal_type,
					n          int,
					macd_fast  int,
					macd_slow  int,
					macd_sig   int,
					timestamp  timestamp   NOT NULL,
					signal_v   float4,
					bbtl 	   float4,
					bbml	   float4,
					bbbl 	   float4,
					macd_h_v   float4,
					created_at bigint NOT NULL,
    				updated_at bigint NOT NULL
				)
			`)
			if err != nil {
				log.Error(err)
				return err
			}
			_, err = db.ExecOne(
				`CREATE UNIQUE INDEX ts_bin_signal_t_idx ON signals (timestamp, bin, signal_t)`,
			)
			if err != nil {
				log.Error(err)
				return err
			}

			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping signals table")
			_, err := db.Exec(`DROP TABLE signals`)
			return err
		})
}
