package service

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"accelerator/internal/features/patterns/repository"
	"context"
)

type PatternsService struct {
	repo *repository.PatternsRepository
}

func NewPatternsService(repo *repository.PatternsRepository) *PatternsService {
	return &PatternsService{
		repo: repo,
	}
}

// --------------------------- СОЗДАНИЕ ШАБЛОНА ---------------------------

func (serv *PatternsService) CreatePatternService(ctx context.Context, callerID, groupID, name, description, summaryPrompt, additionalPrompt string) (*domains.Pattern, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role != "creator" && userInfo.Role != "admin" {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	var patternInfo *domains.Pattern

	if userInfo.Role == "creator" {
		// он создает глобальный шаблон, который не привязан к группе
		patternInfo, err = serv.repo.CreatePattern(ctx, name, description, summaryPrompt, additionalPrompt, callerID, nil)
	} else {
		// это админ, привязываем шаблон к группе
		patternInfo, err = serv.repo.CreatePattern(ctx, name, description, summaryPrompt, additionalPrompt, callerID, groupID)
	}

	patternInfo.ChangeFlag = true // админ или креатор могут менять созданные шаблоны
	return patternInfo, nil

}

// ------------------------ ПОЛУЧЕНИЕ ШАБЛОНА ПО ID -------------------------

func (serv *PatternsService) GetPattern(ctx context.Context, callerID, patternID string) (*domains.Pattern, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	// проверяем шаблон на существование
	patternInfo, err := serv.repo.SelectPatternByID(ctx, patternID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role == "user" || userInfo.Role == "admin" {
		// если шаблон глобальный, возвращаем его
		if patternInfo.GroupID == "" {
			patternInfo.ChangeFlag = false // админ или юзер не может изменть шаблон креатора
			return patternInfo, nil
		}

		// если пользователи состоят в группе, к которой привязан шаблон, возвращаем его
		isConsists, err := serv.repo.IsUserIntoGroup(ctx, userInfo.ID, patternInfo.GroupID)
		if err != nil {
			return nil, err
		}
		// если он состоит, возвращаем
		// ставим флаги
		if isConsists && userInfo.Role == "user" {
			patternInfo.ChangeFlag = false // юзер не может менять групповые шаблоны
			return patternInfo, nil
		}
		if isConsists && userInfo.Role == "admin" {
			patternInfo.ChangeFlag = true // админ может менять групповой шаблон
			return patternInfo, nil
		}
	}

	// для креатора возвращаем любой шаблон
	patternInfo.ChangeFlag = true // креатор может менять все шаблоны
	return patternInfo, nil
}

// ------------------------ ОБЩЕЕ ПОЛУЧЕНИЕ ШАБЛОНОВ ДЛЯ ВСЕХ ПОЛЬЗОВАТЕЛЕЙ -------------------------

// возвращаем два массива, первый глобальные шаблоны, второй групповые
func (serv *PatternsService) GetPatternsInGroupAndGlobal(ctx context.Context, callerID, groupID string) (*[]domains.Pattern, *[]domains.Pattern, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, nil, err
	}

	globalPatternsInfo, err := serv.repo.SelectGlobalPatterns(ctx) 
	if err != nil {
		return nil, nil, err
	}

	groupPatternsInfo, err := serv.repo.SelectGroupPatterns(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	if userInfo.Role == "user" || userInfo.Role == "admin" {
		for index := range *globalPatternsInfo { // глобальные они менять не могут
			(*globalPatternsInfo)[index].ChangeFlag  = false
		}

		if userInfo.Role == "user" {
			for index := range *groupPatternsInfo { // групповые они менять не могут
				(*groupPatternsInfo)[index].ChangeFlag  = false
			}
		} else {
			for index := range *groupPatternsInfo { // а админы могут
				(*groupPatternsInfo)[index].ChangeFlag  = true
			}
		}
	}

	// креатор может менять все
	for index := range *globalPatternsInfo { // глобальные они менять не могут
		(*globalPatternsInfo)[index].ChangeFlag  = true
	}
	for index := range *groupPatternsInfo {
		(*groupPatternsInfo)[index].ChangeFlag  = true
	}

	// для креатора возвращаем любой шаблон
	return globalPatternsInfo, groupPatternsInfo, nil
}

// ------------------------------ ПОЛУЧЕНИЕ СВОИХ ШАБЛОНОВ ДЛЯ КРЕАТОРА ---------------------------------

// возвращаем массив с глобальными шаблонами
func (serv *PatternsService) GetPatternsGlobal(ctx context.Context, callerID string) (*[]domains.Pattern, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role != "creator" {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	globalPatternsInfo, err := serv.repo.SelectGlobalPatterns(ctx) 
	if err != nil {
		return nil, err
	}

	// креатор может менять все
	for index := range *globalPatternsInfo {
		(*globalPatternsInfo)[index].ChangeFlag  = true
	}

	// для креатора возвращаем любой шаблон
	return globalPatternsInfo, nil
}

// ---------------- ПОЛУЧЕНИЕ ВСЕХ ШАБЛОНОВ ПО ГРУППАМ ДЛЯ КРЕАТОРА ------------------

func (serv *PatternsService) GetAllPatternsInGroups(ctx context.Context, callerID string) (*[]domains.GroupWithPatterns, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role != "creator" {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	allPatternsInfo, err := serv.repo.SelectGroupsWithPatterns(ctx) 
	if err != nil {
		return nil, err
	}

	// креатор может менять все
	for indexGroup := range *allPatternsInfo {
		for indexPattern := range (*allPatternsInfo)[indexGroup].Patterns {
			(*allPatternsInfo)[indexGroup].Patterns[indexPattern].ChangeFlag = true
		}
	}

	// для креатора возвращаем любой шаблон
	return allPatternsInfo, nil
}

// ------------------------------ ИЗМЕНЕНИЕ ШАБЛОНОВ ---------------------------------

// возвращает измененный шаблон
func (serv *PatternsService) EditPattern(ctx context.Context, callerID, patternID string, editInfo map[string]any) (*domains.Pattern, error) {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role != "creator" && userInfo.Role != "admin" {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	// проверяем шаблон на существование
	patternInfo, err := serv.repo.SelectPatternByID(ctx, patternID)
	if err != nil {
		return nil, err
	}

	var editPatternInfo *domains.Pattern

	if userInfo.Role == "admin" {
		if patternInfo.GroupID == "" { // админ не может изменять шаблоны креатора
			return nil, error_type.NewNotFound("Шаблон не найден")
		}
		// проверить, что пользователь состоит в группе, к которой привязан шаблон
		isConsists, err := serv.repo.IsUserIntoGroup(ctx, userInfo.ID, patternInfo.GroupID)
		if err != nil {
			return nil, err
		}
		// если он состоит, то изменяем шаблон
		if isConsists {
			editPatternInfo, err = serv.repo.EditPattern(ctx, patternID, editInfo)
			if err != nil {
				return nil, err
			}
		}
	} else { // креатор может изменять любой шаблон
		editPatternInfo, err = serv.repo.EditPattern(ctx, patternID, editInfo)
		if err != nil {
			return nil, err
		}
	}

	// если мы вернули измененный шаблон, значит можно менять
	editPatternInfo.ChangeFlag = true
	return editPatternInfo, nil
}

// ------------------------------ УДАЛЕНИЕ ШАБЛОНОВ ---------------------------------

// возвращает ошибку
func (serv *PatternsService) DeletePattern(ctx context.Context, callerID, patternID string) error {
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return err
	}

	if userInfo.Role != "creator" && userInfo.Role != "admin" {
		return error_type.NewNotFound("Страница не найдена")
	}

	// проверяем шаблон на существование
	patternInfo, err := serv.repo.SelectPatternByID(ctx, patternID)
	if err != nil {
		return err
	}

	if userInfo.Role == "admin" {
		if patternInfo.GroupID == "" { // админ не может удалять шаблоны креатора
			return error_type.NewNotFound("Шаблон не найден")
		}
		// проверить, что пользователь состоит в группе, к которой привязан шаблон
		isConsists, err := serv.repo.IsUserIntoGroup(ctx, userInfo.ID, patternInfo.GroupID)
		if err != nil {
			return err
		}
		// если он состоит, то удаляем шаблон
		if isConsists {
			if err := serv.repo.DeletePattern(ctx, patternID); err != nil {
				return err
			}
		}
	} else { // креатор может удалять любой шаблон
		if err := serv.repo.DeletePattern(ctx, patternID); err != nil {
			return err
		}
	}

	return nil
}
