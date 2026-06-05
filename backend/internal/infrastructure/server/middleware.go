package server

import (
	"context"
	"net/http"

	appauth "Dormitory_Booking/internal/application/auth"
	"Dormitory_Booking/internal/domain/auth"
)

type ctxKey string

const ctxUserKey ctxKey = "auth_user"

// UserFromCtx достаёт аутентифицированного пользователя из контекста запроса.
func UserFromCtx(ctx context.Context) (auth.User, bool) {
	u, ok := ctx.Value(ctxUserKey).(auth.User)
	return u, ok
}

// SessionMiddleware читает cookie `session`, проверяет сессию и кладёт User в контекст.
// Не блокирует неаутентифицированные запросы — просто не устанавливает пользователя.
func SessionMiddleware(authSvc *appauth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authSvc == nil {
				next.ServeHTTP(w, r)
				return
			}
			c, err := r.Cookie("session")
			if err != nil || c.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			_, user, err := authSvc.GetSession(r.Context(), c.Value)
			if err != nil {
				// Сессия истекла — сбрасываем куку.
				http.SetCookie(w, &http.Cookie{
					Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
				})
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
