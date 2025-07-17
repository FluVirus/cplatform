package main

import (
	"context"
	"cplatform/cmd/server/configuration"
	credis "cplatform/internal/infrastructure/cache/redis"
	"cplatform/internal/infrastructure/persistence/postgres"
	presentation "cplatform/internal/presentation/controller/http"
	"cplatform/pkg/slogext"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := configuration.ReadConfigurationFromEnv()
	if err != nil {
		slog.Error("main: error reading configuration from env", slogext.Cause(err))
		os.Exit(2)
	}

	handlerOptions := &slog.HandlerOptions{
		Level: config.SlogLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	logger := slog.New(handler)

	redisOptions, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		logger.Error("main: wrong redis url", slogext.Cause(err))
		os.Exit(3)
	}

	redisClient := redis.NewClient(redisOptions)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("main: error while closing redis client", slogext.Cause(err))
		}
	}()

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer redisCancel()

	status := redisClient.Ping(redisCtx)
	if err = status.Err(); err != nil {
		logger.Error("main: cannot ping redis server", slogext.Cause(err))
		os.Exit(4)
	}

	pgCtx, pgCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer pgCancel()

	pgConnPool, err := pgxpool.New(pgCtx, config.PgsqlUrl)
	if err != nil {
		logger.Error("main: cannot create pgxpool", slogext.Cause(err))
		os.Exit(5)
	}
	defer pgConnPool.Close()

	pgPingCtx, pgPingCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer pgPingCancel()

	err = pgConnPool.Ping(pgPingCtx)
	if err != nil {
		logger.Error("main: cannot ping pgxpool", slogext.Cause(err))
		os.Exit(6)
	}

	logger.Info("main: successfully connected to services, staying online forever")

	// creating services
	redisCache := credis.NewRedisCache(redisClient, logger)
	uowFactory := postgres.NewUnitOfWorkFactory(pgConnPool, logger)

	controller := presentation.NewController(uowFactory, redisCache, logger)

	router := presentation.NewRouter(controller)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	serverChan := make(chan error)
	server := http.Server{
		Addr:         ":80",
		Handler:      router.GetHandler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		defer close(serverChan)

		msg := fmt.Sprintf("server starts at addr %s", server.Addr)
		logger.Info(msg)
		err := server.ListenAndServe()
		if err != nil {
			serverChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
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
