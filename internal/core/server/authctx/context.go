package authctx

import "context"

type contextKey string
// для получения значения из контекста
const userIDKey contextKey = "userID"
const userRoleKey contextKey = "userRole"

// WithUserID возвращает новый контекст с userID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
// возвращает новый контекст с userRole
func WithUserRole(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userRoleKey, userID)
}

// GetUserID извлекает userID из контекста (возвращает значение и флаг успеха)
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
// извлекает userRole из контекста (возвращает значение и флаг успеха)
func GetUserRole(ctx context.Context) (string, bool) {
	userRole, ok := ctx.Value(userRoleKey).(string)
	return userRole, ok
}
