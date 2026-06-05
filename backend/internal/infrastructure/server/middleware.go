package server

import (
	"context"
	"net/http"

	authsvc "Dormitory_Booking/internal/application/auth"
	adminsvc "Dormitory_Booking/internal/application/admin"
	"Dormitory_Booking/internal/domain/user"
)

type contextKey string

const (
	ctxUser    contextKey = "user"
	ctxIsAdmin contextKey = "isAdmin"
)

func userFromCtx(r *http.Request) (user.User, bool) {
	u, ok := r.Context().Value(ctxUser).(user.User)
	return u, ok
}

func isAdminFromCtx(r *http.Request) bool {
	v, _ := r.Context().Value(ctxIsAdmin).(bool)
	return v
}

// SessionMiddleware читает куки сессии и кладёт пользователя в контекст
func SessionMiddleware(authService *authsvc.Service, adminService *adminsvc.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			u, err := authService.GetSession(r.Context(), cookie.Value)
			if err != nil {
				// невалидная сессия, чистим куку
				http.SetCookie(w, &http.Cookie{
					Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
				})
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUser, u)

			isAdmin, _ := adminService.IsAdmin(r.Context(), u.Email)
			ctx = context.WithValue(ctx, ctxIsAdmin, isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth возвращает 401 если нет сессии
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromCtx(r); !ok {
			http.Error(w, "требуется авторизация", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin возвращает 403 если пользователь не администратор
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromCtx(r); !ok {
			http.Error(w, "требуется авторизация", http.StatusUnauthorized)
			return
		}
		if !isAdminFromCtx(r) {
			http.Error(w, "доступ запрещён", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
