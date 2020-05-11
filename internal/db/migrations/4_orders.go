package migrations

import (
	"github.com/go-pg/migrations/v7"
	"github.com/sirupsen/logrus"
)

func RegisterOrdersTable(collections *migrations.Collection, log *logrus.Logger) {
	collections.MustRegisterTx(
		func(db migrations.DB) error {
			log.Info("create orders table")
			_, err := db.Exec(`
				CREATE TABLE orders(
					id         serial PRIMARY KEY,
					account    bigint,
					avgPx      float4,
					clOrdID    varchar(400),
					clOrdLinkID varchar(400),
					contingencyType varchar(400),
					cumQty     bigint,
					currency   varchar(80),
					displayQty bigint,
					exDestination varchar(400),
					execInst varchar(400),
					leavesQty bigint,
					multiLegReportingType varchar(400),
					ordRejReason varchar(400),
					ordStatus varchar(400),
					ordType varchar(400),
					orderID varchar(400),
					orderQty bigint,
					pegOffsetValue float4,
					pegPriceType varchar(400),
					price        float4,
					settlCurrency varchar(400),
					side          varchar(80),
					simpleCumQty  float4,
					simpleLeavesQty float4,
					simpleOrderQty  float4,
					stopPx  float4,
					symbol varchar(80),
					text   varchar(800),
					timeInForce varchar(400),
					timestamp   timestamp,
					transactTime varchar(400),
					triggered     varchar(400),
					workingIndicator boolean,
    				created_at timestamp NOT NULL,
    				updated_at timestamp
				)
			`)
			if err != nil {
				log.Errorf("orders table create error:%v", err)
			}
			return err
		},
		func(db migrations.DB) error {
			log.Info("dropping orders table")
			_, err := db.Exec(`DROP TABLE orders`)
			return err
		})
}
