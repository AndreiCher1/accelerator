package service

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"accelerator/internal/features/admin/repository"
	"accelerator/internal/tools"
	"context"
	"fmt"
)

type AdminService struct {
	repo *repository.AdminRepository
	cfg  *config.Config
}

func NewAdminService(repo *repository.AdminRepository, cfg *config.Config) *AdminService {
	return &AdminService{
		repo: repo,
		cfg:  cfg,
	}
}

// ====================================================== СОЗДАНИЕ КРЕАТОРА ==============================================

func (serv *AdminService) AddCreatorService(ctx context.Context, login, password, fullname, position string) (*domains.User, error) {
	// проверяем, если база данных пользователей пустая, значит можно создать креатора
	isEmptyUsers, err := serv.repo.IsTableUsersEmpty(ctx)
	if err != nil {
		return nil, err
	}

	// если она не пустая, то запрещаем доступ и кидаем 404
	if !isEmptyUsers {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	// хешируем его
	passwordHash, err := tools.GeneratePasswordHash(password)
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("generate password hash: %w", err))
	}

	// добавляем нового пользователя с ролью креатор
	// добавление данных в репозиторий, если логин уже есть, то ошибка 409
	userID, err := serv.repo.InsertNewCreator(ctx, login, passwordHash, fullname, position)
	if err != nil {
		return nil, err
	}

	userInfo := domains.User{
		ID:        userID,
		Login:     login,
		FullName:  fullname,
		Position:  position,
		Role:      "creator",
	}

	return &userInfo, nil

}

// ====================================================== СЕРВИСЫ ДЛЯ ВЗАИМОДЕЙСТВИЯ С ПОЛЬЗОВАТЕЛЯМИ ==============================================

// ====================== РЕГИСТРАЦИЯ НОВОГО РОЛЬЗОВАТЕЛЯ =======================

// возвращает домен юзера и сгенерированный пароль
func (serv *AdminService) RegisterNewUserService(ctx context.Context, callerID, login, fullname, position, role string) (*domains.User, string, error) {
	// получаем по ID из токена информацию о том, кто делает запрос
	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil { // если пользователь не найден, возвращаем ошибку
		return nil, "", err
	}

	// если креатор или админ, юзер не может тут ничего делать
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return nil, "", error_type.NewNotFound("Страница не найдена") // не раскрываем существование ресурса
	}

	// админ может добавлять только юзеров
	if callerUser.Role == "admin" && role != "user" {
		return nil, "", error_type.NewForbidden()
	}

	// если мы дошли сюда, значит мы креатор, а для него нет ограничений
	// в роли либо user будет либо admin

	// генерируем пароль
	password, err := tools.GeneratePassword(serv.cfg)
	if err != nil {
		return nil, "", error_type.NewInternal(fmt.Errorf("generate password: %w", err))
	}

	// хешируем его
	passwordHash, err := tools.GeneratePasswordHash(password)
	if err != nil {
		return nil, "", error_type.NewInternal(fmt.Errorf("generate password hash: %w", err))
	}

	// добавление данных в репозиторий, если логин уже есть, то ошибка 409
	userID, err := serv.repo.InsertNewUser(ctx, login, passwordHash, fullname, position, role)
	if err != nil {
		return nil, "", err
	}

	userInfo := domains.User{
		ID:                userID,
		Login:             login,
		FullName:          fullname,
		Position:          position,
		Role:              role,
	}

	return &userInfo, password, nil
}

// ====================== ПОЛУЧЕНИЕ ПОЛЬЗОВАТЕЛЕЙ =======================

// Возвращает список юзеров, их количество и ошибку
func (serv *AdminService) GetUsersService(ctx context.Context, callerID string, page, limit int) (*[]domains.User, int64, error) {
	// получаем по ID из токена информацию о том, кто делает запрос
	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil { // если пользователь не найден, возвращаем ошибку
		return nil, 0, err
	}

	// если креатор или админ, юзер не может тут ничего делать
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return nil, 0, error_type.NewNotFound("Страница не найдена")
	}

	fmt.Println(callerUser)

	var users *[]domains.User
	var countUser int64

	if callerUser.Role == "creator" {
		// вызываем репозиторий, он возвращает всех пользователей и их количество
		users, err = serv.repo.SelectAllUsersWithAdminFirst(ctx, page, limit)
		if err != nil {
			return nil, 0, err
		}
		countUser, err = serv.repo.SelectCountUsersAndAdmins(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	if callerUser.Role == "admin" {
		// вызываем репозиторий, он возвращает всех c role=user
		users, err = serv.repo.SelectOnlyUsers(ctx, page, limit)
		if err != nil {
			return nil, 0, err
		}
		countUser, err = serv.repo.SelectCountUsers(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	// возвращаем список и общее количество
	return users, countUser, nil
}

// ======================== ИЗМЕНЕНИЕ ПОЛЬЗОВАТЕЛЯ =========================
// возвращает пользователя и ошибку
func (serv *AdminService) EditUserService(ctx context.Context, callerID, targetID string, editInfo map[string]string) (*domains.User, error) {
	// получаем по ID из токена информацию о том, кто делает запрос
	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil { // если пользователь не найден, возвращаем ошибку
		return nil, err
	}

	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return nil, error_type.NewNotFound("Страница не найдена")
	}

	// получаем юзера, над которым собираются совершать манипуляции
	// ошибка, если его не существует
	targetUser, err := serv.repo.SelectUserByID(ctx, targetID)
	if err != nil { // если пользователь не найден, возвращаем ошибку
		return nil, err
	}

	// если админ, то не может изменять никого кроме юзеров
	if callerUser.Role == "admin" && targetUser.Role != "user" {
		return nil, error_type.NewNotFound("Страница не найдена") // если он сюда попал, значит как-то узнал ID админа, скрываем информацию
	}

	// запрещаем менять самому себе роль ради безопасности
	if callerUser.Role == "creator" && targetUser.Role == "creator" {
		return nil, error_type.NewForbidden()
	}

	// админ не может менять роли
	if callerUser.Role == "admin" && editInfo["role"] != "" { // если ключа нет, вернется по умолчанию пустая строка
		return nil, error_type.NewForbidden()
	}

	// креатор меняет на user и admin, в транспорте есть валидация

	// передаем editInfo для изменения в репозиторий
	// внутри репозитория динамически собираем запрос исходя из изменяемых аргументов
	// до этого уже проверяли существование targetUser, поэтому можно опустить
	editUser, err := serv.repo.EditUser(ctx, targetUser.ID, editInfo)
	if err != nil {
		return nil, err
	}

	// репозиторий возвращает доменную структуру, сразу ее возвращаем в транспорт
	return editUser, nil

}

// ====================== СБРОСИТЬ ПАРОЛЬ ДЛЯ ПОЛЬЗОВАТЕЛЯ =======================
// возвращает новый пароль и ошибку
func (serv *AdminService) ResetPasswordService(ctx context.Context, callerID, targetID string) (string, error) {
	// получаем по ID из токена информацию о том, кто делает запрос
	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return "", err
	}

	// если креатор или админ, юзер не может тут ничего делать
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return "", error_type.NewNotFound("Страница не найдена")
	}

	// получаем юзера, над которым собираются совершать манипуляции
	// ошибка, если его не существует
	targetUser, err := serv.repo.SelectUserByID(ctx, targetID)
	if err != nil {
		return "", err
	}

	// если админ, то не может сбрасывать пороль никому кроме юзеров
	if callerUser.Role == "admin" && targetUser.Role != "user" {
		return "", error_type.NewNotFound("Страница не найдена") // значит он как-то попал на админа, скрываем информацию
	}

	// креатор может всем сбрасывать

	// генерируем пароль
	password, err := tools.GeneratePassword(serv.cfg)
	if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("generate password: %w", err))
	}

	// хешируем его
	passwordHash, err := tools.GeneratePasswordHash(password)
	if err != nil {
		return "", error_type.NewInternal(fmt.Errorf("generate password hash: %w", err))
	}

	// обновляем данные о пароле в репозитории
	if err := serv.repo.ResetPassword(ctx, targetUser.ID, passwordHash); err != nil {
		return "", nil
	}

	// возвращаем новый пароль
	return password, nil
}



// ====================== УДАЛИТЬ ПОЛЬЗОВАТЕЛЯ =======================
func (serv *AdminService) DeleteUserService(ctx context.Context, callerID, targetID string) error {
	// получаем по ID из токена информацию о том, кто делает запрос
	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return err
	}

	// если креатор или админ, юзер не может тут ничего удалять
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return error_type.NewNotFound("Страница не найдена")
	}

	// нельзя удалить самого себя как креатору, так и админу
	if callerID == targetID {
		return error_type.NewForbidden()
	}

	// получаем юзера, над которым собираются совершать манипуляции
	// ошибка, если его не существует
	targetUser, err := serv.repo.SelectUserByID(ctx, targetID)
	if err != nil {
		return err
	}

	// если админ, то может удалять только юзеров
	if callerUser.Role == "admin" && targetUser.Role != "user" {
		return error_type.NewNotFound("Страница не найдена") // админ не должен видеть других админов и креатора
	}

	// будет обернуто в транзакцию { !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//		 удаляем пользователя через репозиторий
	// 		 удаляем все его сессии
	// }
	// пока не надо, делаем жесткое удаление пользователя, если его удалили, значит и его сессии удалятся каскадом
	if err := serv.repo.DeleteUser(ctx, targetID); err != nil {
		return err
	}

	return nil
}

// // ====================================================== СЕРВИСЫ ВЗАИМОДЕЙСТВИЯ С ГРУППАМИ ==============================================

// // ========================= СОЗДАТЬ ГРУППУ ==========================

// // возвращает информацию о группе и ошибку
// func (serv *AdminService) CreateGroupService(callerID, name, description string) (domains.Group, error) {
// 	// получаем по ID из токена информацию о том, кто делает запрос
// 	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return error_type.NewForbidden()
// 	}

// 	// в репозитории добавляем группу с created_by = caller.id

// 	// потом ищем креатора в репозитории

// 	// В group_members для GroupID добавляются callerID и creatorID

// 	// вычисляем флаги

// 	// репозиторий возвратил группу, мы ее передаем в транспорт
// }

// // ========================= ПОЛУЧИТЬ УЧАСТНИКОВ ГРУППЫ  ==========================
// // возвращает слайc user и ошибку
// func (serv *AdminService) GetMembersGroupService(callerID, groupID string) (*[]domains.User, error) {
// 	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return nil, error_type.NewForbidden()
// 	}

// 	// креатор видит всех, кроме себя
// 	if callerUser.Role == "creator" {
// 		// репозиторий where role="user" OR role="admin"
// 	}

// 	// админ видит всех user
// 	if callerUser.Role == "admin" {
// 		// проверить, что callerUser.ID есть в этой группе, а потом

// 		// репозиторий where role="user"
// 	}

// 	// вычисляем флаги

// 	// из репозитория получили слайс доменов, сразу кидаем в транспорт

// }

// // ========================= ПОЛУЧИТЬ ВСЕ ГРУППЫ  ==========================
// // возвращает слайc group и ошибку
// func (serv *AdminService) GetGroupsService(callerID string) (*[]domains.Group, error) {
// 	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return nil, error_type.NewForbidden()
// 	}

// 	// креатор видит все группы
// 	if callerUser.Role == "creator" {
// 		// в репозитории возвращаем все группы
// 	}

// 	// админ видит все группы, в которых состоит, т е у которые добавил его креатор
// 	if callerUser.Role == "admin" {
// 		// репозиторий вовзращает все группы, где callerUser.ID есть в group_members
// 	}

// 	// вычисляем флаги для каждой группы и делаем массив из них

// 	// из репозитория получили слайс доменов групп, сразу кидаем в транспорт
// }

// // ======================== ИЗМЕНИТЬ ИНФОРМАЦИЮ О ГРУППЕ  ==========================

// func (serv *AdminService) EditGroupService(callerID, groupID string, editInfo map[string]string) (*domains.Group, error) {
// 	callerUser, err := serv.repo.SelectUserByID(ctx, callerID)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	// если креатор или админ, юзер не может тут ничего изменять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return nil, error_type.NewForbidden()
// 	}

// 	// по groupID проверить, существует ли группа в репозитории
// 	groupInfo :=

// 	if callerUser.Role != "creator" {
// 		// вызываем репозиторий, передаем editInfo, можем изменять любую группу
// 	}

// 	if callerUser.Role != "admin" {
// 		// проверяем, что в группе полученной groupInfo.CreatedBy == callerUser.ID
// 		// вызываем репозиторий, передаем editInfo
// 	}

// 	// вычисляем флаги

// 	// получили данные, кидаем в транспорт, там собираем ответ

// }

// // ====================== ДОБАВИТЬ ПОЛЬЗОВАТЕЛЯ В ГРУППУ  ==========================

// func (serv *AdminService) AddUserGroupService(callerID, targetID, groupID string) (error) {
// 	// получаем по ID из токена информацию о том, кто делает запрос
// 	// ошибка, если не существует
// 	callerUser :=

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return error_type.NewForbidden()
// 	}

// 	// получаем юзера, над которым собираются совершать манипуляции
// 	// ошибка, если его не существует
// 	targetUser :=

// 	// по groupID проверить, существует ли группа в репозитории
// 	groupInfo :=

// 	// если креатор, то кого угодно куда угодно
// 	if callerUser.Role == "creator" {
// 		// вызываем репозиторий без каких либо ограничений
// 		// внутри репозиторий проверяет, если есть таргет, уже в группе, то 409
// 	}

// 	if callerUser.Role == "admin" {
// 		// проверяем, что targetUser.Role == "user"

// 		// проверяем, что callerUser состоит в этой группе в group_members where group_id=groupInfo.ID AND user_id=callerUser.ID

// 		// если все ок, кидаем запрос в репозиторий
// 		// внутри репозиторий проверяет, если есть таргет, уже в группе, то 409
// 	}

// 	// с репозитория только ошибка, мы ее обрабатываем и отсылаем в транспорт

// }

// // ====================== УДАЛИТЬ ПОЛЬЗОВАТЕЛЯ ИЗ ГРУППЫ  ==========================

// func (serv *AdminService) DeleteUserGroupService(callerID, targetID, groupID string) (error) {
// 	// получаем по ID из токена информацию о том, кто делает запрос
// 	// ошибка, если не существует
// 	callerUser :=

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return error_type.NewForbidden()
// 	}

// 	// получаем юзера, над которым собираются совершать манипуляции
// 	// ошибка, если его не существует
// 	targetUser :=

// 	// нельзя самого себя удалить из группы
// 	if targetUser.ID == callerUser.ID {
// 		return error_type.NewForbidden()
// 	}

// 	// по groupID проверить, существует ли группа в репозитории
// 	// ошибка, если нет
// 	groupInfo :=

// 	// если креатор, то кого угодно куда угодно
// 	if callerUser.Role == "creator" {
// 		// вызываем репозиторий без каких либо ограничений
// 		// если target не состоит в группе, то 400 вовзращает репозиторий
// 	}

// 	if callerUser.Role == "admin" {
// 		// проверяем, что targetUser.Role == "user"

// 		// проверяем, что callerUser состоит в этой группе в group_members where group_id=groupInfo.ID AND user_id=callerUser.ID

// 		// если все ок, кидаем запрос в репозиторий
// 		// если target не состоит в группе, то 400 вовзращает репозиторий
// 	}

// 	// с репозитория только ошибка, мы ее обрабатываем и отсылаем в транспорт
// }

// // ============================ УДАЛИТЬ ГРУППУ  ==============================

// func (serv *AdminService) DeleteGroupService(callerID, groupID string) (error) {
// 	// получаем по ID из токена информацию о том, кто делает запрос
// 	// ошибка, если не существует
// 	callerUser :=

// 	// если креатор или админ, юзер не может тут ничего удалять
// 	if callerUser.Role != "creator" && callerUser.Role != "admin" {
// 		return error_type.NewForbidden()
// 	}

// 	// по groupID проверить, существует ли группа в репозитории
// 	// ошибка, если нет
// 	groupInfo :=

// 	// если креатор, то кого угодно куда угодно
// 	if callerUser.Role == "creator" {
// 		// вызываем репозиторий без каких либо ограничений
// 	}

// 	if callerUser.Role == "admin" {

// 		// проверяем, что callerUser.ID == groupInfo.CreatedBy

// 		// если все ок, кидаем запрос в репозиторий
// 		// если target не состоит в группе, то 400 вовзращает репозиторий
// 	}

// 	// с репозитория только ошибка, мы ее обрабатываем и отсылаем в транспорт
// }
