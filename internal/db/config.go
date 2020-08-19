package db

import (
	"database/sql"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
	"github.com/tagirmukail/tccbot-backend/internal/db/models"
)

func saveGlobalConfig(tx *sqlx.Tx, config models.GlobalConfig) (int, error) {
	var globalConfigs = make([]models.GlobalConfig, 0)
	err := tx.Select(&globalConfigs, `
	SELECT 
       id, db_path, order_proc_period_in_sec, created_at, updated_at 
	FROM global_config ORDER BY created_at desc`)
	if err != nil {
		return 0, err
	}

	if len(globalConfigs) == 0 {
		res := tx.MustExec(`INSERT INTO global_config(
                          	db_path, order_proc_period_in_sec, created_at, updated_at
                          )
							VALUES ($1, $2, $3, $4)`,
			config.DBPath, config.OrderProcPeriodInSec, config.CreatedAt, config.CreatedAt)

		globalConfigID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		return int(globalConfigID), nil
	}

	lastGlobalConfig := globalConfigs[0]
	if len(globalConfigs) > 1 {
		_, err = tx.Exec(`DELETE FROM global_config WHERE id!=$1`, lastGlobalConfig.ID)
		if err != nil {
			return 0, err
		}
	}

	res := tx.MustExec(`UPDATE global_config SET db_path=$1, order_proc_period_in_sec=$2, updated_at=$3 
		WHERE id=$4`,
		config.DBPath, config.OrderProcPeriodInSec, config.UpdatedAt, lastGlobalConfig.ID)
	_, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastGlobalConfig.ID, nil
}

func saveAdmin(tx *sqlx.Tx, globalConfigID int, admin models.Admin) error {
	var admins = make([]models.Admin, 0)
	err := tx.Select(&admins, `SELECT * FROM admin WHERE global_id=$1 LIMIT 1`, globalConfigID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if len(admins) == 0 {
		res := tx.MustExec(`INSERT INTO admin(
                  exchange, username, secret_token, token, global_id
                  ) VALUES ($1, $2, $3, $4, $5)`,
			admin.Exchange, admin.Username, admin.SecretToken, admin.Token, globalConfigID)
		var adminID int64
		adminID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		res = tx.MustExec(`UPDATE global_config SET admin_id=$1 WHERE id=$2`, adminID, globalConfigID)
		_, err = res.LastInsertId()
		if err != nil {
			return err
		}

		return nil
	}

	res := tx.MustExec(`UPDATE admin SET exchange=$1, username=$2, secret_token=$3, token=$4 WHERE global_id=$5`,
		admin.Exchange, admin.Username, admin.SecretToken, admin.Token, globalConfigID)
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func saveExchangeAPISettings(tx *sqlx.Tx, globalConfigID int, settings models.ExchangeAPISettings) error {
	var exchangeAPISettings = make([]models.ExchangeAPISettings, 0)
	err := tx.Select(&exchangeAPISettings, `SELECT * FROM exchanges_api_settings WHERE global_id=$1`,
		globalConfigID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if len(exchangeAPISettings) == 0 {
		res := tx.MustExec(`INSERT INTO exchanges_api_settings(
                                   exchange,
                                   enable,
                                   test,
                                   ping_sec,
                                   timeout_sec,
                                   retry_sec,
                                   buffer_size,
                                   currency,
                                   symbol,
                                   order_type,
                                   max_amount,
                                   limit_contracts_cnt,
                                   sell_order_coef,
                                   buy_order_coef,
                                   global_id
							) VALUES (
          							$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
							)`,
			settings.Exchange,
			settings.Enable,
			settings.Test,
			settings.PingSec,
			settings.TimeoutSec,
			settings.RetrySec,
			settings.BufferSize,
			settings.Currency,
			settings.Symbol,
			settings.OrderType,
			settings.MaxAmount,
			settings.LimitContractsCount,
			settings.SellOrderCoef,
			settings.BuyOrderCoef,
			globalConfigID)
		_, err := res.LastInsertId()
		if err != nil {
			return err
		}

		return nil
	}

	res := tx.MustExec(`UPDATE exchanges_api_settings SET exchange=$1, enable=$2, test=$3, ping_sec=$4, 
							timeout_sec=$5, retry_sec=$6, buffer_size=$7, currency=$8, symbol=$9, order_type=$10,
							max_amount=$11, limit_contracts_cnt=$11, sell_order_coef=$12, buy_order_coef=$13
							WHERE global_id=$14`,
		settings.Exchange, settings.Enable, settings.Test, settings.PingSec,
		settings.TimeoutSec, settings.RetrySec, settings.BufferSize, settings.Currency,
		settings.Symbol, settings.OrderType, settings.MaxAmount, settings.LimitContractsCount,
		settings.SellOrderCoef, settings.BuyOrderCoef, globalConfigID)
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func saveExchangeAccess(tx *sqlx.Tx, globalConfigID int, settings models.ExchangeAccess) error {
	var exchangesAccess = make([]models.ExchangeAccess, 0)
	err := tx.Select(&exchangesAccess, `SELECT * FROM exchange_access WHERE global_id=$1`,
		globalConfigID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(exchangesAccess) == 0 {
		res := tx.MustExec(`INSERT INTO exchange_access(
                            exchange, test, key, secret, global_id
				) VALUES ($1, $2, $3, $4, $5)`,
			settings.Exchange,
			settings.Test,
			settings.Key,
			settings.Secret,
			globalConfigID,
		)
		_, err := res.LastInsertId()
		if err != nil {
			return err
		}

		return nil
	}

	res := tx.MustExec(`UPDATE exchange_access SET exchange=$1, test=$2, "key"=$3, secret=$4 
 		WHERE global_id=$5`,
		settings.Exchange, settings.Test, settings.Key, settings.Secret, globalConfigID)
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func saveScheduler(tx *sqlx.Tx, globalConfigID int, settings models.Scheduler) error {
	var schedulers = make([]models.Scheduler, 0)
	err := tx.Select(&schedulers, `SELECT * FROM scheduler WHERE global_id=$1`,
		globalConfigID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(schedulers) == 0 {
		res := tx.MustExec(`INSERT INTO scheduler(
                      "type", enable, price_trailing, 
                      profit_close_btc, loss_close_btc, 
                      profit_pnl_diff, loss_pnl_diff, 
                      global_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			settings.Type,
			settings.Enable,
			settings.PriceTrailing,
			settings.ProfitCloseBTC,
			settings.LossCloseBTC,
			settings.ProfitPnlDiff,
			settings.LossPnlDiff,
			globalConfigID,
		)
		_, err := res.LastInsertId()
		if err != nil {
			return err
		}

		return nil
	}

	res := tx.MustExec(`UPDATE scheduler SET  "type"=$1, enable=$2, price_trailing=$3, profit_close_btc=$4, 
                     loss_close_btc=$5, profit_pnl_diff=$6, loss_pnl_diff=$7 WHERE global_id=$8`,
		settings.Type,
		settings.Enable,
		settings.PriceTrailing,
		settings.ProfitCloseBTC,
		settings.LossCloseBTC,
		settings.ProfitPnlDiff,
		settings.LossPnlDiff,
		globalConfigID)
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

//nolint:funlen
func saveStrategies(tx *sqlx.Tx, globalConfigID int, settings models.StrategiesConfig) error {
	var strategies = make([]models.StrategiesConfig, 0)
	err := tx.Select(&strategies, `SELECT * FROM strategies_config WHERE global_id=$1`,
		globalConfigID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(strategies) == 0 {
		res := tx.MustExec(`INSERT INTO strategies_config(
                              bin,
                              enable_rsi_bb,
                              retry_process_count,
                              get_candles_count,
                              trend_filter_enable,
                              candles_filter_enable,
                              max_filter_trend_count,
                              max_candles_filter_count,
                              bb_last_candles_count,
                              rsi_count,
                              rsi_min_border,
                              rsi_max_border,
                              rsi_trade_coef,
                              macd_fast_count,
                              macd_slow_count,
                              macd_sig_count,
                              global_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
			settings.Bin.String(),
			settings.EnableRSIBB,
			settings.RetryProcessCount,
			settings.GetCandlesCount,
			settings.TrendFilterEnable,
			settings.CandlesFilterEnable,
			settings.MaxFilterTrendCount,
			settings.MaxCandlesFilterCount,
			settings.BBLastCandlesCount,
			settings.RsiCount,
			settings.RsiMinBorder,
			settings.RsiMaxBorder,
			settings.RsiTradeCoef,
			settings.MacdFastCount,
			settings.MacdSlowCount,
			settings.MacdSigCount,
			globalConfigID)
		_, err := res.LastInsertId()
		if err != nil {
			return err
		}

		return nil
	}

	res := tx.MustExec(`UPDATE strategies_config SET 
                              bin=$1,
                              enable_rsi_bb=$2,
                              retry_process_count=$3,
                              get_candles_count=$4,
                              trend_filter_enable=$5,
                              candles_filter_enable=$6,
                              max_filter_trend_count=$7,
                              max_candles_filter_count=$8,
                              bb_last_candles_count=$9,
                              rsi_count=$10,
                              rsi_min_border=$11,
                              rsi_max_border=$12,
                              rsi_trade_coef=$13,
                              macd_fast_count=$14,
                              macd_slow_count=$15,
                              macd_sig_count=$16 WHERE global_id=$17`,
		settings.Bin.String(),
		settings.EnableRSIBB,
		settings.RetryProcessCount,
		settings.GetCandlesCount,
		settings.TrendFilterEnable,
		settings.CandlesFilterEnable,
		settings.MaxFilterTrendCount,
		settings.MaxCandlesFilterCount,
		settings.BBLastCandlesCount,
		settings.RsiCount,
		settings.RsiMinBorder,
		settings.RsiMaxBorder,
		settings.RsiTradeCoef,
		settings.MacdFastCount,
		settings.MacdSlowCount,
		settings.MacdSigCount,
		globalConfigID)
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) saveConfiguration(config models.GlobalConfig) error {
	tx, err := db.db.Beginx()
	if err != nil {
		return err
	}
	defer func(tx *sqlx.Tx, err error) {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				db.log.Errorf("save configuration rollback tx failed: %v", err)
			}
			return
		}
		txErr := tx.Commit()
		if txErr != nil {
			db.log.Errorf("save configuration commit tx failed: %v", err)
		}
	}(tx, err)

	var globalConfigID int
	globalConfigID, err = saveGlobalConfig(tx, config)
	if err != nil {
		return err
	}

	err = saveAdmin(tx, globalConfigID, config.Admin)
	if err != nil {
		return err
	}

	err = saveExchangeAPISettings(tx, globalConfigID, config.ExchangeAPISettings)
	if err != nil {
		return err
	}

	err = saveExchangeAccess(tx, globalConfigID, config.ExchangeAccess)
	if err != nil {
		return err
	}

	err = saveScheduler(tx, globalConfigID, config.Scheduler)
	if err != nil {
		return err
	}

	err = saveStrategies(tx, globalConfigID, config.StrategiesConfig)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) SaveConfiguration(config models.GlobalConfig) error {
	for i := 0; i < 10; i++ {
		err := db.saveConfiguration(config)
		if err != nil {
			if err == sqlite3.ErrBusy {
				time.Sleep(350 * time.Millisecond)
				continue
			}
			return err
		}

		break
	}

	return nil
}

func (db *DB) GetConfiguration() (cfg *models.GlobalConfig, err error) {
	for i := 0; i < 10; i++ {
		cfg, err = db.getConfiguration()
		if err != nil {
			if err == sqlite3.ErrBusy {
				time.Sleep(350 * time.Millisecond)
				continue
			}

			return nil, err
		}

		break
	}

	return cfg, nil
}

func (db *DB) getConfiguration() (*models.GlobalConfig, error) {
	var globalConfig models.GlobalConfig
	err := db.db.QueryRow(`
	SELECT 
	   gc.id, gc.db_path, gc.order_proc_period_in_sec, gc.created_at, gc.updated_at,

       a.id, a.exchange, a.username, a.secret_token, a.token,

       eas.id, eas.exchange, eas.enable, eas.test, eas.ping_sec, eas.timeout_sec, eas.retry_sec, eas.buffer_size,
       eas.currency, eas.symbol, eas.order_type, eas.max_amount, eas.limit_contracts_cnt, eas.sell_order_coef,
       eas.buy_order_coef,

       ea.id, ea.exchange, ea.test, ea.key, ea.secret,

       s.id, s.type, s.enable, s.price_trailing, s.profit_close_btc, s.loss_close_btc, s.profit_pnl_diff,
       s.loss_pnl_diff,

       sc.id, sc.bin, sc.enable_rsi_bb, sc.retry_process_count, sc.get_candles_count, sc.trend_filter_enable,
       sc.candles_filter_enable, sc.max_filter_trend_count, sc.max_candles_filter_count, sc.bb_last_candles_count,
       sc.rsi_count, sc.rsi_min_border, sc.rsi_max_border, sc.rsi_trade_coef, sc.macd_fast_count, sc.macd_slow_count,
       sc.macd_sig_count
	FROM global_config as gc
		CROSS JOIN  admin as a ON gc.admin_id=a.id AND a.global_id=gc.id
		CROSS JOIN exchanges_api_settings as eas ON eas.global_id=gc.id
		CROSS JOIN exchange_access as ea ON ea.global_id=gc.id
		CROSS JOIN scheduler as s ON s.global_id=gc.id
		CROSS JOIN strategies_config as sc ON sc.global_id=gc.id LIMIT 1`).Scan(
		&globalConfig.ID, &globalConfig.DBPath, &globalConfig.OrderProcPeriodInSec, &globalConfig.CreatedAt,
		&globalConfig.UpdatedAt,

		&globalConfig.Admin.ID, &globalConfig.Admin.Exchange, &globalConfig.Admin.Username,
		&globalConfig.Admin.SecretToken,
		&globalConfig.Admin.Token,

		&globalConfig.ExchangeAPISettings.ID, &globalConfig.ExchangeAPISettings.Exchange,
		&globalConfig.ExchangeAPISettings.Enable, &globalConfig.ExchangeAPISettings.Test,
		&globalConfig.ExchangeAPISettings.PingSec, &globalConfig.ExchangeAPISettings.TimeoutSec,
		&globalConfig.ExchangeAPISettings.RetrySec, &globalConfig.ExchangeAPISettings.BufferSize,
		&globalConfig.ExchangeAPISettings.Currency, &globalConfig.ExchangeAPISettings.Symbol,
		&globalConfig.ExchangeAPISettings.OrderType, &globalConfig.ExchangeAPISettings.MaxAmount,
		&globalConfig.ExchangeAPISettings.LimitContractsCount, &globalConfig.ExchangeAPISettings.SellOrderCoef,
		&globalConfig.ExchangeAPISettings.BuyOrderCoef,

		&globalConfig.ExchangeAccess.ID, &globalConfig.ExchangeAccess.Exchange, &globalConfig.ExchangeAccess.Test,
		&globalConfig.ExchangeAccess.Key, &globalConfig.ExchangeAccess.Secret,

		&globalConfig.Scheduler.ID, &globalConfig.Scheduler.Type, &globalConfig.Scheduler.Enable,
		&globalConfig.Scheduler.PriceTrailing, &globalConfig.Scheduler.ProfitCloseBTC,
		&globalConfig.Scheduler.LossCloseBTC, &globalConfig.Scheduler.ProfitPnlDiff,
		&globalConfig.Scheduler.LossPnlDiff,

		&globalConfig.StrategiesConfig.ID, &globalConfig.StrategiesConfig.Bin,
		&globalConfig.StrategiesConfig.EnableRSIBB, &globalConfig.StrategiesConfig.RetryProcessCount,
		&globalConfig.StrategiesConfig.GetCandlesCount, &globalConfig.StrategiesConfig.TrendFilterEnable,
		&globalConfig.StrategiesConfig.CandlesFilterEnable, &globalConfig.StrategiesConfig.MaxFilterTrendCount,
		&globalConfig.StrategiesConfig.MaxCandlesFilterCount, &globalConfig.StrategiesConfig.BBLastCandlesCount,
		&globalConfig.StrategiesConfig.RsiCount, &globalConfig.StrategiesConfig.RsiMinBorder,
		&globalConfig.StrategiesConfig.RsiMaxBorder, &globalConfig.StrategiesConfig.RsiTradeCoef,
		&globalConfig.StrategiesConfig.MacdFastCount, &globalConfig.StrategiesConfig.MacdSlowCount,
		&globalConfig.StrategiesConfig.MacdSigCount,
	)
	if err != nil {
		return nil, err
	}

	return &globalConfig, nil
}
