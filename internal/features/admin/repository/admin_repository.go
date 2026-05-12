package repository

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"context"
	"database/sql"
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
	sqlQuery := `DELETE FROM users WHERE id = $1`

	result, err := repo.pool.Exec(ctx, sqlQuery, userID)
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

type executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (repo *AdminRepository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return repo.pool.BeginTx(ctx, opts)
}

// ---------------- ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ----------------

// возвращает ID единственного креатора
func (repo *AdminRepository) SelectCreatorID(ctx context.Context) (string, error) {
	const query = `SELECT id FROM users WHERE role = 'creator'`
	var creatorID string
	err := repo.pool.QueryRow(ctx, query).Scan(&creatorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", error_type.NewInternal(fmt.Errorf("creator ID not found"))
	} else if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("select creator ID: %w", err))
	}
	return creatorID, nil
}

// ---------------- ГРУППЫ: СОЗДАНИЕ, ИЗМЕНЕНИЕ, УДАЛЕНИЕ ----------------

// универсальная функция для создания группы (через pool или tx)
func (repo *AdminRepository) createGroup(ctx context.Context, e executor, name, description, createdByID string, ownerID any) (string, error) {
	query := `
        INSERT INTO groups (name, description, created_by, owner_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `

	var groupID string
	err := e.QueryRow(ctx, query, name, description, createdByID, ownerID).Scan(&groupID)
	if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("create group: %w", err))
	}
	return groupID, nil
}

func (repo *AdminRepository) CreateGroup(ctx context.Context, name, description, createdByID string, ownerID any) (string, error) {
	return repo.createGroup(ctx, repo.pool, name, description, createdByID, ownerID)
}

func (repo *AdminRepository) CreateGroupTx(ctx context.Context, tx pgx.Tx, name, description, createdByID string, ownerID any) (string, error) {
	return repo.createGroup(ctx, tx, name, description, createdByID, ownerID)
}

// меняет владельца группы (owner_id)
func (repo *AdminRepository) UpdateGroupOwner(ctx context.Context, groupID string, newOwnerID *string) error {
	query := `UPDATE groups SET owner_id = $2 WHERE id = $1`
	_, err := repo.pool.Exec(ctx, query, groupID, newOwnerID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("update group owner: %w", err))
	}
	return nil
}

// ---------------- ЧЛЕНСТВО В ГРУППАХ ----------------

// добавляет пользователя в группу
func (repo *AdminRepository) insertUserIntoGroup(ctx context.Context, e executor, groupID, userID string) error {
	query := `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`
	_, err := e.Exec(ctx, query, groupID, userID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("add group member: %w", err))
	}
	return nil
}

func (repo *AdminRepository) InsertUserIntoGroup(ctx context.Context, groupID, userID string) error {
	return repo.insertUserIntoGroup(ctx, repo.pool, groupID, userID)
}

func (repo *AdminRepository) InsertUserIntoGroupTx(ctx context.Context, tx pgx.Tx, groupID, userID string) error {
	return repo.insertUserIntoGroup(ctx, tx, groupID, userID)
}

// проверяет членство в группе = true - есть в группе
func (repo *AdminRepository) IsUserIntoGroup(ctx context.Context, userID, groupID string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM group_members WHERE user_id = $1 AND group_id = $2)`
	var exists bool
	err := repo.pool.QueryRow(ctx, query, userID, groupID).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
	}
	return exists, nil
}

// проверяет наличие админа в группе = true - есть
func (repo *AdminRepository) GroupHasAdmin(ctx context.Context, groupID string) (bool, error) {
	sqlQuery := `
		SELECT EXISTS(
			SELECT 1 FROM group_members gm
			JOIN users u ON u.id = gm.user_id
			WHERE gm.group_id = $1 AND u.role = 'admin'
		)
	`
	var exists bool
	err := repo.pool.QueryRow(ctx, sqlQuery, groupID).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
	}
	return exists, nil
}

// ---------------- ПОЛУЧЕНИЕ ИНФОРМАЦИИ О ГРУППАХ И УЧАСТНИКАХ ----------------

// возвращает информацию о группе, включая owner_id
func (repo *AdminRepository) SelectGroupInfo(ctx context.Context, groupID string) (*domains.Group, error) {
	query := `
        SELECT id, name, description, created_by, owner_id, created_at,
               (SELECT COUNT(*) - 1 FROM group_members WHERE group_id = groups.id) AS member_count
        FROM groups
        WHERE id = $1
    `
	var g domains.Group
	var ownerID sql.NullString // чтобы считать либо null либо строку
	err := repo.pool.QueryRow(ctx, query, groupID).Scan(
		&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &ownerID, &g.CreatedAt, &g.MemberCount,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, error_type.NewNotFound("группа не найдена")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("get group info: %w", err))
	}

	// если nil преобразует в пустую строку, чтобы в дальнейшем нигде в сервисе случайно не разыменовать указаатель на nil,
	// если бы я хранил *string, а так я храню string и если нет, то ""
	g.OwnerID = ownerID.String

	return &g, nil
}

// список участников с ролью 'user' для админа
func (repo *AdminRepository) SelectOnlyUsersIntoGroup(ctx context.Context, groupID string) (*[]domains.User, error) {
	query := `
        SELECT u.id, u.login, u.full_name, u.position, u.role, gm.added_at
        FROM users u
        JOIN group_members gm ON u.id = gm.user_id
        WHERE gm.group_id = $1 AND u.role = 'user'
        ORDER BY u.full_name ASC
    `
	return repo.fetchUsers(ctx, query, groupID)
}

// список участников для креатора (сначала админ, потом юзеры)
func (repo *AdminRepository) SelectUsersWithAdminFirstIntoGroup(ctx context.Context, groupID string) (*[]domains.User, error) {
	query := `
        SELECT u.id, u.login, u.full_name, u.position, u.role, gm.added_at
        FROM users u
        JOIN group_members gm ON u.id = gm.user_id
        WHERE gm.group_id = $1 AND u.role != 'creator'
        ORDER BY CASE u.role WHEN 'admin' THEN 1 WHEN 'user' THEN 2 END, u.full_name ASC
    `
	return repo.fetchUsers(ctx, query, groupID)
}

func (repo *AdminRepository) fetchUsers(ctx context.Context, query, groupID string) (*[]domains.User, error) {
	rows, err := repo.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("query users: %w", err))
	}
	defer rows.Close()

	users := make([]domains.User, 0)
	for rows.Next() {
		var u domains.User
		if err := rows.Scan(&u.ID, &u.Login, &u.FullName, &u.Position, &u.Role, &u.AddedAt); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan user: %w", err))
		}
		users = append(users, u)
	}
	return &users, nil
}

// все группы для креатора (с owner_id)
func (repo *AdminRepository) SelectGroupsForCreator(ctx context.Context) (*[]domains.Group, error) {
	query := `
        SELECT g.id, g.name, g.description, g.created_by, g.owner_id, g.created_at,
               COALESCE(COUNT(gm.user_id), 0) - 1 AS member_count
        FROM groups g
        LEFT JOIN group_members gm ON g.id = gm.group_id
        GROUP BY g.id
        ORDER BY g.name ASC
    `
	return repo.fetchGroups(ctx, query)
}

// группы, в которых состоит админ, с owner_id
func (repo *AdminRepository) SelectGroupsForAdmin(ctx context.Context, adminID string) (*[]domains.Group, error) {
	query := `
        SELECT g.id, g.name, g.description, g.created_by, g.owner_id, g.created_at,
               COALESCE(user_counts.cnt, 0) AS member_count
        FROM groups g
        JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = $1
        LEFT JOIN (
            SELECT group_id, COUNT(*) as cnt
            FROM group_members gm2
            JOIN users u ON gm2.user_id = u.id AND u.role = 'user'
            GROUP BY group_id
        ) user_counts ON g.id = user_counts.group_id
        ORDER BY g.name ASC
    `
	return repo.fetchGroups(ctx, query, adminID)
}

func (repo *AdminRepository) fetchGroups(ctx context.Context, query string, args ...any) (*[]domains.Group, error) {
	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("query groups: %w", err))
	}
	defer rows.Close()

	groups := make([]domains.Group, 0)
	for rows.Next() {
		var g domains.Group

		var ownerID sql.NullString // чтобы суметь прочитать NULL
		if err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &ownerID, &g.CreatedAt, &g.MemberCount); err != nil {
			return nil, error_type.NewInternal(fmt.Errorf("scan group: %w", err))
		}

		g.OwnerID = ownerID.String

		groups = append(groups, g)
	}
	return &groups, nil
}

// динамическое обновление полей группы (точка входа для сервиса)
func (repo *AdminRepository) editGroup(ctx context.Context, e executor, groupID string, editInfo map[string]string) (*domains.Group, error) {
	// Фильтруем допустимые поля: name, description, owner_id
	// валидируем в хэндлере, но пусть будет на всякий
	allowed := map[string]bool{"name": true, "description": true, "owner_id": true}
	setClauses := make([]string, 0)
	args := make([]any, 0)
	i := 1
	for field, value := range editInfo {
		// проверяем на допустимость переданное поле
		if !allowed[field] {
			continue
		}

		// не можем в бд добавить пустую строку в owner_id, поэтому передаем nil
		if field == "owner_id" && value == "" {
			args = append(args, nil) // передаём nil
		} else {
			args = append(args, value)
		}

		// собираем массив для SET
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		i++
	}
	// также валидируем в хэндлере, но пусть будет
	// если где-то ошиблись с названиями
	if len(setClauses) == 0 {
		// Если нет допустимых полей, возвращаем текущую информацию о группе без изменений
		return repo.SelectGroupInfo(ctx, groupID)
	}
	// добавляем в конец ID чтобы потом распаковать
	args = append(args, groupID)

	// для админа возвращается на один больше, виксится в сервисе
	query := fmt.Sprintf(`
        UPDATE groups SET %s
        WHERE id = $%d
        RETURNING id, name, description, created_by, owner_id, created_at,
                  (SELECT COUNT(*) - 1 FROM group_members WHERE group_id = groups.id) AS member_count
    `, strings.Join(setClauses, ", "), i)

	var g domains.Group
	var ownerID sql.NullString // чтобы читать NULL

	err := e.QueryRow(ctx, query, args...).Scan(
		&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &ownerID, &g.CreatedAt, &g.MemberCount,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, error_type.NewNotFound("группа не найдена")
	} else if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("edit group: %w", err))
	}

	g.OwnerID = ownerID.String

	return &g, nil
}

func (repo *AdminRepository) EditGroup(ctx context.Context, groupID string, editInfo map[string]string) (*domains.Group, error) {
	return repo.editGroup(ctx, repo.pool, groupID, editInfo)
}

func (repo *AdminRepository) EditGroupTx(ctx context.Context, e executor, groupID string, editInfo map[string]string) (*domains.Group, error) {
	return repo.editGroup(ctx, e, groupID, editInfo)
}

// ------------------------ УДАЛЕНИЕ ПОЛЬЗОВАТЕЛЕЙ И ГРУПП ---------------------

// удаляет пользователя из группы, возвращает ошибку, если его нет в группе
func (repo *AdminRepository) deleteUserFromGroup(ctx context.Context, e executor, groupID, targetID string) error {
	sqlQuery := `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2;
	`

	result, err := repo.pool.Exec(ctx, sqlQuery, groupID, targetID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("delete user into group: %w", err))
	}

	rowsAffected := result.RowsAffected() // возвращает количество удаленных строк, если ноль
	if rowsAffected == 0 {                // то пользователь не найден
		return error_type.NewNotFound("пользователь не найден")
	}

	return nil
}

func (repo *AdminRepository) DeleteUserFromGroup(ctx context.Context, groupID, targetID string) error {
	return repo.deleteUserFromGroup(ctx, repo.pool, groupID, targetID)
}

func (repo *AdminRepository) DeleteUserFromGroupTx(ctx context.Context, e executor, groupID, targetID string) error {
	return repo.deleteUserFromGroup(ctx, e, groupID, targetID)
}

// удаляет пользователя из группы, возвращает ошибку, если его нет в группе
func (repo *AdminRepository) DeleteGroup(ctx context.Context, groupID string) error {
	sqlQuery := `
		DELETE FROM groups
		WHERE id = $1;
	`

	result, err := repo.pool.Exec(ctx, sqlQuery, groupID)
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("delete group: %w", err))
	}

	rowsAffected := result.RowsAffected() // возвращает количество удаленных строк, если ноль
	if rowsAffected == 0 {                // то пользователь не найден
		return error_type.NewNotFound("пользователь не найден")
	}

	return nil
}

// ======================================================================================
// ------------------------- МЕТОДЫ ПРОВЕРКИ ДЛЯ ПОЛЬЗОВАТЕЛЕЙ --------------------------
// ======================================================================================



// проверяет, если пользователь состоит в группах, где уже назначен админ или еще не назначен
// то мы не сможем его повысить до админа
func (repo *AdminRepository) CheckPromoteConflict(ctx context.Context, userID string) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1 FROM group_members gm
			JOIN groups g ON g.id = gm.group_id
			WHERE gm.user_id = $1
			AND (g.owner_id IS NULL OR g.owner_id != $1)
		)
	`
	var exists bool
	err := repo.pool.QueryRow(ctx, sqlQuery, userID).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
	}
	return exists, nil
}

// возвращает true, если удаляемый пользователь является владельцем хоть одной группы
func (repo *AdminRepository) CheckUserIsOwnerConflict(ctx context.Context, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM groups
			WHERE owner_id = $1
		)
    `

	var exists bool
	err := repo.pool.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, error_type.NewInternal(fmt.Errorf("check membership: %w", err))
	}
	return exists, nil
}
