package pipeline

import (
	"cplatform/internal/presentation"
	presentation_http "cplatform/internal/presentation/http"
	"cplatform/pkg/slogext"
	"log/slog"
	"net/http"
)

type RecoverMiddleware struct {
	logger *slog.Logger
}

func NewRecoverMiddleware(logger *slog.Logger) *RecoverMiddleware {
	return &RecoverMiddleware{
		logger: logger,
	}
}

func (m *RecoverMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic recovered", slogext.Reason(err), slogext.Trace())

				werr := presentation_http.WriteErrors(w, http.StatusInternalServerError, presentation.ErrUnknown)
				if werr != nil {
					m.logger.Warn("cannot write panic recovered errors", slogext.Cause(werr))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
