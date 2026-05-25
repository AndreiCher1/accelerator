package transport

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/server/authctx"
	"accelerator/internal/features/auth/service"
	"accelerator/internal/tools"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type AuthTransport struct {
	serv     *service.AuthService
	validate *validator.Validate
}

func NewAuthTransport(serv *service.AuthService, validate *validator.Validate) *AuthTransport {
	return &AuthTransport{
		serv:     serv,
		validate: validate,
	}
}

// =========================== ВХОД ПОЛЬЗОВАТЕЛЯ ====================================

type RequestAuthDTO struct {
	Login    string `json:"login" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ResponceTokensDTO struct {
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token,omitempty"` // не всегда нужен
	AccessExpireTime  time.Time `json:"expires_at"`
	Role              string    `json:"user_role"`
	TokenType         string    `json:"token_type"`
	TemporaryPassword bool      `json:"temporary_password"`
}

// возвращает temporaryPassword, если оно true при входе, то нужно перенаправить пользователя на страницу для смены временного пароля
func (trans *AuthTransport) LoginHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// парсим json в структуру для дальнейшей работы
	newRequest := RequestAuthDTO{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Не удалось распарсить JSON"))
		return
	}
	// валидируем полученный json по тегам
	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка валидации"))
		return
	}
	// если все окей, отправляем запрос в сервис для входа и получения токенов
	tokensInfo, userRole, temporaryPassword, err := trans.serv.LoginUserService(ctx, newRequest.Login, newRequest.Password)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// отправляем токены клиенту
	newResponse := ResponceTokensDTO{
		AccessToken:       tokensInfo.AccessToken,
		RefreshToken:      tokensInfo.RefreshToken,
		AccessExpireTime:  tokensInfo.AccessExpireTime,
		TokenType:         "Bearer",
		Role:              userRole,
		TemporaryPassword: temporaryPassword,
	}

	// записываем ответ с токенами пользователю
	tools.WriteJSON(w, http.StatusOK, newResponse)
}

// =========================== ОБНОВЛЕНИЕ ТОКЕНОВ ====================================

type RequestTokenDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (trans *AuthTransport) RefreshHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// парсим json в структуру для дальнейшей работы
	newRequest := RequestTokenDTO{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Не удалось распарсить JSON"))
		return
	}
	// валидируем полученный json по тегам
	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Отсутствует токен авторизации"))
		return
	}
	// если все окей, отправляем запрос в сервис для входа и получения токенов
	tokensInfo, userRole, err := trans.serv.RefreshUserService(ctx, newRequest.RefreshToken)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// отправляем токены клиенту
	newResponse := ResponceTokensDTO{
		AccessToken:      tokensInfo.AccessToken,
		RefreshToken:     tokensInfo.RefreshToken,
		AccessExpireTime: tokensInfo.AccessExpireTime,
		Role:             userRole,
		TokenType:        "Bearer",
	}

	// записываем ответ с токенами пользователю
	tools.WriteJSON(w, http.StatusOK, newResponse)
}

// =========================== ИЗМЕНЕНИЕ ВРЕМЕННОГО ПАРОЛЯ ====================================

type ChangeTempPasswordRequestDTO struct {
	Password string `json:"password" validate:"required,min=8"`
}

func (trans *AuthTransport) ChangeTempPasswordHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
		return
	}

	// парсим json в структуру для дальнейшей работы
	newRequest := ChangeTempPasswordRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Не удалось распарсить JSON"))
		return
	}
	// валидируем полученный json по тегам
	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Минимальная длина поля - 8 символов: password"))
		return
	}

	// сохраняем временный пароль
	if err := trans.serv.ChangeTempPasswordService(ctx, callerID, newRequest.Password); err != nil {
		tools.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
