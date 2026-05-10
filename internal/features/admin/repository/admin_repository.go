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
	query := `SELECT EXISTS (SELECT 1 FROM users);`
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

// для операций с транзакциями
type executor interface { 
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (repo *AdminRepository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return repo.pool.BeginTx(ctx, opts)
}


// возвращает ID креатора
// если его нет в бд, возвращает ошибку
func (repo *AdminRepository) SelectCreatorID(ctx context.Context) (string, error) {
	sqlQuery := `
	SELECT id
	FROM users
	WHERE role = 'creator';
	`
	var creatorID string

	err := repo.pool.QueryRow(ctx, sqlQuery).Scan(&creatorID)
	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return "", error_type.NewInternal(fmt.Errorf("creator ID not found")) // возвращаем 500, чтобы тот, кто добвлял в группу не мог понять, что существует креатор
	} else if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("select creator ID: %w", err))
	}

	return creatorID, nil
}

// создает группу
// возвращает ее ID и ошибку
func (repo *AdminRepository) createGroup(ctx context.Context, e executor, name, description, createdByID string) (string, error) {
	sqlQuery := `
	INSERT INTO groups (name, description, created_by)
	VALUES ($1, $2, $3)
	RETURNING id;
	`

	var groupID string

	if err := e.QueryRow(ctx, sqlQuery, name, description, createdByID).Scan(&groupID); err != nil {
		return "", error_type.NewInternal(fmt.Errorf("create group: %w", err))
	}

	return groupID, nil
}

func (repo *AdminRepository) CreateGroup(ctx context.Context, name, description, createdByID string) (string, error) {
	return repo.createGroup(ctx, repo.pool, name, description, createdByID)
}

func (repo *AdminRepository) CreateGroupTx(ctx context.Context, e executor, name, description, createdByID string) (string, error) {
	return repo.createGroup(ctx, e, name, description, createdByID)
}

// Добавляет существующего пользователя в группу по ID
// до этого нужно проверить, существует ли пользователь
// возвращает ошибку
func (repo *AdminRepository) insertUserIntoGroup(ctx context.Context, e executor, groupID, userID string) error {
	sqlQuery := `
	INSERT INTO group_members (group_id, user_id)
	VALUES ($1, $2)
	`

	if _, err := e.Exec(ctx, sqlQuery, groupID, userID); err != nil {
		return error_type.NewInternal(fmt.Errorf("add group member: %w", err))
	}

	return nil
}

func (repo *AdminRepository) InsertUserIntoGroup(ctx context.Context, groupID, userID string) error {
	return repo.insertUserIntoGroup(ctx, repo.pool, groupID, userID)
}

func (repo *AdminRepository) InsertUserIntoGroupTx(ctx context.Context, e executor, groupID, userID string) error {
	return repo.insertUserIntoGroup(ctx, e, groupID, userID)
}

// принимает ID группы
// возвращает список пользователей, который видит админ
// если группа пуста, возвращает пустой список
func (repo *AdminRepository) SelectOnlyUsersIntoGroup(ctx context.Context, groupID string) (*[]domains.User, error) {
	sqlQuery := `
	SELECT u.id, u.login, u.full_name, u.position, u.role, u.created_at
	FROM users u
	JOIN group_members gm ON u.id = gm.user_id  -- связывам строки из user и members
	WHERE gm.group_id = $1 AND u.role = 'user'  -- получаем только те из них, которые в группе
	ORDER BY u.full_name ASC;
	`

	rows, err := repo.pool.Query(ctx, sqlQuery, groupID)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users info into group: %w", err))
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
			return nil, error_type.NewInternal(fmt.Errorf("scan user into group: %w", err))
		}

		usersInfo = append(usersInfo, userInfo)
	}

	return &usersInfo, nil
}


// принимает ID группы
// возвращает список пользователей, который видит креатор
// если группа пуста, возвращает пустой список
func (repo *AdminRepository) SelectUsersWithAdminFirstIntoGroup(ctx context.Context, groupID string) (*[]domains.User, error) {
	sqlQuery := `
		SELECT u.id, u.login, u.full_name, u.position, u.role, u.created_at
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id      -- связывам строки из user и members
		WHERE gm.group_id = $1 AND u.role != 'creator'  -- получаем только те из них, которые в группе
		ORDER BY
			CASE u.role 
				WHEN 'admin' THEN 1  -- сначала по админу
				WHEN 'user' THEN 2 -- потом все юзеры
			END,
		full_name ASC; -- а эти группы по алфавиту
	`

	rows, err := repo.pool.Query(ctx, sqlQuery, groupID)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users or admins info into group: %w", err))
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
			return nil, error_type.NewInternal(fmt.Errorf("scan user or admin into group: %w", err))
		}

		usersInfo = append(usersInfo, userInfo)
	}

	return &usersInfo, nil
}


// проверяет, состоит ли пользователь в группе
// true - состоит, false - нет
func (repo *AdminRepository) IsUserIntoGroup(ctx context.Context, userID, groupID string) (bool, error) {
    query := `
        SELECT EXISTS (
            SELECT 1 FROM group_members
            WHERE user_id = $1 AND group_id = $2
        )`
    
    var consists bool
    err := repo.pool.QueryRow(ctx, query, userID, groupID).Scan(&consists)
    if err != nil {
        return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
    }
    return consists, nil
}


// ищет пользователя по ID
// возвращает информацию о нем
func (repo *AdminRepository) SelectGroupInfo(ctx context.Context, groupID string) (*domains.Group, error) {
	sqlQuery := `
		SELECT id, name, description, created_by, created_at
		FROM groups
		where id = $1;
	`

	var groupInfo domains.Group
	err := repo.pool.QueryRow(ctx, sqlQuery, groupID).Scan(
		&groupInfo.GroupID,
		&groupInfo.Name,
		&groupInfo.Description,
		&groupInfo.CreatedBy,
		&groupInfo.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("группа, над которой хотят совершить действие не найдена")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get group info: %w", err))
	}

	return &groupInfo, nil
}


// возвращает список групп, который видит креатор
// если нет групп, то возвращает пустой список
func (repo *AdminRepository) SelectGroupsForCreator(ctx context.Context) (*[]domains.Group, error) {
	sqlQuery := `
		SELECT 
			g.id,
			g.name,
			g.description,
			g.created_by,
			g.created_at,
		COALESCE(COUNT(gm.user_id), 0) - 1 AS member_count
		FROM groups g
		GROUP BY g.id, g.name, g.description, g.created_by, g.created_at
		ORDER BY g.name ASC;
	`

	rows, err := repo.pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get users or admins info into group: %w", err))
	}

	groupsInfo := make([]domains.Group, 0)

	for rows.Next() {
		var groupInfo domains.Group

		if err := rows.Scan(
			&groupInfo.GroupID,
			&groupInfo.Name,
			&groupInfo.Description,
			&groupInfo.CreatedBy,
			&groupInfo.CreatedAt,
			&groupInfo.MemberCount,
		); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan user or admin into group: %w", err))
		}

		groupsInfo = append(groupsInfo, groupInfo)
	}

	return &groupsInfo, nil
}

// возвращает список групп, в которых состоит администратор,
// а также количество участников с ролью 'user' в каждой группе.
func (repo *AdminRepository) SelectGroupsForAdmin(ctx context.Context, adminID string) (*[]domains.Group, error) {
    query := `
        SELECT 
            g.id,
            g.name,
            g.description,
            g.created_by,
            g.created_at,
            COUNT(u_user.id) AS member_count
        FROM groups g
        JOIN group_members gm_admin ON gm_admin.group_id = g.id AND gm_admin.user_id = $1
        LEFT JOIN group_members gm_all ON gm_all.group_id = g.id
        LEFT JOIN users u_user ON gm_all.user_id = u_user.id AND u_user.role = 'user'
        GROUP BY g.id, g.name, g.description, g.created_by, g.created_at
        ORDER BY g.name ASC
    `
    rows, err := repo.pool.Query(ctx, query, adminID)
    if err != nil {
        return nil, error_type.NewInternal(fmt.Errorf("query groups for admin: %w", err))
    }
    defer rows.Close()

    var groups []domains.Group
    for rows.Next() {
        var g domains.Group
        if err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt, &g.MemberCount); err != nil {
            return nil, error_type.NewInternal(fmt.Errorf("select groups for admin: %w", err))
        }
        groups = append(groups, g)
    }
    return &groups, nil
}



// принимает ID и мапу с полями и значениями для изменения
// возвращает измененную группу и ошибку, если группы нет
// также считается количество пользователей для креатора
func (repo *AdminRepository) EditGroupForCreator(ctx context.Context, groupID string, editInfo map[string]string) (*domains.Group, error) {
	// Собираем части SET и аргументы
	setClauses := make([]string, 0, len(editInfo))
	args := make([]any, 0, len(editInfo)+1)
	i := 1
	for field, value := range editInfo {
		// формируем массив типа ["name = $1", "description = $2"]
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		// добавляем аргументы в отдельный массив для передачи
		args = append(args, value)
		i++
	}
	args = append(args, groupID) // чтобы потом распаковать

	sqlQuery := fmt.Sprintf(
		`
		UPDATE groups SET %s 
		WHERE id = $%d 
		RETURNING id, name, description, created_by, created_at,
    		(SELECT COUNT(*) - 1 FROM group_members WHERE group_id = groups.id) AS member_count;
		`,
		strings.Join(setClauses, ", "), // формируем строку типа "login = $1, role = $2"
		i,                              // передаем индекс параметра для ID
	)

	var group domains.Group

	// передаем строку, запрос вместе с параметрами
	err := repo.pool.QueryRow(ctx, sqlQuery, args...).Scan(
		&group.GroupID,
		&group.Name,
		&group.Description,
		&group.CreatedBy,
		&group.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("группа, над которой хотят совершить действие не найдена")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("edit user: %w", err))
	}

	return &group, nil
}


// принимает ID и мапу с полями и значениями для изменения
// возвращает измененную группу и ошибку, если группы нет
// также считается количество пользователей в группе для админа
func (repo *AdminRepository) EditGroupForAdmin(ctx context.Context, groupID string, editInfo map[string]string) (*domains.Group, error) {
	// Собираем части SET и аргументы
	setClauses := make([]string, 0, len(editInfo))
	args := make([]any, 0, len(editInfo)+1)
	i := 1
	for field, value := range editInfo {
		// формируем массив типа ["name = $1", "description = $2"]
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		// добавляем аргументы в отдельный массив для передачи
		args = append(args, value)
		i++
	}
	args = append(args, groupID) // чтобы потом распаковать

	sqlQuery := fmt.Sprintf(
		`
		UPDATE groups SET %s 
		WHERE id = $%d 
		RETURNING id, name, description, created_by, created_at,
    		(SELECT COUNT(*) - 1 FROM group_members WHERE group_id = groups.id) AS member_count;
		`,
		strings.Join(setClauses, ", "), // формируем строку типа "login = $1, role = $2"
		i,                              // передаем индекс параметра для ID
	)

	var group domains.Group

	// передаем строку, запрос вместе с параметрами
	err := repo.pool.QueryRow(ctx, sqlQuery, args...).Scan(
		&group.GroupID,
		&group.Name,
		&group.Description,
		&group.CreatedBy,
		&group.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) { // специальный тип ошибки, если ничего не вернулось
		return nil, error_type.NewNotFound("группа, над которой хотят совершить действие не найдена")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("edit user: %w", err))
	}

	return &group, nil
}