package repository

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{
		pool: pool,
	}
}

// ================================================ МЕТОДЫ РЕПОЗИТОРИЯ ДЛЯ СОЗДАНИЕ КРЕАТОРА ==============================================

// IsTableEmpty возвращает true, если таблица не содержит строк.
func (repo *AdminRepository) IsTableUsersEmpty(ctx context.Context) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users LIMIT 1)`
	err := repo.pool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check if table empty: %w", err))
	}

	return !exists, nil // true — пусто, false — не пусто
}

// добавляет нового пользователя, если логин уникальный
// возвращает его id
func (repo *AdminRepository) InsertNewCreator(
	ctx context.Context,
	login, passwordHash, fullname, position string,
) (string, error) {
	sqlQuery := `
	INSERT INTO users (login, password_hash, full_name, position, role, temporary_password)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	var userID string
	err := repo.pool.QueryRow(
		ctx, sqlQuery, login, passwordHash, fullname, position, "creator", false,
	).Scan(&userID)

	if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("create user creator: %w", err))
	}

	return userID, nil
}

// ================================================ МЕТОДЫ РЕПОЗИТОРИЯ ДЛЯ ПОЛЬЗОВАТЕЛЕЙ ==============================================

// ищет пользователя по ID
// возвращает информацию о нем
func (repo *AdminRepository) SelectUserByID(ctx context.Context, userID string) (*domains.User, error) {
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

// добавляет нового пользователя, если логин уникальный
// возвращает его id
func (repo *AdminRepository) InsertNewUser(
		ctx context.Context, 
		login, passwordHash, fullname, position, role string,
	) (string, error) {
	sqlQuery := `
	INSERT INTO users (login, password_hash, full_name, position, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	var userID string
	err := repo.pool.QueryRow(
		ctx, sqlQuery, login, passwordHash, fullname, position, role,
	).Scan(&userID)

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
		return "", error_type.NewConflict("A user with such an email already exists")
	} else if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("create user: %w", err))
	}

	return userID, nil
}

// ищет пользователей по роли делающего запрос
// возвращает список найденных пользователей, сортированный от admin->user в алфавитном порядке
func (repo *AdminRepository) SelectAllUsersWithAdminFirst(ctx context.Context, page, limit int) (*[]domains.User, error) {
	sqlQuery := `
	SELECT id, login, full_name, position, role, created_at
	FROM users
	WHERE role IN ('admin', 'user')
	ORDER BY 
		CASE role 
			WHEN 'admin' THEN 1  -- сначала по админу
			WHEN 'user' THEN 2 -- потом все юзеры
		END,
		full_name ASC -- а эти группы по алфавиту
	LIMIT $2
	OFFSET ($1 - 1) * $2;
	`

	rows, err := repo.pool.Query(ctx, sqlQuery, page, limit)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users info: %w", err))
	}

	usersInfo := make([]domains.User, 0)

	for rows.Next() {
		var userInfo domains.User

		if err := rows.Scan(
			&userInfo.ID,
			&userInfo.Login,
			&userInfo.FullName,
			&userInfo.Position,
			&userInfo.Role,
			&userInfo.CreatedAt,
		); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan user: %w", err))
		}

		usersInfo = append(usersInfo, userInfo)
	}

	return &usersInfo, nil
}

// ищет пользователей по роли делающего запрос
// возвращает список найденных пользователей user, сортированный в алфавитном порядке
func (repo *AdminRepository) SelectOnlyUsers(ctx context.Context, page, limit int) (*[]domains.User, error) {
	sqlQuery := `
	SELECT id, login, full_name, position, role, created_at
	FROM users
	WHERE role = 'user'
	ORDER BY full_name ASC 
	LIMIT $2
	OFFSET ($1 - 1) * $2;
	`

	rows, err := repo.pool.Query(ctx, sqlQuery, page, limit)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users info: %w", err))
	}

	usersInfo := make([]domains.User, 0)

	for rows.Next() {
		var userInfo domains.User

		if err := rows.Scan(
			&userInfo.ID,
			&userInfo.Login,
			&userInfo.FullName,
			&userInfo.Position,
			&userInfo.Role,
			&userInfo.CreatedAt,
		); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan user: %w", err))
		}

		usersInfo = append(usersInfo, userInfo)
	}

	return &usersInfo, nil
}

func (repo *AdminRepository) SelectCountUsersAndAdmins(ctx context.Context) (int64, error) {
	sqlQuery := `
	SELECT COUNT(*)
    FROM users
    WHERE role IN ('admin', 'user');
	`

	var countUsers int64 // из бд мы получаем 64, чтобы не обрезать при больших значениях будем получать int64
	err := repo.pool.QueryRow(ctx, sqlQuery).Scan(&countUsers)

	if err != nil {
		return 0, error_type.NewInternal(fmt.Errorf("select count users and admins: %w", err))
	}

	return countUsers, nil
}

func (repo *AdminRepository) SelectCountUsers(ctx context.Context) (int64, error) {
	sqlQuery := `
	SELECT COUNT(*)
    FROM users
    WHERE role = 'user';
	`

	var countUsers int64 // из бд мы получаем 64, чтобы не обрезать при больших значениях будем получать int64
	err := repo.pool.QueryRow(ctx, sqlQuery).Scan(&countUsers)

	if err != nil {
		return 0, error_type.NewInternal(fmt.Errorf("select count users: %w", err))
	}

	return countUsers, nil
}

// принимает ID и мапу с полями и значениями для изменения
// возвращает измененного пользователя и ошибку, если пользователя нет
func (repo *AdminRepository) EditUser(ctx context.Context, userID string, editInfo map[string]string) (*domains.User, error) {
	// Собираем части SET и аргументы
	setClauses := make([]string, 0, len(editInfo))
	args := make([]any, 0, len(editInfo)+1)
	i := 1
	for field, value := range editInfo {
		// формируем массив типа ["login = $1", "role = $2"]
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		// добавляем аргументы в отдельный массив для передачи
		args = append(args, value)
		i++
	}
	args = append(args, userID) // чтобы потом распаковать

	sqlQuery := fmt.Sprintf(
		`
		UPDATE users SET %s 
		WHERE id = $%d 
		RETURNING id, login, full_name, position, role, created_at;
		`,
		strings.Join(setClauses, ", "), // формируем строку типа "login = $1, role = $2"
		i,                              // передаем индекс параметра для ID
	)

	var user domains.User

	// передаем строку, запрос вместе с параметрами
	err := repo.pool.QueryRow(ctx, sqlQuery, args...).Scan(
		&user.ID,
		&user.Login,
		&user.FullName,
		&user.Position,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("Пользователь, над которым хотят совершить действие не найден")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("edit user: %w", err))
	}

	return &user, nil

}

// принимает хеш пароля
// меняет его и возвращает ошибку
func (repo *AdminRepository) ResetPassword(ctx context.Context, userID, newPasswordHash string) error {
	sqlQuery := `
	UPDATE users
	SET password_hash = $1
	WHERE id = $2;
	`
	if _, err := repo.pool.Exec(ctx, sqlQuery, newPasswordHash, userID); err != nil {
		return error_type.NewInternal(fmt.Errorf("reset password hash: %w", err))
	}

	return nil
}

// принимает id пользователя
// удаляет его, если его нет в базе данных, то возвращает ошибку
func (repo *AdminRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := repo.pool.Exec(ctx, query, userID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("delete user: %w", err))
	}

	rowsAffected := result.RowsAffected() // возвращает количество удаленных строк, если ноль
	if rowsAffected == 0 {                // то пользователь не найден
		return error_type.NewNotFound("пользователь не найден")
	}

	return nil
}

// ====================================================== МЕТОДЫ РЕПОЗИТОРИЯ ДЛЯ ГРУПП ==============================================
