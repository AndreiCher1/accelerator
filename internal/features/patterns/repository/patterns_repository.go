package repository

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PatternsRepository struct {
	pool *pgxpool.Pool
}

func NewPatternsRepository(pool *pgxpool.Pool) *PatternsRepository {
	return &PatternsRepository{
		pool: pool,
	}
}

// -------------------------- ПОЛУЧЕНИЕ ПОЛЬЗОВАТЕЛЕЙ ----------------------------

// ищет пользователя по ID
// возвращает информацию о нем
func (repo *PatternsRepository) SelectUserByID(ctx context.Context, userID string) (*domains.User, error) {
	sqlQuery := `
		SELECT id, login, full_name, position, role, temporary_password, created_at
		FROM users
		where id = $1;
	`

	var userInfo domains.User
	err := repo.pool.QueryRow(ctx, sqlQuery, userID).Scan(
		&userInfo.ID,
		&userInfo.Login,
		&userInfo.FullName,
		&userInfo.Position,
		&userInfo.Role,
		&userInfo.TemporaryPassword, // проверке на входе через эту функцию
		&userInfo.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("Пользователь, над которым хотят совершить действие не найден")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get user info: %w", err))
	}

	return &userInfo, nil
}

// ----------------------------- СОЗДАНИЕ ШАБЛОНОВ ------------------------------

func (repo *PatternsRepository) CreatePattern(
	ctx context.Context,
	name, description, summaryPrompt, additionalPrompt, callerID string, groupID any,
) (*domains.Pattern, error) {
	sqlQuery := `
		INSERT INTO patterns (group_id, name, description, summary_prompt, additional_prompt, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING (id, group_id, name, description, summary_prompt, additional_prompt, created_by, created_at)
	`

	var patternInfo domains.Pattern
	var stringGroupID sql.NullString
	if err := repo.pool.QueryRow(
		ctx, sqlQuery, groupID, name, description, summaryPrompt, additionalPrompt, callerID,
	).Scan(
		&patternInfo.ID,
		&stringGroupID,
		&patternInfo.Name,
		&patternInfo.Description,
		&patternInfo.SummaryPrompt,
		&patternInfo.AdditionalPrompt,
		&patternInfo.CreatedBy,
		&patternInfo.CreatedAt,
	); err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("create pattern: %w", err))
	}

	patternInfo.GroupID = stringGroupID.String // если ничего нет, то будет пустая строка

	return &patternInfo, nil
}

// ----------------------------- ПОЛУЧЕНИЕ ШАБЛОНОВ ------------------------------

// ищет шаблон по ID
// возвращает информацию о шаблоне
func (repo *PatternsRepository) SelectPatternByID(ctx context.Context, patternID string) (*domains.Pattern, error) {
	sqlQuery := `
		SELECT id, group_id, name, description, summary_prompt, additional_prompt, created_by, created_at
		FROM patterns
		where id = $1;
	`

	var patternInfo domains.Pattern
	var stringGroupID sql.NullString
	err := repo.pool.QueryRow(ctx, sqlQuery, patternID).Scan(
		&patternInfo.ID,
		&stringGroupID,
		&patternInfo.Name,
		&patternInfo.Description,
		&patternInfo.SummaryPrompt,
		&patternInfo.AdditionalPrompt,
		&patternInfo.CreatedBy,
		&patternInfo.CreatedAt,
	)

	patternInfo.GroupID = stringGroupID.String // если ничего нет, то будет пустая строка

	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("Шаблон, над которым хотят совершить действие не найден")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get pattern info: %w", err))
	}

	return &patternInfo, nil
}

func (repo *PatternsRepository) SelectGlobalPatterns(ctx context.Context) (*[]domains.Pattern, error) {
	sqlQuery := `
		SELECT id, group_id, name, description, summary_prompt, additional_prompt, created_by, created_at
		FROM patterns
		where group_id = NULL;
	`

	return repo.fetchPatterns(ctx, sqlQuery)
}

func (repo *PatternsRepository) SelectGroupPatterns(ctx context.Context, groupID string) (*[]domains.Pattern, error) {
	sqlQuery := `
		SELECT id, group_id, name, description, summary_prompt, additional_prompt, created_by, created_at
		FROM patterns
		where group_id = $1;
	`

	return repo.fetchPatterns(ctx, sqlQuery, groupID)
}

func (repo *PatternsRepository) fetchPatterns(ctx context.Context, sqlQuery string, args ...any) (*[]domains.Pattern, error) {
	rows, err := repo.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("query patterns: %w", err))
	}
	defer rows.Close()

	patterns := make([]domains.Pattern, 0)
	for rows.Next() {
		var pattern domains.Pattern

		var stringGroupID sql.NullString // чтобы суметь прочитать NULL
		if err := rows.Scan(
			&pattern.ID,
			&stringGroupID,
			&pattern.Name,
			&pattern.Description,
			&pattern.SummaryPrompt,
			&pattern.CreatedBy,
			&pattern.CreatedAt,
		); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan patterns: %w", err))
		}

		pattern.GroupID = stringGroupID.String

		patterns = append(patterns, pattern)
	}
	return &patterns, nil
}

// возвращает мапу групп, где ключ - ID группы, с массивом шаблонов внутри для каждой группы
func (r *PatternsRepository) SelectGroupsWithPatterns(ctx context.Context) (*[]domains.GroupWithPatterns, error) {
	query := `
        SELECT 
            g.id,
            g.name,
            g.description,
            p.id,
            p.group_id,
            p.name,
            p.description,
            p.summary_prompt,
            p.additional_prompt,
            p.created_by,
            p.created_at
        FROM groups g
        LEFT JOIN patterns p ON p.group_id = g.id
        ORDER BY g.name, p.name
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("query groups with patterns: %w", err))
	}
	defer rows.Close()

	groupMap := make(map[string]domains.GroupWithPatterns)

	for rows.Next() {
		var (
			groupID, groupName, groupDesc                                       string
			patternID, patternGroupID, patternName, patternDesc, patternSummary string
			patternAdditional                                                   json.RawMessage
			patternCreatedBy                                                    string
			patternCreatedAt                                                    time.Time
		)

		err := rows.Scan(
			&groupID, &groupName, &groupDesc,
			&patternID, &patternGroupID, &patternName, &patternDesc, &patternSummary,
			&patternAdditional, &patternCreatedBy, &patternCreatedAt,
		)
		if err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan row patterns in groups: %w", err))
		}

		// Получаем или создаём запись группы
		group, exists := groupMap[groupID]
		if !exists {
			group = domains.GroupWithPatterns{
				Name:        groupName,
				Description: groupDesc,
				Patterns:    []domains.Pattern{},
			}
			groupMap[groupID] = group
		}

		pattern := domains.Pattern{
			ID:               patternID,
			GroupID:          patternGroupID,
			Name:             patternName,
			Description:      patternDesc,
			SummaryPrompt:    patternSummary,
			AdditionalPrompt: patternAdditional, // nil, если NULL
			CreatedBy:        patternCreatedBy,
			CreatedAt:        patternCreatedAt,
		}
		group.Patterns = append(group.Patterns, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("rows iteration patterns in groups: %w", err))
	}

	// Преобразуем map в слайс
    result := make([]domains.GroupWithPatterns, 0, len(groupMap))
    for groupID, group := range groupMap {
		group.GroupID = groupID
        result = append(result, group)
    }

	return &result, nil
}

// ----------------------------- ИЗМЕНЕНИЕ ШАБЛОНОВ ------------------------------

func (repo *PatternsRepository) EditPattern(ctx context.Context, patternID string, editInfo map[string]any) (*domains.Pattern, error) {
	// Фильтруем допустимые поля: name, description, owner_id
	// валидируем в хэндлере, но пусть будет на всякий
	allowed := map[string]bool{"name": true, "description": true, "summary_prompt": true, "additional_prompt": true}
	setClauses := make([]string, 0)
	args := make([]any, 0)
	i := 1
	for field, value := range editInfo {
		// проверяем на допустимость переданное поле
		if !allowed[field] {
			continue
		}

		args = append(args, value)

		// собираем массив для SET
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		i++
	}
	// также валидируем в хэндлере, но пусть будет
	// если где-то ошиблись с названиями
	if len(setClauses) == 0 {
		// Если нет допустимых полей, возвращаем текущую информацию о шаблоне без изменений
		return repo.SelectPatternByID(ctx, patternID)
	}
	// добавляем в конец ID чтобы потом распаковать
	args = append(args, patternID)

	sqlQuery := fmt.Sprintf(`
        UPDATE groups SET %s
        WHERE id = $%d
        RETURNING id, group_id, name, description, summary_prompt, additional_prompt, created_by, created_at;
    `, strings.Join(setClauses, ", "), i)

	var patternInfo domains.Pattern
	var stringGroupID sql.NullString
	err := repo.pool.QueryRow(ctx, sqlQuery, args...).Scan(
		&patternInfo.ID,
		&stringGroupID,
		&patternInfo.Name,
		&patternInfo.Description,
		&patternInfo.SummaryPrompt,
		&patternInfo.AdditionalPrompt,
		&patternInfo.CreatedBy,
		&patternInfo.CreatedAt,
	)

	patternInfo.GroupID = stringGroupID.String // если ничего нет, то будет пустая строка

	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("Шаблон, над которым хотят совершить действие не найден")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("edit pattern: %w", err))
	}

	return &patternInfo, nil
}

// ------------------------------- УДАЛЕНИЕ ШАБЛОНОВ --------------------------------

// принимает ID шаблона
// возвращает ошибку, если шаблона нет
func (repo *PatternsRepository) DeletePattern(ctx context.Context, patternID string) error {
	sqlQuery := `
		DELETE FROM patterns
		WHERE group_id = $1;
	`

	result, err := repo.pool.Exec(ctx, sqlQuery, patternID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("delete pattern into group: %w", err))
	}

	rowsAffected := result.RowsAffected() // возвращает количество удаленных строк, если ноль
	if rowsAffected == 0 {                // то пользователь не найден
		return error_type.NewNotFound("Шаблон не найден")
	}

	return nil
}

// ------------------------------- ФУНКЦИИ-ПРОВЕРКИ ----------------------------------

// проверяет членство в группе = true - есть в группе
func (repo *PatternsRepository) IsUserIntoGroup(ctx context.Context, userID, groupID string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM group_members WHERE user_id = $1 AND group_id = $2)`
	var exists bool
	err := repo.pool.QueryRow(ctx, query, userID, groupID).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
	}
	return exists, nil
}
