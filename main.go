package main

import (
	"fmt"

	"github.com/andyle182810/gframework/goredis"
	"github.com/andyle182810/gframework/httpserver"
	"github.com/andyle182810/gframework/metricserver"
	"github.com/andyle182810/gframework/notifylog"
	"github.com/andyle182810/gframework/postgres"
	"github.com/andyle182810/gframework/runner"
	"github.com/jackc/pgx/v5/tracelog"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/thienhaole92/aws-auto-go-app/internal/config"
	"github.com/thienhaole92/aws-auto-go-app/internal/service"
)

const (
	serviceName = "aws-auto-go-app"
)

func main() {
	log := notifylog.New(serviceName, notifylog.JSON)

	if err := run(log); err != nil {
		log.Fatal().Err(err).Msg("Application exited with an error")
	}

	log.Info().Msg("Application shutdown complete")
}

func run(log notifylog.NotifyLog) error {
	cfg, err := config.New(nil)
	if err != nil {
		return err
	}

	postgresClient, err := newPostgresClient(cfg)
	if err != nil {
		return err
	}

	redisClient, err := newRedisClient(cfg)
	if err != nil {
		return err
	}

	apiService := service.New(
		redisClient,
		postgresClient,
	)

	app := &application{
		cfg:     cfg,
		log:     log,
		service: apiService,
	}

	appRunner := runner.New(
		runner.WithCoreService(app.newMetricServer()),
		runner.WithCoreService(app.newHTTPServer()),
		runner.WithInfrastructureService(postgresClient),
		runner.WithInfrastructureService(postgresClient),
	)

	appRunner.Run()

	return nil
}

type application struct {
	cfg     *config.Config
	log     notifylog.NotifyLog
	service *service.Service
}

func (app *application) newHTTPServer() *httpserver.Server {
	httpCfg := &httpserver.Config{
		Host:         app.cfg.HTTPServerHost,
		Port:         app.cfg.HTTPServerPort,
		EnableCors:   app.cfg.HTTPEnableCORS,
		BodyLimit:    app.cfg.HTTPBodyLimit,
		ReadTimeout:  app.cfg.HTTPServerReadTimeout,
		WriteTimeout: app.cfg.HTTPServerWriteTimeout,
		GracePeriod:  app.cfg.GracefulShutdownPeriod,
	}

	svr := httpserver.New(httpCfg)
	app.registerRoutes(svr.Root)

	return svr
}

func (app *application) newMetricServer() *metricserver.Server {
	metricCfg := &metricserver.Config{
		Host:         app.cfg.MetricServerHost,
		Port:         app.cfg.MetricServerPort,
		ReadTimeout:  app.cfg.MetricServerReadTimeout,
		WriteTimeout: app.cfg.MetricServerWriteTimeout,
		GracePeriod:  app.cfg.GracefulShutdownPeriod,
	}

	return metricserver.New(metricCfg)
}

func (app *application) registerRoutes(root *echo.Group) {
	root.GET("/health", httpserver.Wrapper(app.service.CheckHealth))
}

func newPostgresClient(cfg *config.Config) (*postgres.Postgres, error) {
	if cfg.MigrationEnabled {
		log.Info().Str("source", cfg.MigrationSource).Msg("Starting database migration process...")

		if err := postgres.MigrateUp(cfg.PostgresURL, cfg.MigrationSource); err != nil {
			return nil, fmt.Errorf("postgresql migration failed: %w", err)
		}

		log.Info().Msg("Database migration process completed successfully")
	}

	pgCfg := &postgres.Config{
		URL:                   cfg.PostgresURL,
		MaxConnection:         cfg.PostgresMaxConnection,
		MinConnection:         cfg.PostgresMinConnection,
		MaxConnectionIdleTime: cfg.PostgresMaxConnectionIdleTime,
		LogLevel:              tracelog.LogLevel(cfg.PostgresLogLevel),
	}

	postgresDB, err := postgres.New(pgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	log.Info().Msg("PostgreSQL client initialized successfully")

	return postgresDB, nil
}

func newRedisClient(cfg *config.Config) (*goredis.Redis, error) {
	redisCfg := &goredis.Config{
		Host:         cfg.KeyvalHost,
		Port:         cfg.KeyvalPort,
		Password:     cfg.KeyvalPassword,
		DB:           cfg.KeyvalDB,
		DialTimeout:  cfg.KeyvalDialTimeout,
		MaxIdleConns: cfg.KeyvalMaxIdleConns,
		MinIdleConns: cfg.KeyvalMinIdleConns,
		PingTimeout:  cfg.KeyvalPingTimeout,
		PoolSize:     cfg.KeyvalPoolSize,
		ReadTimeout:  cfg.KeyvaReadTimeout,
		WriteTimeout: cfg.KeyvaWriteTimeout,
	}

	redisClient, err := goredis.New(redisCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis client: %w", err)
	}

	log.Info().Msg("Redis client initialized successfully")

	return redisClient, nil
}
