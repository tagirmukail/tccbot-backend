package migrations

import (
	"github.com/go-pg/migrations/v7"
	"github.com/sirupsen/logrus"
)

func RegisterCandleTable(collections *migrations.Collection, log *logrus.Logger) {
	collections.MustRegisterTx(
		func(db migrations.DB) error {
			log.Info("create candles table")
			_, err := db.Exec(`
				CREATE TABLE candles(
					id         serial PRIMARY KEY,
					theme 	   varchar(80) NOT NULL,
					symbol     varchar(80) NOT NULL,
					timestamp  timestamp   NOT NULL,
					open       float4 NOT NULL,
					high	   float4 NOT NULL,
					low		   float4 NOT NULL,
					close	   float4 NOT NULL,
					trades	   int,
					volume	   int,
					vwap       float4 NOT NULL,
					lastsize   int,
					turnover   bigint NOT NULL,
					homenotional float4,
					foreignnotional float4,
    				created_at bigint NOT NULL,
    				updated_at bigint NOT NULL
				)
			`)
			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping candles table")
			_, err := db.Exec(`DROP TABLE candles`)
			return err
		})
}
