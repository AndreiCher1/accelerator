package middleware

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/server/authctx"
	"accelerator/internal/tools"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет токен и кладёт userID в контекст
func AuthMiddleware(cfg *config.Config) func (http.Handler) http.Handler { // в самой верхней части получаем конфиг и возвращаем саму функцию миддлваре, по сути обертка для приема значений, возвращает обычный миддлваре
	return func(next http.Handler) http.Handler { // чтобы миддлваре соответствовал паттерну всех миддлваре
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // так как HandleFunc принимает только wr, то next нужно передавать в обертке
			// 1. Получаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				tools.WriteError(w, error_type.NewUnauthorized("Не передан access токен"))
				return
			}

			// 2. Проверяем формат "Bearer <token>"
			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				tools.WriteError(w, error_type.NewUnauthorized("Неправильно передан access токен"))
				return
			}
			tokenString := strings.TrimPrefix(authHeader, prefix)

			// 3. Валидируем токен через готовую функцию
			userID, userRole, err := tools.ParseAccessToken(tokenString, cfg)
			if err != nil {
				tools.WriteError(w, error_type.NewUnauthorized("Невалидный access токен"))
				return
			}

			// 4. Кладём userID в контекст
			ctx := authctx.WithUserID(r.Context(), userID)
			ctx = authctx.WithUserRole(ctx, userRole)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
