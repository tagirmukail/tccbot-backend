package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tagirmukail/tccbot-backend/internal/config"

	"github.com/tagirmukail/tccbot-backend/internal/scheduler"

	bitmextradedata "github.com/tagirmukail/tccbot-backend/internal/tradedata/bitmex"

	migrate_db "github.com/tagirmukail/tccbot-backend/internal/db/migrate-db"

	"github.com/tagirmukail/tccbot-backend/internal/candlecache"
	"github.com/tagirmukail/tccbot-backend/internal/db"

	"github.com/tagirmukail/tccbot-backend/internal/strategies/strategy"

	"github.com/sirupsen/logrus"

	"github.com/tagirmukail/tccbot-backend/internal/orderproc"
	"github.com/tagirmukail/tccbot-backend/internal/strategies"
	"github.com/tagirmukail/tccbot-backend/internal/types"
	"github.com/tagirmukail/tccbot-backend/internal/utils/logger"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi"
	"github.com/tagirmukail/tccbot-backend/pkg/tradeapi/bitmex/ws"
)

const (
	maxCandles = 100
	PidEnvKey  = "TCCBOTBACKENDPID"
)

//Application version
var (
	Version   string
	DateBuild string
	GitHash   string
)

// TODO новая арбитражная стратегия - выставлять сразу и на покупку и на продажу при срабатывании сигналов,
//  и при необходимости перевыставлять,
//  учитывать при выставлении в какую сторону сработали сигналы, учитывать доступный балан и позицию

//  atr signal/ strategy ( Awesome Oscillator + Accelerator Oscillator + Parabolic SAR)

// TODO попробовать пакет кобра для запуска команд
// TODO вынести инициализацию зависимостей бота отдельно
func main() { // nolint:funlen
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

	err := killIfRun()
	if err != nil {
		logrus.Fatal(err)
	}

	err = setPidEnv()
	if err != nil {
		logrus.Fatal(err)
	}

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

	log, err := logger.New(logLevel, logDir,
		&logrus.TextFormatter{DisableColors: false, ForceColors: true, FullTimestamp: true})
	if err != nil {
		logrus.Fatal(err)
	}
	log.Info("service tccbot started...")

	log.Infof("\nversion: %v;\ndate_build: %v;\ngit_hash: %v",
		Version, DateBuild, GitHash)

	updConfig, err := config.NewUpdConfigurator(configPath)
	if err != nil {
		log.Fatal(err)
	}
	updConfig.UpdateRun(5*time.Second, configPath)

	var dbManager db.DatabaseManager
	dbManager, err = db.NewDB(updConfig, nil, log, migrate_db.Command(migrationCommand), step)
	if err != nil {
		log.Fatalf("migartion failed: %v", err)
	}
	if migrationOnly {
		log.Info("migrations completed")
		return
	}

	var tradeThemes = updConfig.GetConfig().GlobStrategies.GetThemes()

	bitmexKey, bitmexSecret := updConfig.GetConfig().Accesses.Bitmex.Key, updConfig.GetConfig().Accesses.Bitmex.Secret
	if testMode {
		bitmexKey = updConfig.GetConfig().Accesses.Bitmex.Testnet.Key
		bitmexSecret = updConfig.GetConfig().Accesses.Bitmex.Testnet.Secret
	}

	tradeAPI := tradeapi.NewTradeAPI(
		bitmexKey,
		bitmexSecret,
		log,
		testMode,
		ws.NewWS(
			log,
			testMode,
			cfg.ExchangesSettings.Bitmex.PingSec,
			cfg.ExchangesSettings.Bitmex.TimeoutSec,
			uint32(cfg.ExchangesSettings.Bitmex.RetrySec),
			append([]types.Theme{types.Position}, tradeThemes...),
			types.Symbol(cfg.ExchangesSettings.Bitmex.Symbol),
			bitmexKey,
			bitmexSecret,
		),
	)

	var bitmexSubscribers []*bitmextradedata.Subscriber

	ordProc := orderproc.New(tradeAPI, cfg, log)

	caches := candlecache.NewBinToCache(
		cfg.GlobStrategies.GetBinSizes(), maxCandles, types.Symbol(cfg.ExchangesSettings.Bitmex.Symbol), log,
	)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	bitmexSubsTradeForStrategies := bitmextradedata.NewSubscriber(tradeThemes)
	bitmexSubscribers = append(bitmexSubscribers, bitmexSubsTradeForStrategies)

	var schedulr scheduler.Scheduler
	if cfg.Scheduler.Position.Enable {
		bitmexSubsForScheduler := bitmextradedata.NewSubscriber([]types.Theme{types.Position})
		bitmexSubscribers = append(bitmexSubscribers, bitmexSubsForScheduler)
		schedulr = scheduler.NewPositionScheduler(
			cfg, scheduler.LimitPositionPnls, tradeAPI, ordProc, bitmexSubsForScheduler, log,
		)
	}

	bitmexDataSender := bitmextradedata.New(tradeAPI.GetBitmex().GetWS().GetMessages(), log, bitmexSubscribers...)

	bbRsi := strategy.NewBBRSIStrategy(cfg, tradeAPI, ordProc, dbManager, caches, log)

	wg := &sync.WaitGroup{}
	strategiesTypes := strategies.New(wg, cfg, tradeAPI, ordProc, bitmexDataSender, bitmexSubsTradeForStrategies,
		schedulr, dbManager, log, initSignals, bbRsi, caches)
	strategiesTypes.Start()
	<-done

	dbManager.Close()
	log.Infof("service tccbot stopped")
}

func setPidEnv() error {
	pid := strconv.Itoa(os.Getpid())

	err := ioutil.WriteFile(PidEnvKey, []byte(pid), 0600)
	if err != nil {
		return err
	}

	logrus.Infof("started service pid: %v", pid)
	return nil
}

func killIfRun() error {
	f, err := os.OpenFile(PidEnvKey, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	pid, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	f.Close()

	if len(pid) == 0 {
		logrus.Infof("service  not runned")
		return nil
	}

	pidNumb, err := strconv.Atoi(string(pid))
	if err != nil {
		return err
	}

	logrus.Infof("service stopped, pid: %v", pidNumb)

	err = syscall.Kill(pidNumb, syscall.SIGTERM)
	if err != nil {
		if strings.Contains(err.Error(), "no such process") {
			return nil
		}
		return err
	}

	return nil
}
