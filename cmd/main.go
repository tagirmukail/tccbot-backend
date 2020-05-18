package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"

	migrate_db "github.com/tagirmukail/tccbot-backend/internal/db/migrate-db"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"
	"github.com/tagirmukail/tccbot-backend/internal/db"

	"github.com/tagirmukail/tccbot-backend/internal/strategies/strategy"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/config"
	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	"github.com/tagirmukail/tccbot-backend/internal/strategies"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/internal/utils/logger"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws"
)

const maxCandles = 100

// TODO перенести скрипты миграций в го файлы, чтоб не тащить повсюду за собой

func main() {
	var (
		configPath       string
		logLevel         uint
		migrationCommand string
		testMode         bool
		prof             string
		logDir           string
		initSignals      bool
		migrationOnly    bool
		step             int
	)

	flag.StringVar(&prof, "prof", "", "file name for profiling")
	flag.StringVar(&configPath, "config", "config-yaml/config-local.yaml", "configuration file")
	flag.UintVar(&logLevel, "level", 4, "log level")
	flag.BoolVar(&migrationOnly, "onlymigration", false, "run only migrations")
	flag.StringVar(
		&migrationCommand,
		"migration",
		"up",
		"database migration command:init, up, down, reset, version, set_version")
	flag.IntVar(&step, "step", 0, "migration step")
	flag.BoolVar(&testMode, "test", false, "Use exchanges to test mode")
	flag.StringVar(&logDir, "logdir", "", "logs save directory")
	flag.BoolVar(&initSignals, "siginit", false, "initialization previous signals.By default disabled")
	flag.Parse()

	if prof != "" {
		f, err := os.Create(prof)
		if err != nil {
			println(err)
			return
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			println(err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	log, err := logger.New(logLevel, logDir, &logrus.TextFormatter{DisableColors: false, ForceColors: true, FullTimestamp: true})
	if err != nil {
		logrus.Fatal(err)
	}
	log.Info("service tccbot started...")

	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	var dbManager db.DBManager
	dbManager, err = db.NewDB(cfg, nil, log, migrate_db.Command(migrationCommand), step)
	if err != nil {
		log.Fatalf("migartion failed: %v", err)
	}
	if migrationOnly {
		log.Info("migrations completed")
		return
	}

	var themes = cfg.GlobStrategies.GetThemes()

	tradeApi := tradeapi.NewTradeApi(
		cfg.Accesses.Bitmex.Key,
		cfg.Accesses.Bitmex.Secret,
		cfg.Accesses.Bitmex.Testnet.Key,
		cfg.Accesses.Bitmex.Testnet.Secret,
		log,
		testMode,
		ws.NewWS(
			log,
			testMode,
			cfg.ExchangesSettings.Bitmex.PingSec,
			cfg.ExchangesSettings.Bitmex.TimeoutSec,
			uint32(cfg.ExchangesSettings.Bitmex.RetrySec),
			themes,
			types.Symbol(cfg.ExchangesSettings.Bitmex.Symbol),
		),
	)

	ordProc := orderproc.New(tradeApi, cfg, log)

	caches := candlecache.NewBinToCache(
		cfg.GlobStrategies.GetBinSizes(), maxCandles, types.Symbol(cfg.ExchangesSettings.Bitmex.Symbol), log,
	)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	wg := &sync.WaitGroup{}
	strategiesType := strategies.New(wg, cfg, tradeApi, ordProc, dbManager, log, initSignals,
		strategy.NewBBRSIStrategy(cfg, tradeApi, ordProc, dbManager, caches, log),
		caches,
	)
	strategiesType.Start()
	<-done

	dbManager.Close()
	log.Infof("service tccbot stopped")
}
