package slogext

import (
	"log/slog"
	"os"
)

func Cause(err error) slog.Attr {
	return slog.String("cause", err.Error())
}

func Signal(signal os.Signal) slog.Attr {
	return slog.String("signal", signal.String())
}
