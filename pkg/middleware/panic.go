package middleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func PanicMiddleware(lgr *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)

				lgr.Errorw(fmt.Sprintf("recovered from error: %s", fmt.Sprint(err)),
					"method", r.Method,
					"remote_addr", r.RemoteAddr,
					"url", r.URL.Path,
				)
			}
			next.ServeHTTP(w, r)
		}()
	})
}
