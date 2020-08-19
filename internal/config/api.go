package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type ExchangesSettings struct {
	Bitmex APISettings
}

type ExchangeSettings struct {
	Enable bool
	API    APISettings
}

type APISettings struct {
	Test                bool
	PingSec             int
	TimeoutSec          int
	RetrySec            int
	BufferSize          int
	Currency            string
	Symbol              string
	OrderType           types.OrderType
	MaxAmount           float64
	ClosePositionMinBTC float64
	LimitContractsCount int
	SellOrderCoef       float64
	BuyOrderCoef        float64
}

type ExchangesAccess struct {
	Bitmex Access `json:"bitmex"`
}

type Access struct {
	Key     string `json:"key"`
	Secret  string `json:"secret"`
	Testnet struct {
		Key    string
		Secret string
	}
}

func initExchangesAPI() ExchangesSettings {
	var bitmex APISettings
	bitmexSettings := viper.GetStringMap("exchanges_settings.bitmex")
	if len(bitmexSettings) == 0 {
		// default
		bitmex = APISettings{
			Test:                true,
			PingSec:             20,
			TimeoutSec:          30,
			RetrySec:            5,
			BufferSize:          10,
			Currency:            "XBt",
			Symbol:              "XBTUSD",
			OrderType:           types.Limit,
			MaxAmount:           130,
			ClosePositionMinBTC: 0.0005,
			LimitContractsCount: 300,
			BuyOrderCoef:        0.2,
			SellOrderCoef:       0.1,
		}
	} else {
		bitmex = APISettings{
			Test:                viper.GetBool("exchanges_settings.bitmex.test"),
			PingSec:             viper.GetInt("exchanges_settings.bitmex.ping_sec"),
			TimeoutSec:          viper.GetInt("exchanges_settings.bitmex.timeout_sec"),
			RetrySec:            viper.GetInt("exchanges_settings.bitmex.retry_sec"),
			BufferSize:          viper.GetInt("exchanges_settings.bitmex.buffer_size"),
			Symbol:              viper.GetString("exchanges_settings.bitmex.symbol"),
			Currency:            viper.GetString("exchanges_settings.bitmex.currency"),
			OrderType:           types.OrderType(viper.GetString("exchanges_settings.bitmex.order_type")),
			MaxAmount:           viper.GetFloat64("exchanges_settings.bitmex.max_amount"),
			ClosePositionMinBTC: viper.GetFloat64("exchanges_settings.bitmex.close_position_min_btc"),
			LimitContractsCount: viper.GetInt("exchanges_settings.bitmex.limit_contracts_cnt"),
			BuyOrderCoef:        viper.GetFloat64("exchanges_settings.bitmex.buy_order_coef"),
			SellOrderCoef:       viper.GetFloat64("exchanges_settings.bitmex.sell_order_coef"),
		}
	}
	fmt.Println("--------------------------------------------")
	fmt.Printf("bitmex settings: %#v\n", bitmex)
	fmt.Println("--------------------------------------------")

	return ExchangesSettings{
		Bitmex: bitmex,
	}
}

func initExchangesAccesses() ExchangesAccess {
	return ExchangesAccess{
		Bitmex: Access{
			Key:    viper.GetString("exchanges_access.bitmex.key"),
			Secret: viper.GetString("exchanges_access.bitmex.secret"),
			Testnet: struct {
				Key    string
				Secret string
			}{
				Key:    viper.GetString("exchanges_access.bitmex.testnet.key"),
				Secret: viper.GetString("exchanges_access.bitmex.testnet.secret"),
			},
		},
	}
}
