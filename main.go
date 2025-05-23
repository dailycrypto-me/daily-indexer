//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=models/models.cfg.yaml api/openapi.yaml

package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"strconv"

	"github.com/dailycrypto-me/daily-indexer/api"
	"github.com/dailycrypto-me/daily-indexer/internal/chain"
	"github.com/dailycrypto-me/daily-indexer/internal/common"
	"github.com/dailycrypto-me/daily-indexer/internal/indexer"
	"github.com/dailycrypto-me/daily-indexer/internal/logging"
	"github.com/dailycrypto-me/daily-indexer/internal/metrics"
	"github.com/dailycrypto-me/daily-indexer/internal/storage/pebble"
	migration "github.com/dailycrypto-me/daily-indexer/internal/storage/pebble/migrations"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
	log "github.com/sirupsen/logrus"
)

var (
	http_port                        *int
	metrics_port                     *int
	blockchain_ws                    *string
	data_dir                         *string
	log_level                        *string
	yield_saving_interval            *int
	validators_yield_saving_interval *int
	sync_queue_limit                 *int
	chain_stats_interval             *int
)

func init() {
	http_port = flag.Int("http_port", 8080, "port to listen")
	metrics_port = flag.Int("metrics_port", 2112, "metrics http port")
	blockchain_ws = flag.String("blockchain_ws", "wss://ws.mainnet.dailycrypto.net", "ws url to connect to blockchain")
	data_dir = flag.String("data_dir", "./data", "path to directory where indexer database will be saved")
	log_level = flag.String("log_level", "info", "minimum log level. could be only [trace, debug, info, warn, error, fatal]")
	yield_saving_interval = flag.Int("yield_saving_interval", 25000, "interval for saving total yield")
	validators_yield_saving_interval = flag.Int("validators_yield_saving_interval", 25000, "interval for saving validators yield")
	sync_queue_limit = flag.Int("sync_queue_limit", 10, "limit of blocks in the sync queue")
	chain_stats_interval = flag.Int("chain_stats_interval", 100, "interval for saving chain stats")

	flag.Parse()

	logging.Config(filepath.Join(*data_dir, "logs"), *log_level)
	log.Print("\n\n\n")
	log.WithFields(log.Fields{
		"http_port":     *http_port,
		"blockchain_ws": *blockchain_ws,
		"data_dir":      *data_dir,
		"log_level":     *log_level}).
		Info("Application initialized")
}

func setupCloseHandler(fn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-c
		fn()
		os.Exit(0)
	}()
}

func main() {
	log.Info("Starting Daily Indexer")
	st := pebble.NewStorage(filepath.Join(*data_dir, "db"))
	setupCloseHandler(func() { st.Close() })
	fin := st.GetFinalizationData()

	swagger, err := api.GetSwagger()
	if err != nil {
		log.WithError(err).Fatal("Error loading swagger spec")
	}

	manager := migration.NewManager(st, *blockchain_ws)
	err = manager.ApplyAll()
	if err != nil {
		log.WithError(err).Fatal("Error applying migrations")
	}

	swagger.Servers = nil

	e := echo.New()

	e.Use(echomiddleware.OapiRequestValidator(swagger))
	// Add http error handler to return a proper error JSON on request error
	e.HTTPErrorHandler = func(err error, ctx echo.Context) {
		_ = ctx.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	c := common.DefaultConfig()
	c.TotalYieldSavingInterval = uint64(*yield_saving_interval)
	c.ValidatorsYieldSavingInterval = uint64(*validators_yield_saving_interval)
	c.SyncQueueLimit = uint64(*sync_queue_limit)
	c.ChainStatsInterval = *chain_stats_interval

	log.WithFields(log.Fields{"pbft_count": fin.PbftCount, "dag_count": fin.DagCount, "trx_count": fin.TrxCount}).Info("Loaded db with")
	chainStats := chain.MakeStats(c.ChainStatsInterval)
	apiHandler := api.NewApiHandler(st, c, chainStats)
	api.RegisterHandlers(e, apiHandler)

	go indexer.MakeAndRun(*blockchain_ws, st, c, chainStats)

	// start a http server for prometheus on a separate go routine
	go metrics.RunPrometheusServer(":" + strconv.FormatInt(int64(*metrics_port), 10))

	err = e.Start(":" + strconv.FormatInt(int64(*http_port), 10))
	log.WithError(err).Fatal("Can't start http server")
}
