package authctx

import "context"

type contextKey string
// для получения значения из контекста
const userIDKey contextKey = "userID"

// WithUserID возвращает новый контекст с userID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID извлекает userID из контекста (возвращает значение и флаг успеха)
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
