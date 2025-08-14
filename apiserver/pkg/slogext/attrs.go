package slogext

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
)

func Cause(err error) slog.Attr {
	return slog.String("cause", err.Error())
}

func Reason(reason any) slog.Attr {
	return slog.String("reason", fmt.Sprintf("%v", reason))
}

func Trace() slog.Attr {
	return slog.String("trace", string(debug.Stack()))
}

func Signal(signal os.Signal) slog.Attr {
	return slog.String("signal", signal.String())
}
