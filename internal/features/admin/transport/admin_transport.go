package transport

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/server/authctx"
	"accelerator/internal/features/admin/service"
	"accelerator/internal/features/admin/transport/dto"
	"accelerator/internal/tools"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type AdminTransport struct {
	serv     *service.AdminService
	validate *validator.Validate
}

func NewAdminTransport(serv *service.AdminService, validate *validator.Validate) *AdminTransport {
	return &AdminTransport{
		serv:     serv,
		validate: validate,
	}
}

// ====================================================== СОЗДАНИЕ КРЕАТОРА ==============================================

type AddCreatorRequestDTO struct {
	Login    string `json:"login" validate:"required,email,max=255"`
	FullName string `json:"full_name" validate:"required,fio,max=100"`
	Position string `json:"position" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

type AddCreatorResponseDTO struct {
	UserID   string `json:"user_id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Position string `json:"position"`
	Role     string `json:"role"`
}

func (trans *AdminTransport) AddCreatorHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	newRequest := AddCreatorRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("не удалось распарсить json"))
		return
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	userInfo, err := trans.serv.AddCreatorService(
		ctx,
		newRequest.Login, newRequest.Password, newRequest.FullName, newRequest.Position,
	)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// маппим в дто и отправляем пользователю
	newResponse := AddCreatorResponseDTO{
		UserID:   userInfo.ID,
		Login:    userInfo.Login,
		FullName: userInfo.FullName,
		Position: userInfo.Position,
		Role:     userInfo.Role,
	}

	// записываем данные в ответ
	tools.WriteJSON(w, http.StatusCreated, newResponse)

}

// ================================================= МЕТОДЫ ВЗАИМОДЕЙСТВИЯ С ПОЛЬЗОВАТЕЛЯМИ ==============================================

// ====================== РЕГИСТРАЦИЯ НОВОГО РОЛЬЗОВАТЕЛЯ =======================

type RegisterUserRequestDTO struct {
	Login    string `json:"login" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,fio"`
	Position string `json:"position" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

type RegisterUserResponseDTO struct {
	UserID   string `json:"user_id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Position string `json:"position"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

func (trans *AdminTransport) RegisterNewUserHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		// Не должно случиться, если middleware правильно настроен, но на всякий случай
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
	}

	newRequest := RegisterUserRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("не удалось распарсить json"))
		return
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// вызываем сервис, он генерирует пароль и возвращает данные о пользователе
	// проверить, чтобы админов мог только креатор добавлять, от админов только юзеров принимать
	userInfo, userPassword, err := trans.serv.RegisterNewUserService(
		ctx, callerID,
		newRequest.Login, newRequest.FullName, newRequest.Position, newRequest.Role,
	)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// маппим данные из домена в dto response
	newResponse := RegisterUserResponseDTO{
		UserID:   userInfo.ID,
		Login:    userInfo.Login,
		FullName: userInfo.FullName,
		Position: userInfo.Position,
		Role:     userInfo.Role,
		Password: userPassword,
	}

	// записываем данные в ответ
	tools.WriteJSON(w, http.StatusCreated, newResponse)

}

// ====================== ПОЛУЧЕНИЕ ПОЛЬЗОВАТЕЛЕЙ =======================

type GetUsersRequestDTO struct {
	Page  string `validate:"required,number"`
	Limit string `validate:"required,number"`
}

type GetUsersResponseDTO struct {
	Users      []dto.UserResponseDTO     `json:"users"`
	Pagination dto.PaginationResponseDTO `json:"pagination"`
}

func (trans *AdminTransport) GetUsersHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
	}

	newRequest := GetUsersRequestDTO{
		Page:  r.URL.Query().Get("page"),
		Limit: r.URL.Query().Get("limit"),
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// переводим данные в integer, уже валидировали, так что ошибку не получаем, там точно int
	pageInt, _ := strconv.Atoi(newRequest.Page)
	limitInt, _ := strconv.Atoi(newRequest.Limit)

	// вызываем сервис, он возвращает список пользователей и их общее количество
	// присылает пользователей исходя от роли, если креатор, то все, если админ, то только юзеры
	users, usersCount, err := trans.serv.GetUsersService(
		ctx, callerID,
		pageInt, limitInt,
	)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// закидываем данные в dto response и отправляем
	usersResponse := make([]dto.UserResponseDTO, 0)

	for i := 0; i < len(*users); i++ {
		userResponse := dto.UserResponseDTO{
			UserID:    (*users)[i].ID,
			Login:     (*users)[i].Login,
			FullName:  (*users)[i].FullName,
			Position:  (*users)[i].Position,
			Role:      (*users)[i].Role,
			CreatedAt: (*users)[i].CreatedAt,
		}

		usersResponse = append(usersResponse, userResponse)
	}

	pagination := dto.PaginationResponseDTO{
		Page:  pageInt,
		Limit: limitInt,
		Total: usersCount,
	}

	newResponse := GetUsersResponseDTO{
		Users:      usersResponse,
		Pagination: pagination,
	}

	// записываем в ответ
	tools.WriteJSON(w, http.StatusOK, newResponse)
}

// ======================== ИЗМЕНЕНИЕ ПОЛЬЗОВАТЕЛЯ =========================
// можно либо передать только поля, которые изменяешь
// либо все измененные поля

// получаем указатели, чтобы отличить, когда пользователь ввел пустую строку
// от того, что поле просто не было указано в json, тогда будет nil
type EditUserRequestDTO struct {
	Login    *string `json:"login" validate:"omitempty,email"`
	FullName *string `json:"full_name" validate:"omitempty,fio"`
	Position *string `json:"position" validate:"omitempty"`
	Role     *string `json:"role" validate:"omitempty,oneof=admin user"`
}

func (trans *AdminTransport) EditUserHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
	}

	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	newRequest := EditUserRequestDTO{}

	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("не удалось распарсить json"))
		return
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// собираем только данные, которые не равны пустой строке или nil и которые нужно изменить
	// если пустые строки, то кидаем ошибку, что поле не может быть пустым
	updateData := make(map[string]string)

	// Логин пройдет по правилу fio, не нужна проверка на пустую строку
	if newRequest.Login != nil {
		updateData["login"] = *newRequest.Login
	}
	// ФИО валидатор отсечет ""
	if newRequest.FullName != nil {
		updateData["full_name"] = *newRequest.FullName
	}
	// Должность, а здесь проверяем
	if newRequest.Position != nil {
		if *newRequest.Position == "" {
			tools.WriteError(w, error_type.NewBadRequest("должность не может быть пустой строкой"))
			return
		}
		updateData["position"] = *newRequest.Position
	}
	// Роль валидатор отсечет ""
	if newRequest.Role != nil {
		updateData["role"] = *newRequest.Role
	}

	// Проверка, что есть что обновлять
	if len(updateData) == 0 {
		tools.WriteError(w, error_type.NewBadRequest("нет ни одного переданного аргумента для изменения"))
		return
	}

	// вызываем сервис, он возвращает доменную структуру юзера
	// проверить, что только креатор может повышать пользователей до админов
	userEditInfo, err := trans.serv.EditUserService(ctx, callerID, newRequestUserID.UserID, updateData)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// маппим данные из домена в dto response
	newResponse := dto.UserResponseDTO{
		UserID:    userEditInfo.ID,
		Login:     userEditInfo.Login,
		FullName:  userEditInfo.FullName,
		Position:  userEditInfo.Position,
		Role:      userEditInfo.Role,
		CreatedAt: userEditInfo.CreatedAt,
	}

	// записываем данные в ответ
	tools.WriteJSON(w, http.StatusOK, newResponse)

}

// ====================== СБРОСИТЬ ПАРОЛЬ ДЛЯ ПОЛЬЗОВАТЕЛЯ =======================

type ResetPasswordResponseDTO struct {
	Password string `json:"password"`
}

func (trans *AdminTransport) ResetPasswordHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
	}

	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// передаем id пользователя, сервис генерирует новый пароль, сохраняет его и передает сюда
	// проверить, что только креатор может сбрасывать админов, а админ только юзеров, по идее сюда нельзя будет попасть с ui, но узнав id креатора или админа, админ сможет сбросить им пароль
	newPassword, err := trans.serv.ResetPasswordService(ctx, callerID, newRequestUserID.UserID)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	// записываем его в ResetPasswordResponseDTO и передаем на клиент
	newResponse := ResetPasswordResponseDTO{
		Password: newPassword,
	}

	tools.WriteJSON(w, http.StatusOK, newResponse)
}

// ====================== УДАЛИТЬ ПОЛЬЗОВАТЕЛЯ =======================

type DeleteUserRequestDTO struct {
	UserID string `validate:"required,uuid4"`
}

func (trans *AdminTransport) DeleteUserHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	callerID, ok := authctx.GetUserID(ctx)
	if !ok {
		tools.WriteError(w, error_type.NewUnauthorized("missing authentication context"))
	}

	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID пользователя"))
		return
	}

	// передаем сюда id, удаляем пользователя
	// проверить, что только креатор может удалять админов, а админ только юзеров, по идее сюда нельзя будет попасть с ui, но узнав id креатора или админа, админ сможет удалить им аккаунт
	if err := trans.serv.DeleteUserService(ctx, callerID, newRequestUserID.UserID); err != nil {
		tools.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ====================================================== МЕТОДЫ ВЗАИМОДЕЙСТВИЯ С ГРУППАМИ ==============================================

// ---------- Вспомогательные методы ----------

// getCallerID извлекает callerID из контекста (установлен middleware)
func getCallerID(r *http.Request) string {
	id, _ := r.Context().Value("userID").(string)
	return id
}

// ====================== ГРУППЫ ======================

// CreateGroupHandle создаёт новую группу
func (trans *AdminTransport) CreateGroupHandle(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGroupRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("не удалось распарсить json"))
		return
	}
	if err := trans.validate.Struct(req); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("ошибка во входных данных"))
		return
	}

	callerID := getCallerID(r)

	group, err := trans.serv.CreateGroupService(r.Context(), callerID, req.Name, req.Description, req.OwnerID)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	resp := dto.GroupResponseDTO{
		GroupID:     group.GroupID,
		Name:        group.Name,
		Description: group.Description,
		OwnerID:     group.OwnerID,
		CreatedAt:   group.CreatedAt,
		CanEdit:     group.CanEdit,
		CanDelete:   group.CanDelete,
	}
	tools.WriteJSON(w, http.StatusCreated, resp)
}

// GetMembersGroupHandle возвращает информацию о группе и список участников
func (trans *AdminTransport) GetMembersGroupHandle(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	if err := trans.validate.Struct(dto.GroupIDRequestDTO{GroupID: groupID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID группы"))
		return
	}

	callerID := getCallerID(r)
	members, group, err := trans.serv.GetMembersGroupService(r.Context(), callerID, groupID)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	memberDTOs := make([]dto.MembersResponseDTO, 0, len(*members))
	for _, m := range *members {
		memberDTOs = append(memberDTOs, dto.MembersResponseDTO{
			UserID:   m.ID,
			Login:    m.Login,
			FullName: m.FullName,
			Position: m.Position,
			Role:     m.Role,
			AddedAt:  m.AddedAt,
		})
	}

	resp := dto.GetMembersGroupResponseDTO{
		GroupID:     group.GroupID,
		Name:        group.Name,
		Description: group.Description,
		Members:     memberDTOs,
		OwnerID:     group.OwnerID,
		CreatedAt:   group.CreatedAt,
		CanEdit:     group.CanEdit,
		CanDelete:   group.CanDelete,
	}
	tools.WriteJSON(w, http.StatusOK, resp)
}

// GetGroupsHandle возвращает список групп, доступных пользователю
func (trans *AdminTransport) GetGroupsHandle(w http.ResponseWriter, r *http.Request) {
	callerID := getCallerID(r)
	groups, err := trans.serv.GetGroupsService(r.Context(), callerID)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	groupDTOs := make([]dto.GroupResponseDTO, 0, len(*groups))
	for _, g := range *groups {
		groupDTOs = append(groupDTOs, dto.GroupResponseDTO{
			GroupID:     g.GroupID,
			Name:        g.Name,
			Description: g.Description,
			MemberCount: g.MemberCount,
			OwnerID:     g.OwnerID,
			CreatedAt:   g.CreatedAt,
			CanEdit:     g.CanEdit,
			CanDelete:   g.CanDelete,
		})
	}
	resp := dto.GetGroupsResponseDTO{Groups: groupDTOs}
	tools.WriteJSON(w, http.StatusOK, resp)
}

// EditGroupHandle изменяет название, описание или владельца группы
func (trans *AdminTransport) EditGroupHandle(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	if err := trans.validate.Struct(dto.GroupIDRequestDTO{GroupID: groupID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID группы"))
		return
	}

	var req dto.EditGroupRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("не удалось распарсить json"))
		return
	}
	if err := trans.validate.Struct(req); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("ошибка во входных данных"))
		return
	}

	updateData := make(map[string]string)
	if req.Name != nil {
		if *req.Name == "" {
			tools.WriteError(w, error_type.NewBadRequest("название группы не может быть пустым"))
			return
		}
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		if *req.Description == "" {
			tools.WriteError(w, error_type.NewBadRequest("описание группы не может быть пустым"))
			return
		}
		updateData["description"] = *req.Description
	}
	if req.OwnerID != nil {
		updateData["owner_id"] = *req.OwnerID
	}
	if len(updateData) == 0 {
		tools.WriteError(w, error_type.NewBadRequest("не передано ни одного поля для изменения"))
		return
	}

	callerID := getCallerID(r)
	updatedGroup, err := trans.serv.EditGroupService(r.Context(), callerID, groupID, updateData)
	if err != nil {
		tools.WriteError(w, err)
		return
	}

	resp := dto.GroupResponseDTO{
		GroupID:     updatedGroup.GroupID,
		Name:        updatedGroup.Name,
		Description: updatedGroup.Description,
		MemberCount: updatedGroup.MemberCount,
		OwnerID:     updatedGroup.OwnerID,
		CreatedAt:   updatedGroup.CreatedAt,
		CanEdit:     updatedGroup.CanEdit,
		CanDelete:   updatedGroup.CanDelete,
	}
	tools.WriteJSON(w, http.StatusOK, resp)
}

// AddUserGroupHandle добавляет пользователя в группу
func (trans *AdminTransport) AddUserGroupHandle(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	userID := chi.URLParam(r, "userID")
	if err := trans.validate.Struct(dto.GroupIDRequestDTO{GroupID: groupID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID группы"))
		return
	}
	if err := trans.validate.Struct(dto.UserIDRequestDTO{UserID: userID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID пользователя"))
		return
	}

	callerID := getCallerID(r)
	if err := trans.serv.AddUserGroupService(r.Context(), callerID, userID, groupID); err != nil {
		tools.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// DeleteUserGroupHandle удаляет пользователя из группы
func (trans *AdminTransport) DeleteUserGroupHandle(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	userID := chi.URLParam(r, "userID")
	if err := trans.validate.Struct(dto.GroupIDRequestDTO{GroupID: groupID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID группы"))
		return
	}
	if err := trans.validate.Struct(dto.UserIDRequestDTO{UserID: userID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID пользователя"))
		return
	}

	callerID := getCallerID(r)
	if err := trans.serv.DeleteUserGroupService(r.Context(), callerID, userID, groupID); err != nil {
		tools.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteGroupHandle удаляет группу
func (trans *AdminTransport) DeleteGroupHandle(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	if err := trans.validate.Struct(dto.GroupIDRequestDTO{GroupID: groupID}); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("некорректный ID группы"))
		return
	}

	callerID := getCallerID(r)
	if err := trans.serv.DeleteGroupService(r.Context(), callerID, groupID); err != nil {
		tools.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
