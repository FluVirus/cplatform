package main

import (
	"context"
	"cplatform/cmd/server/configuration"
	"cplatform/internal/application/authentication/basic"
	"cplatform/internal/di/middleware"
	"cplatform/internal/di/scope"
	credis "cplatform/internal/infrastructure/cache/redis"
	"cplatform/internal/infrastructure/persistence/postgres"
	controller_http "cplatform/internal/presentation/http/controller"
	"cplatform/internal/presentation/http/pipeline"
	"cplatform/pkg/slogext"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
)

func main() {
	config, err := configuration.ReadConfigurationFromEnv()
	if err != nil {
		slog.Error("main: error reading configuration from env", slogext.Cause(err))
		os.Exit(2)
	}

	// LOGGER
	handlerOptions := &slog.HandlerOptions{
		Level: config.SlogLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	logger := slog.New(handler)

	// REDIS
	redisOptions, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		logger.Error("main: fail parse redis url", slogext.Cause(err))
		return
	}

	redisClient := redis.NewClient(redisOptions)

	// POSTGRES
	pgConfig, err := pgxpool.ParseConfig(config.PgsqlUrl)
	if err != nil {
		logger.Error("main: fail parse pgsql url", slogext.Cause(err))
		return
	}

	pgPool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		logger.Error("main: fail create pgxpool", slogext.Cause(err))
		return
	}

	// INFRASTRUCTURE
	uowFactory := postgres.NewUnitOfWorkFactory(pgPool, logger)
	redisCache := credis.NewRedisCache(redisClient, logger)

	// SCOPES
	scopeFactory := scope.NewFactory(uowFactory, redisCache, logger)

	// API
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"*"},
		AllowedHeaders: []string{"*"},
	})

	recoverMiddleware := pipeline.NewRecoverMiddleware(logger)

	isoLevelMiddleware := middleware.NewIsoLevelMiddleware(logger)
	scopeMiddleware := middleware.NewScopeMiddleware(logger, scopeFactory)
	basicAuthMiddleware := basic.NewBasicAuthMiddleware(logger)

	controller := controller_http.NewController(logger)
	router := controller_http.NewRouter(controller, isoLevelMiddleware, scopeMiddleware, basicAuthMiddleware, logger)
	p := pipeline.NewPipeline(router, recoverMiddleware, corsMiddleware, logger)

	// HTTP
	server := http.Server{
		Addr:         ":80",
		Handler:      p.CreateHandler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	serverChan := make(chan error)
	go func() {
		serverChan <- server.ListenAndServe()
	}()

	select {
	case sig := <-sigChan:
		logger.Info("main: start shutdown", slogext.Signal(sig))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("main: cannot shutdown server gracefully", slogext.Signal(sig), slogext.Cause(err))
			os.Exit(32)
		} else {
			logger.Info("main: server shutdown gracefully", slogext.Signal(sig))
		}
	case err := <-serverChan:
		logger.Error("main: server error", slogext.Cause(err))
		os.Exit(33)
	}
}
