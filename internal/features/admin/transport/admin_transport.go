package transport

import (
	"accelerator/internal/core/error_type"
	"accelerator/internal/features/admin/service"
	"accelerator/internal/features/admin/transport/dto"
	"accelerator/internal/tools"
	"encoding/json"
	"net/http"

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
	callerID := "" // !!!!!!!!!!!!!!!!!!!!!!!!!!!

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
	callerID := "" // !!!!!!!!!!!!!!!!!!!!!!!!!!!

	newRequest := GetUsersRequestDTO{
		Page:  r.URL.Query().Get("page"),
		Limit: r.URL.Query().Get("limit"),
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// вызываем сервис, он возвращает список пользователей и их общее количество
	// присылает пользователей исходя от роли, если креатор, то все, если админ, то только юзеры
	users, usersCount, err := trans.serv.GetUsersService(
		ctx, callerID,
		newRequest.Page, newRequest.Limit,
	)
	if err != nil {
		tools.WriteError(w, err)
	}

	// закидываем данные в dto response и отправляем
	var usersResponse []dto.UserResponseDTO

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
		Page:  newRequest.Page,
		Limit: newRequest.Limit,
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

	// записываем в UserResponseDTO
}

// ====================== СБРОСИТЬ ПАРОЛЬ ДЛЯ ПОЛЬЗОВАТЕЛЯ =======================

type ResetPasswordResponseDTO struct {
	Password string `json:"password"`
}

func (trans *AdminTransport) ResetPasswordHandle(w http.ResponseWriter, r *http.Request) {
	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// передаем id пользователя, сервис генерирует новый пароль, сохраняет его и передает сюда
	// проверить, что только креатор может сбрасывать админов, а админ только юзеров, по идее сюда нельзя будет попасть с ui, но узнав id креатора или админа, админ сможет сбросить им пароль

	// записываем его в ResetPasswordResponseDTO и передаем на клиент
}

// ====================== УДАЛИТЬ ПОЛЬЗОВАТЕЛЯ =======================

type DeleteUserRequestDTO struct {
	UserID string `validate:"required,uuid4"`
}

func (trans *AdminTransport) DeleteUserHandle(w http.ResponseWriter, r *http.Request) {
	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID пользователя"))
		return
	}

	// передаем сюда id, удаляем пользователя
	// проверить, что только креатор может удалять админов, а админ только юзеров, по идее сюда нельзя будет попасть с ui, но узнав id креатора или админа, админ сможет удалить им аккаунт

	w.WriteHeader(http.StatusNoContent)
}

// ====================================================== МЕТОДЫ ВЗАИМОДЕЙСТВИЯ С ГРУППАМИ ==============================================

// ========================= СОЗДАТЬ ГРУППУ ==========================
type CreateGroupRequestDTO struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (trans *AdminTransport) CreateGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequest := CreateGroupRequestDTO{}

	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Не удалось распарсить json"))
		return
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// вызываем сервис, получаем от него модель группы

	// маппим модель в GroupResponseDTO и отсылаем на клиент
}

// ========================= ПОЛУЧИТЬ УЧАСТНИКОВ ГРУППЫ  ==========================

type GetMembersGroupRequestDTO struct {
	GroupID     string                   `json:"group_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Members     []dto.MembersResponseDTO `json:"members"`
	CreatedBy   string                   `json:"created_by"`
	CreatedAt   string                   `json:"created_at"`
}

func (trans *AdminTransport) GetMembersGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequestGroupID := dto.GroupIDRequestDTO{
		GroupID: chi.URLParam(r, "groupID"),
	}

	if err := trans.validate.Struct(newRequestGroupID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID группы"))
		return
	}

	// в сервисе получаем информацию о группе из groups, далее user_id и added_at из таблицы members,
	// а дальше получаем по user_id информацию о пользователях из users
	// нужно также отфильтровать исходя из роли, чтобы админ не получал других админов, а креатор получал всех
	// также исключить чтобы админ не видел себя и креатор не видел себя

	// формируем json ответ в GetMembersGroupRequestDTO
}

// ========================= ПОЛУЧИТЬ ВСЕ ГРУППЫ  ==========================
type GetGroupsResponseDTO struct {
	Groups []dto.GroupResponseDTO
}

func (trans *AdminTransport) GetGroupsHandle(w http.ResponseWriter, r *http.Request) {

	// сервис будет возвращать только те группы, в которых состоит пользователь

	// записываем результат в GetGroupsResponseDTO
}

// ======================== ИЗМЕНИТЬ ИНФОРМАЦИЮ О ГРУППЕ  ==========================

type EditGroupRequestDTO struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
}

func (trans *AdminTransport) EditGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequestGroupID := dto.GroupIDRequestDTO{
		GroupID: chi.URLParam(r, "groupID"),
	}

	if err := trans.validate.Struct(newRequestGroupID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID группы"))
		return
	}

	newRequest := EditGroupRequestDTO{}

	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Не удалось распарсить json"))
		return
	}

	if err := trans.validate.Struct(newRequest); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Ошибка во входных данных"))
		return
	}

	// создаем мапу для измененных значений
	updateData := make(map[string]string)

	// если nil, значит значение просто не передавали
	if newRequest.Name != nil {
		// может быть передана пустая строка
		if *newRequest.Description == "" {
			tools.WriteError(w, error_type.NewBadRequest("название группы не может быть пустой строкой"))
			return
		}
		updateData["name"] = *newRequest.Name
	}
	if newRequest.Description != nil {
		if *newRequest.Description == "" {
			tools.WriteError(w, error_type.NewBadRequest("описание группы не может быть пустой строкой"))
			return
		}
		updateData["description"] = *newRequest.Description
	}

	// Проверка, что есть что обновлять
	if len(updateData) == 0 {
		tools.WriteError(w, error_type.NewBadRequest("нет ни одного переданного аргумента для изменения"))
		return
	}

	// сервис получает updateData изменяет данные группы и возвращает их в модели
	// проверка, что пользователь состоит в этой группе на всякий случай

	// записываем результат в GroupResponseDTO и возвращаем ответ
}

// ====================== ДОБАВИТЬ ПОЛЬЗОВАТЕЛЯ В ГРУППУ  ==========================

func (trans *AdminTransport) AddUserGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequestGroupID := dto.GroupIDRequestDTO{
		GroupID: chi.URLParam(r, "groupID"),
	}
	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestGroupID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID группы"))
		return
	}
	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID пользователя"))
		return
	}

	// сервис добавляет юзера в группу
	// проверка, что только креатор может добавлять админов в группы, а админ только юзеров в группы, в которых он состоит

	w.WriteHeader(http.StatusCreated)
}

// ====================== УДАЛИТЬ ПОЛЬЗОВАТЕЛЯ ИЗ ГРУППЫ  ==========================

func (trans *AdminTransport) DeleteUserGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequestGroupID := dto.GroupIDRequestDTO{
		GroupID: chi.URLParam(r, "groupID"),
	}
	newRequestUserID := dto.UserIDRequestDTO{
		UserID: chi.URLParam(r, "userID"),
	}

	if err := trans.validate.Struct(newRequestGroupID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID группы"))
		return
	}
	if err := trans.validate.Struct(newRequestUserID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID пользователя"))
		return
	}

	// сервис удаляет юзера в группу
	// проверка, что только креатор может удалять админов из всех группы, а админ только юзеров из группы, в которых он состоит

	w.WriteHeader(http.StatusNoContent)
}

// ============================ УДАЛИТЬ ГРУППУ  ==============================

func (trans *AdminTransport) DeleteGroupHandle(w http.ResponseWriter, r *http.Request) {
	newRequestGroupID := dto.GroupIDRequestDTO{
		GroupID: chi.URLParam(r, "groupID"),
	}

	if err := trans.validate.Struct(newRequestGroupID); err != nil {
		tools.WriteError(w, error_type.NewBadRequest("Некорректный ID группы"))
		return
	}

	// сервис удаляет группу
	// проверка, что только креатор может удалять все группы, а админ только группы, в которых он состоит

	w.WriteHeader(http.StatusNoContent)
}
