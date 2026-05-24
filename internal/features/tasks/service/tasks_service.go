package service

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/storage"
	"accelerator/internal/domains"
	"accelerator/internal/features/tasks/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type TasksService struct {
	repo  *repository.TasksRepo
	minio *storage.MinIOClient
	cfg   *config.Config
}

func NewTasksService(repo *repository.TasksRepo, minio *storage.MinIOClient, cfg *config.Config) *TasksService {
	return &TasksService{
		repo:  repo,
		minio: minio,
		cfg:   cfg,
	}
}

// ==================================== ЗАГРУЗКА ЗАДАЧИ ============================================

func (serv *TasksService) UploadTaskService(
	ctx context.Context,
	userID, groupID,
	taskName, taskDescription, meetingDate, patternID,
	fileName, filePath, statusTask string,
) (*domains.Task, error) {
	// проверяем, состоит ли пользователь в группе, в которую он хочет загрузить задачу
	consist, err := serv.repo.IsUserIntoGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	if !consist {
		return nil, error_type.NewNotFound("Группа не найдена")
	}

	task, err := serv.repo.CreateTask(
		ctx, userID, groupID,
		taskName, taskDescription, meetingDate, patternID,
		filePath, fileName, statusTask,
	)
	if err != nil {
		return nil, err
	}

	// сразу после создания ни удалять ни изменять нельзя
	task.ChangeFlag = false

	return task, nil

}

// ================================== ПОЛУЧЕНИЕ СТАТУСА ЗАДАЧИ =====================================

// возвращает:
// - статус,
// - флаг процесс или ожидание true - процесс
// - количество человек в очереди с высшим приоритетом (при ожидании процесса),
// - примерное время ожидания (для начатого процесса) в минутах
func (serv *TasksService) GetTaskStatusService(ctx context.Context, callerID, taskID string) (*domains.TaskCheck, error) {
	// проверяем, имеет ли пользователь доступ к этой задаче
	exists, err := serv.repo.CheckUserInTaskGroup(ctx, callerID, taskID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, error_type.NewNotFound("Задача не найдена")
	}

	// получаем статус задачи, а заодно и проверяем ее существование
	status, err := serv.repo.SelectTaskStatus(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// получаем объект конвеера и флаг, процесс или ожидание
	currentStatus := domains.TaskStatus(status)
	isProcessing := currentStatus.IsProcessing()

	// динамические переменные, по умолчанию 0
	var queueCountBefore int
	var approximateLeadTimeProcess int

	// получаем количество в очереди до нас если статус ожидания
	if !isProcessing {
		queueCountBefore, err = serv.repo.GetQueuePosition(ctx, status, taskID)
		if err != nil {
			return nil, err
		}
	}

	// если идет процесс, получаем его примерную длительность !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	if isProcessing {

	}

	statusCheck := domains.TaskCheck{
		Status:                     status,
		IsProcess:                  isProcessing,
		InTheQueueBefore:           queueCountBefore,
		ApproximateLeadTimeProcess: approximateLeadTimeProcess,
	}

	return &statusCheck, nil
}

// ====================================== ПОЛУЧЕНИЕ ЗАДАЧИ =========================================

func (serv *TasksService) GetTaskService(ctx context.Context, callerID, taskID string) (*domains.Task, error) {
	// получаем данные о пользователе
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	if userInfo.Role == "admin" || userInfo.Role == "user" {
		// проверяем, имеет ли пользователь доступ к этой задаче если он не креатор
		exists, err := serv.repo.CheckUserInTaskGroup(ctx, callerID, taskID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, error_type.NewNotFound("Задача не найдена")
		}
	}

	// проверка существования задачи и получение задачи
	taskInfo, err := serv.repo.SelectTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// ставим флаги
	if userInfo.Role == "user" && taskInfo.Status == string(domains.StatusDone) {
		if callerID == taskInfo.UserID {
			taskInfo.ChangeFlag = true
		} else {
			taskInfo.ChangeFlag = false
		}
	}

	// админ может менять в своих группах (проверка была выше), а креатор все
	// если мы дошли до этого момента, значит мы админ в группе, либа креатор
	// можем изменять и удалять все, которые завершили обработку
	if userInfo.Role == "admin" || userInfo.Role == "creator" {
		if taskInfo.Status == string(domains.StatusDone) {
			taskInfo.ChangeFlag = true
		} else {
			taskInfo.ChangeFlag = false
		}
	}

	return taskInfo, nil
}

// ===================================== ПОЛУЧЕНИЕ ВСЕХ ЗАДАЧ В ГРУППЕ ==========================================

// возвращает список всех задач и их общее количество
func (serv *TasksService) GetAllTaskInGroupService(ctx context.Context, callerID, groupID string, page, limit int) (*[]domains.Task, int64, error) {
	// получаем данные о пользователе
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, 0, err
	}

	// проверка существования группы
	valid, err := serv.repo.CheckGroup(ctx, groupID)
	if err != nil {
		return nil, 0, err
	}
	if !valid {
		return nil, 0, error_type.NewNotFound("Группа не найдена")
	}

	if userInfo.Role == "admin" || userInfo.Role == "user" {
		// проверяем, состоит ли пользователь в переданной группе
		exists, err := serv.repo.IsUserIntoGroup(ctx, callerID, groupID)
		if err != nil {
			return nil, 0, err
		}
		if !exists {
			return nil, 0, error_type.NewNotFound("Группа не найдена")
		}
	}

	tasksInfo, totalTasks, err := serv.repo.GetTasksByGroupID(ctx, groupID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// ставим флаги

	// юзер только изменяет задачи, которые он создал
	if userInfo.Role == "user" {
		for index := range *tasksInfo {
			// также чтобы изменять, задача должна быть выполнена
			if (*tasksInfo)[index].UserID == callerID && (*tasksInfo)[index].Status == string(domains.StatusDone) {
				(*tasksInfo)[index].ChangeFlag = true
			} else {
				(*tasksInfo)[index].ChangeFlag = false
			}
		}
	}

	// креатор или админ в своей группе могут менять все, которые завершенные
	if userInfo.Role == "admin" || userInfo.Role == "creator" {
		for index := range *tasksInfo {
			if (*tasksInfo)[index].Status == string(domains.StatusDone) {
				(*tasksInfo)[index].ChangeFlag = true
			} else {
				(*tasksInfo)[index].ChangeFlag = false
			}

		}
	}

	return tasksInfo, totalTasks, nil
}

// ============================ ИЗМЕНЕНИЕ СТАТУСА ЗАДАЧ (доп функции при upload ВНУТРЕННИЕ) =====================================

func (serv *TasksService) UpdateTaskStatusService(ctx context.Context, callerID, taskID, statusTask string) error {
	if err := serv.repo.UpdateTaskStatus(ctx, taskID, statusTask); err != nil {
		return err
	}

	return nil
}

func (serv *TasksService) UpdateTaskSuccessUploadService(ctx context.Context, callerID, taskID, filePath string, duration int, newStatus, inputKey, outputKey string) error {
	tx, err := serv.repo.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return error_type.NewInternal(fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback(ctx)

	if err := serv.repo.UpdateTaskSuccessUploadTx(ctx, tx, taskID, filePath, duration, newStatus); err != nil {
		return err
	}

	if err := serv.repo.UpdateObjectKeysTx(ctx, tx, taskID, inputKey, outputKey); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return error_type.NewInternal(fmt.Errorf("commit tx: %w", err))
	}

	return nil
}

// ===================================== ИЗМЕНЕНИЕ ЗАДАЧИ ===========================================
func (serv *TasksService) EditTaskService(ctx context.Context, callerID, taskID string, editInfo map[string]string) (*domains.Task, error) {
	// получаем данные о пользователе
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	// получаем данные о задаче
	taskInfo, err := serv.repo.SelectTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// если задача уже обрабатывается воркерами, то запрещаем удаление, чтобы не было такого
	// что воркеры работают над удаленной задачей
	currentStatus := domains.TaskStatus(taskInfo.Status)
	if currentStatus != domains.StatusDone {
		return nil, error_type.NewBadRequest("Нельзя изменять задачу, которая находится в обработке")
	}

	if userInfo.Role == "admin" || userInfo.Role == "user" {
		// проверяем, имеет ли пользователь доступ к этой задаче
		// для админа этой проверки достаточно
		exists, err := serv.repo.CheckUserInTaskGroup(ctx, callerID, taskID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, error_type.NewNotFound("Задача не найдена")
		}

		// пользователь может изменять только созданую им задачу
		if userInfo.Role == "user" {
			if taskInfo.UserID != userInfo.ID {
				return nil, error_type.NewNotFound("Задача не найдена")
			}
		}
	}

	// если мы здесь, значит мы либо креатор, либо админ или пользователь с соблюденными условиями, можем изменять
	editTask, err := serv.repo.EditTask(ctx, taskID, editInfo)
	if err != nil {
		return nil, err
	}

	// если задача была изменена, значит мы можем ее менять
	editTask.ChangeFlag = true

	return editTask, nil
}

// ===================================== УДАЛЕНИЕ ЗАДАЧИ ===========================================
func (serv *TasksService) DeleteTaskService(ctx context.Context, callerID, taskID string) error {
	// получаем данные о пользователе
	userInfo, err := serv.repo.SelectUserByID(ctx, callerID)
	if err != nil {
		return err
	}

	// получаем данные о задаче
	taskInfo, err := serv.repo.SelectTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// если задача уже обрабатывается воркерами, то запрещаем удаление, чтобы не было такого
	// что воркеры работают над удаленной задачей
	currentStatus := domains.TaskStatus(taskInfo.Status)
	if currentStatus != domains.StatusDone {
		return error_type.NewBadRequest("Нельзя удалить задачу, которая находится в обработке")
	}

	if userInfo.Role == "admin" || userInfo.Role == "user" {
		// проверяем, имеет ли пользователь доступ к этой задаче
		// для админа этой проверки достаточно
		exists, err := serv.repo.CheckUserInTaskGroup(ctx, callerID, taskID)
		if err != nil {
			return err
		}
		if !exists {
			return error_type.NewNotFound("Задача не найдена")
		}

		// пользователь может удалять только созданую им задачу
		if userInfo.Role == "user" {
			if taskInfo.UserID != userInfo.ID {
				return error_type.NewNotFound("Задача не найдена")
			}
		}
	}

	// если мы здесь, значит мы либо креатор, либо админ или пользователь с соблюденными условиями, можем удалять
	if err := serv.repo.DeleteTask(ctx, taskID); err != nil {
		return err
	}

	return nil
}
