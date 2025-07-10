package main

import (
	"context"
	"cplatform/internal/configuration"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := configuration.ReadConfigurationFromEnv()
	if err != nil {
		slog.Error("main: error reading configuration from env", "cause", err.Error())
		os.Exit(2)
	}

	handlerOptions := &slog.HandlerOptions{
		Level: config.SlogLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	logger := slog.New(handler)

	redisOptions, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		logger.Error("main: wrong redis url", "cause", err.Error())
		os.Exit(3)
	}

	redisClient := redis.NewClient(redisOptions)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("main: error while closing redis client", "cause", err.Error())
		}
	}()

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer redisCancel()

	status := redisClient.Ping(redisCtx)
	if err = status.Err(); err != nil {
		logger.Error("main: cannot ping redis server", "cause", err.Error())
		os.Exit(4)
	}

	pgCtx, pgCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer pgCancel()

	pgConnPool, err := pgxpool.New(pgCtx, config.PgsqlUrl)
	if err != nil {
		logger.Error("main: cannot create pgxpool", "cause", err.Error())
		os.Exit(5)
	}
	defer pgConnPool.Close()

	pgPingCtx, pgPingCancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer pgPingCancel()

	err = pgConnPool.Ping(pgPingCtx)
	if err != nil {
		logger.Error("main: cannot ping pgxpool", "cause", err.Error())
		os.Exit(6)
	}

	logger.Info("main: successfully connected to services, staying online forever")

	<-sigCtx.Done()
}
