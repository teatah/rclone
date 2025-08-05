package middleware

import (
	"context"
	"net/http"

	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/responses"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/token"
	"go.uber.org/zap"
)

func AuthMiddleware(
	sm session.SessionManager,
	lgr *zap.SugaredLogger,
	next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := token.TokenFromHeader(r)

		sess, err := sm.Check(r.Context(), tokenString)
		if err != nil {
			rc := &responses.ResponseContext{Logger: lgr, Writer: w, Request: r}
			rc.HandleError(err)

			return
		}

		sessValue := session.SessionCtxValue("session")

		ctx := r.Context()
		ctx = context.WithValue(ctx, sessValue, sess)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
