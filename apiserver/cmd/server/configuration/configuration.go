package configuration

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Configuration struct {
	RedisUrl  string
	PgsqlUrl  string
	SlogLevel slog.Level
}

var (
	ErrRedisUrlNotFound    = errors.New("redis URL not found")
	ErrPgsqlUrlNotFound    = errors.New("pgsql connection string not found")
	ErrLoggingLevelInvalid = errors.New("logging level invalid")
)

var strToSlog = map[string]slog.Level{
	"":      slog.LevelInfo,
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func ReadConfigurationFromEnv() (*Configuration, error) {
	errs := make([]error, 0)

	redisUrl := os.Getenv("APISERVER_REDIS_URL")
	if redisUrl == "" {
		errs = append(errs, ErrRedisUrlNotFound)
	}

	pgsqlUrl := os.Getenv("APISERVER_PGSQL_URL")
	if pgsqlUrl == "" {
		errs = append(errs, ErrPgsqlUrlNotFound)
	}

	logLevelStr := os.Getenv("APISERVER_LOG_LEVEL")
	logLevelStr = strings.ToLower(logLevelStr)

	logLevel, ok := strToSlog[logLevelStr]
	if !ok {
		errs = append(errs, fmt.Errorf("%w: %q level is not recognized; valid are: debug, info, warn, error; default is info", ErrLoggingLevelInvalid, logLevelStr))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	configuration := &Configuration{
		RedisUrl:  redisUrl,
		PgsqlUrl:  pgsqlUrl,
		SlogLevel: logLevel,
	}

	return configuration, nil
}
