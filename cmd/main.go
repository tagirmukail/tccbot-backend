package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/tagirmukail/tccbot-backend/internal/orderproc"

	"github.com/tagirmukail/tccbot-backend/internal/strategies"

	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"

	"github.com/tagirmukail/tccbot-backend/internal/db"

	"github.com/tagirmukail/tccbot-backend/internal/config"

	"github.com/tagirmukail/tccbot-backend/internal/utils/logger"

	"github.com/sirupsen/logrus"
)

// TODO добавить инициализацию предыдущих сигналов если отсутствуют

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

	log.Infof("configuration: %#v", cfg)

	if migrationOnly {
		log.Info("migrations started")
	}

	dbManager, err := db.New(cfg, nil, log, db.Command(migrationCommand))
	if err != nil {
		log.Fatalf("db.New() error: %v", err)
	}

	if migrationOnly {
		log.Info("migrations completed")
		return
	}

	tradeApi := tradeapi.NewTradeApi(
		cfg.Accesses.Bitmex.Key,
		cfg.Accesses.Bitmex.Secret,
		log,
		testMode,
	)

	ordProc := orderproc.New(300, tradeApi, cfg, log)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	wg := &sync.WaitGroup{}
	strategiesType := strategies.New(wg, cfg, tradeApi, ordProc, dbManager, log, initSignals)
	strategiesType.Start()
	<-done

	dbManager.Close()
	log.Infof("service tccbot stopped")
}
