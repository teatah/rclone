package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func LogMiddleware(lgr *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr.Infow("New request",
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
			"url", r.URL.Path,
		)
		next.ServeHTTP(w, r)
	})
}
