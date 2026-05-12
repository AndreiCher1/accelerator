package transport

import (
	"accelerator/internal/domains"
	"accelerator/internal/features/patterns/service"
	"context"

	"github.com/go-playground/validator/v10"
)

type PatternsTransport struct {
	serv     *service.PatternsService
	validate *validator.Validate
}

func NewPatternsTransport(serv *service.PatternsService, validate *validator.Validate) *PatternsTransport {
	return &PatternsTransport{
		serv:     serv,
		validate: validate,
	}
}

// ============================== СОЗДАНИЕ НОВОГО ШАБЛОНА ==============================
func (serv *PatternsTransport) CreatePattern(ctx context.Context, callerID, name, description, summaryPrompt, additionalPrompt string) (*domains.Pattern, error) {

}

// ================================= ПОЛУЧЕНИЕ ШАБЛОНА =================================
func (serv *PatternsTransport) GetPattern(ctx context.Context, callerID, patternID string) (*[]domains.Pattern, error) {

}

// ============================== ПОЛУЧЕНИЕ ВСЕХ ШАБЛОНОВ ===============================
func (serv *PatternsTransport) GetAllPatterns(ctx context.Context, callerID string) (*[]domains.Pattern, error) {
	
}

// =============================== ИЗМЕНЕНИЕ ШАБЛОНА ===================================
func (serv *PatternsTransport) EditPattern(ctx context.Context, callerID, patternID string, editInfo map[string]string) (*domains.Pattern, error) {

}

// ============================= УДАЛЕНИЕ ШАБЛОНА =====================================
func (serv *PatternsTransport) DeletePattern(ctx context.Context, callerID, patternID string) error {
	// проверяем, что вызывающий пользователь это админ или креатор

}