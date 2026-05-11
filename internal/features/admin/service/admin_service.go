package service

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/domains"
	"accelerator/internal/features/admin/repository"
	"accelerator/internal/tools"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		ID:       userID,
		Login:    login,
		FullName: fullname,
		Position: position,
		Role:     "creator",
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
		ID:       userID,
		Login:    login,
		FullName: fullname,
		Position: position,
		Role:     role,
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

	// если хотят повысить пользователя до админа, проверяем, состоит ли пользователь в других группах, где уже есть админ
	if editInfo["role"] == "admin" {
		isConflict, err := serv.repo.CheckPromoteConflict(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if isConflict {
			return nil, error_type.NewConflict(
				"Пользователь, которого хотят повысить до админа, уже состоит в группах, где назначен админ, либо удалите его из этих групп, либо переназначьте его админом этих групп",
			)
		}
	}

	// если же хотят понизить админа до юзера, то он не должен быть владельцем ни одной из групп
	if editInfo["role"] == "user" {
		isConflict, err := serv.repo.CheckUserIsOwnerConflict(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if isConflict {
			return nil, error_type.NewConflict(
				"Пользователь, которого хотят понизить до юзера, назначен админом в одной или нескольких группах, перед удалением переназначьте админа в этих группах",
			)
		}
	}

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

	// если хотят удалить пользователя, который является владельцем какой-либо группы
	isConflict, err := serv.repo.CheckUserIsOwnerConflict(ctx, targetID)
	if err != nil {
		return err
	}
	if isConflict {
		return error_type.NewConflict("Пользователь, которого хотят удалить является владельцем одной или нескольких групп, пожалуйста, назначьте нового владельца перед удалением")
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

// ======================================================
//            ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ======================================================

// getUserByID возвращает пользователя по ID или ошибку, если не найден.
func (serv *AdminService) getUserByID(ctx context.Context, userID string) (*domains.User, error) {
	user, err := serv.repo.SelectUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// computeGroupFlags вычисляет флаги редактирования/удаления для группы.
// Креатор может всё; админ — ничего (только свои группы, но теперь он их не создаёт).
func (serv *AdminService) computeGroupFlags(callerRole, callerID string, group *domains.Group) {
	if callerRole == "creator" {
		group.CanEdit = true
		group.CanDelete = true
	} else {
		group.CanEdit = (group.OwnerID == callerID) // если он админ, в группе, можно менять информацию о группе
		group.CanDelete = false
	}
}

// ======================================================
//          СОЗДАНИЕ ГРУППЫ (ТОЛЬКО ДЛЯ CREATOR)
// ======================================================

// CreateGroupService создаёт новую группу.
// Параметры:
//   - callerID: идентификатор создателя (всегда creator)
//   - name, description: свойства группы
//   - ownerID: указатель на ID пользователя-админа, которого креатор назначает владельцем группы (может быть nil)
//
// Возвращает: созданную группу с флагами и ошибку.
func (serv *AdminService) CreateGroupService(ctx context.Context, callerID, name, description string, ownerID *string) (*domains.Group, error) {
	tx, err := serv.repo.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback(ctx)

	// 1. Получаем информацию о вызывающем (должен быть creator)
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if callerUser.Role != "creator" {
		return nil, error_type.NewForbidden()
	}

	// 2. Если передан ownerID, валидируем его
	if ownerID != nil {
		ownerUser, err := serv.getUserByID(ctx, *ownerID)
		if err != nil {
			return nil, err
		}
		if ownerUser.Role != "admin" {
			return nil, error_type.NewBadRequest("владелец группы должен обладать ролью admin")
		}
	}

	// 4. Создаём группу в транзакции
	groupID, err := serv.repo.CreateGroupTx(ctx, tx, name, description, callerUser.ID, ownerID)
	if err != nil {
		return nil, err
	}

	// 5. Добавляем в участники креатора и, если назначен, владельца
	if err := serv.repo.InsertUserIntoGroupTx(ctx, tx, groupID, callerUser.ID); err != nil {
		return nil, err
	}
	if ownerID != nil {
		if err := serv.repo.InsertUserIntoGroupTx(ctx, tx, groupID, *ownerID); err != nil {
			return nil, err
		}
	}

	// 6. Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("commit tx: %w", err))
	}

	// 7. если пользователь не передал, то обращаем nil в пустую строку
	// для избегания ошибок при разыменовании в модели
	ownerStr := ""
	if ownerID != nil {
		ownerStr = *ownerID
	}

	// 8. Формируем ответ
	group := &domains.Group{
		GroupID:     groupID,
		Name:        name,
		Description: description,
		CreatedBy:   callerUser.ID,
		OwnerID:     ownerStr,
		CreatedAt:   time.Now(),
	}
	// Креатор всегда может редактировать/удалить созданную им группу
	group.CanEdit = true
	group.CanDelete = true

	return group, nil
}

// ======================================================
//       ПОЛУЧЕНИЕ УЧАСТНИКОВ ГРУППЫ + ИНФО О ГРУППЕ
// ======================================================

// GetMembersGroupService возвращает список участников группы и информацию о группе с флагами.
func (serv *AdminService) GetMembersGroupService(ctx context.Context, callerID, groupID string) (*[]domains.User, *domains.Group, error) {
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return nil, nil, err
	}

	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return nil, nil, error_type.NewForbidden()
	}

	var users *[]domains.User

	// Выборка участников в зависимости от роли
	if callerUser.Role == "creator" {
		users, err = serv.repo.SelectUsersWithAdminFirstIntoGroup(ctx, groupID)
	} else { // admin
		// Админ должен состоять в группе
		consists, err := serv.repo.IsUserIntoGroup(ctx, callerID, groupID)
		if err != nil {
			return nil, nil, err
		}
		if !consists {
			return nil, nil, error_type.NewNotFound("группа не найдена")
		}
		users, err = serv.repo.SelectOnlyUsersIntoGroup(ctx, groupID)
	}
	if err != nil {
		return nil, nil, err
	}

	// Получаем информацию о группе (+ флаги)
	groupInfo, err := serv.repo.SelectGroupInfo(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	serv.computeGroupFlags(callerUser.Role, callerUser.ID, groupInfo)

	return users, groupInfo, nil
}

// ======================================================
//          ПОЛУЧЕНИЕ ВСЕХ ГРУПП (С ФЛАГАМИ)
// ======================================================

// GetGroupsService возвращает список групп, видимых пользователю, с флагами.
func (serv *AdminService) GetGroupsService(ctx context.Context, callerID string) (*[]domains.Group, error) {
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return nil, error_type.NewForbidden()
	}

	var groups *[]domains.Group

	if callerUser.Role == "creator" {
		groups, err = serv.repo.SelectGroupsForCreator(ctx)
	} else { // admin
		groups, err = serv.repo.SelectGroupsForAdmin(ctx, callerID)
	}
	if err != nil {
		return nil, err
	}

	// Заполняем флаги
	for i := range *groups {
		serv.computeGroupFlags(callerUser.Role, callerID, &(*groups)[i])
	}

	return groups, nil
}

// ======================================================
//                   ИЗМЕНЕНИЕ ГРУППЫ
// ======================================================

// EditGroupService изменяет свойства группы (название, описание, владельца).
// owner_id в editInfo может присутствовать только для креатора.
// EditGroupService изменяет свойства группы (название, описание, владельца).
// owner_id в editInfo может присутствовать только для креатора.
func (serv *AdminService) EditGroupService(ctx context.Context, callerID, groupID string, editInfo map[string]string) (*domains.Group, error) {
	// 0. Открываем транзакцию
	tx, err := serv.repo.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback(ctx)

	// 1. Кто вызывает
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	// 2. Существование группы
	groupInfo, err := serv.repo.SelectGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// 3. Проверка прав
	switch callerUser.Role {
	case "creator":
		// creator может всё
	case "admin":
		// админ может редактировать только свои группы
		if groupInfo.OwnerID != callerID {
			return nil, error_type.NewForbidden()
		}
		// и только название / описание; попытка сменить владельца запрещена
		if _, exists := editInfo["owner_id"]; exists {
			return nil, error_type.NewForbidden()
		}
	default:
		return nil, error_type.NewForbidden()
	}

	// 4. Если передано поле owner_id (только creator)
	if _, ok := editInfo["owner_id"]; ok {
		newOwnerID := editInfo["owner_id"]

		switch newOwnerID {
		case groupInfo.OwnerID: // нет изменений
			delete(editInfo, "owner_id") 
		case "": // если передали пустой, снимаем владельца и ставим NULL
			// === Снятие владельца ===
			// Удаляем текущего владельца из участников (если он был)
			if groupInfo.OwnerID != "" {
				if err := serv.repo.DeleteUserFromGroupTx(ctx, tx, groupID, groupInfo.OwnerID); err != nil {
					return nil, err
				}
			}
			// Устанавливаем owner_id в NULL – можно оставить ключ пустым,
			// тогда editInfo["owner_id"] = "" передастся в EditGroupTx,
			// и он обновит поле в NULL.
		default:
			// === Смена владельца на нового админа ===
			// проверяем, что передан корректный UUID
			if _, err := uuid.Parse(newOwnerID); err != nil {
				return nil, error_type.NewBadRequest("некорректный UUID владельца")
			}

			// Проверяем, что новый владелец существует и имеет роль admin
			ownerUser, err := serv.getUserByID(ctx, newOwnerID)
			if err != nil {
				return nil, err
			}
			if ownerUser.Role != "admin" {
				return nil, error_type.NewBadRequest("владелец должен иметь роль admin")
			}

			// Добавляем нового владельца в группу, если его ещё нет
			alreadyMember, err := serv.repo.IsUserIntoGroup(ctx, newOwnerID, groupID)
			if err != nil {
				return nil, err
			}
			if !alreadyMember {
				if err := serv.repo.InsertUserIntoGroupTx(ctx, tx, groupID, newOwnerID); err != nil {
					return nil, err
				}
			}

			// ===== ВАЖНО: удаляем предыдущего владельца из участников =====
			if groupInfo.OwnerID != "" {
				if err := serv.repo.DeleteUserFromGroupTx(ctx, tx, groupID, groupInfo.OwnerID); err != nil {
					return nil, err
				}
			}
		}
	}

	// 5. Редактирование, если owner_id будет "", поставит NULL
	updatedGroup, err := serv.repo.EditGroupTx(ctx, tx, groupID, editInfo)
	if err != nil {
		return nil, err
	}

	// 6. Флаги для UI
	serv.computeGroupFlags(callerUser.Role, callerID, updatedGroup)

	// 7. Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, error_type.NewInternal(fmt.Errorf("commit tx: %w", err))
	}

	return updatedGroup, nil
}

// ======================================================
//         ДОБАВЛЕНИЕ УЧАСТНИКА В ГРУППУ
// ======================================================

// AddUserGroupService добавляет пользователя targetID в группу groupID.
func (serv *AdminService) AddUserGroupService(ctx context.Context, callerID, targetID, groupID string) error {
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return err
	}
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return error_type.NewForbidden()
	}

	targetUser, err := serv.getUserByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Проверяем существование группы
	groupInfo, err := serv.repo.SelectGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}
	_ = groupInfo

	// Если добавляемый пользователь — admin, проверяем, что в группе ещё нет админа
	if targetUser.Role == "admin" {
		hasAdmin, err := serv.repo.GroupHasAdmin(ctx, groupID)
		if err != nil {
			return err
		}
		if hasAdmin {
			return error_type.NewConflict("в группе уже есть администратор")
		}
	}

	if callerUser.Role == "creator" {
		// Креатор может добавить любого
		// (дополнительно можно запретить добавлять другого creator, но creator один)
	} else { // admin
		// Админ может добавлять только пользователей с ролью "user"
		if targetUser.Role != "user" {
			return error_type.NewForbidden()
		}
		// Админ должен состоять в группе
		consists, err := serv.repo.IsUserIntoGroup(ctx, callerID, groupID)
		if err != nil {
			return err
		}
		if !consists {
			return error_type.NewNotFound("группа не найдена")
		}
	}

	// Проверяем, не состоит ли уже в группе
	already, err := serv.repo.IsUserIntoGroup(ctx, targetID, groupID)
	if err != nil {
		return err
	}
	if already {
		return error_type.NewConflict("пользователь уже состоит в группе")
	}

	// Добавляем
	if err := serv.repo.InsertUserIntoGroup(ctx, groupID, targetID); err != nil {
		return err
	}
	return nil
}

// ======================================================
//         УДАЛЕНИЕ УЧАСТНИКА ИЗ ГРУППЫ
// ======================================================

// DeleteUserGroupService удаляет пользователя targetID из группы groupID.
func (serv *AdminService) DeleteUserGroupService(ctx context.Context, callerID, targetID, groupID string) error {
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return err
	}
	if callerUser.Role != "creator" && callerUser.Role != "admin" {
		return error_type.NewForbidden()
	}

	targetUser, err := serv.getUserByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Нельзя удалить самого себя
	if callerID == targetID {
		return error_type.NewForbidden()
	}

	// Проверяем существование группы
	groupInfo, err := serv.repo.SelectGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}

	// Проверяем, что удаляемый пользователь не является владельцем группы
	if groupInfo.OwnerID != "" && groupInfo.OwnerID == targetID {
		return error_type.NewConflict("нельзя удалить владельца группы. Сначала назначьте нового владельца")
	}

	if callerUser.Role == "admin" {
		// Админ может удалять только user
		if targetUser.Role != "user" {
			return error_type.NewForbidden()
		}
		// Админ должен состоять в группе
		consists, err := serv.repo.IsUserIntoGroup(ctx, callerID, groupID)
		if err != nil {
			return err
		}
		if !consists {
			return error_type.NewNotFound("группа не найдена")
		}
	}

	// Удаляем (репозиторий проверит, что пользователь был в группе)
	if err := serv.repo.DeleteUserFromGroup(ctx, groupID, targetID); err != nil {
		return err
	}
	return nil
}

// ======================================================
//                УДАЛЕНИЕ ГРУППЫ
// ======================================================

// DeleteGroupService удаляет группу. Только creator может удалить любую группу.
func (serv *AdminService) DeleteGroupService(ctx context.Context, callerID, groupID string) error {
	callerUser, err := serv.getUserByID(ctx, callerID)
	if err != nil {
		return err
	}
	if callerUser.Role != "creator" {
		return error_type.NewForbidden()
	}

	// Проверяем существование группы
	_, err = serv.repo.SelectGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}

	// Удаляем группу (каскадное удаление участников, задачи нужно обработать отдельно)
	if err := serv.repo.DeleteGroup(ctx, groupID); err != nil {
		return err
	}
	return nil
}
