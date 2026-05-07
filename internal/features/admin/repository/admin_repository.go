package repository

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"context"
	"errors"
	"fmt"

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

// ищет пользователя по ID
// возвращает информацию о нем
func (repo *AdminRepository) SelectUserByID(ctx context.Context, userID string) (*domains.User, error) {
	sqlQuery := `
	SELECT id, email, full_name, position, role, created_at
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
// возвращает его id и время создания
func (repo *AdminRepository) InsertNewUser(
	ctx context.Context,
	login, passwordHash, fullname, position, role string,
) (string, error) {
	sqlQuery := `
	INSERT INTO, email, password_hash, fullname, position, role
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at;
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
	SELECT id, email, full_name, position, role, created_at
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

	var usersInfo []domains.User

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
	SELECT id, email, full_name, position, role, created_at
	FROM users
	WHERE role = "user"
	ORDER BY full_name ASC 
	LIMIT $2
	OFFSET ($1 - 1) * $2;
	`

	
	rows, err := repo.pool.Query(ctx, sqlQuery, page, limit)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users info: %w", err))
	}

	var usersInfo []domains.User

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