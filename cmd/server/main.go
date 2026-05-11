package main

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/logger"
	"accelerator/internal/core/server"
	adminRepository "accelerator/internal/features/admin/repository"
	authRepository "accelerator/internal/features/auth/repository"
	tasksRepository "accelerator/internal/features/tasks/repository"
	"accelerator/internal/tools"

	adminService "accelerator/internal/features/admin/service"
	authService "accelerator/internal/features/auth/service"
	tasksService "accelerator/internal/features/tasks/service"

	adminTransport "accelerator/internal/features/admin/transport"
	authTransport "accelerator/internal/features/auth/transport"
	tasksTransport "accelerator/internal/features/tasks/transport"
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"github.com/jackc/pgx/v5/pgxpool"
)

// что можно добавить в будущем для безопасности?
// 2) хранение в сессии еще и fingerprint, чтобы привязывать сессию к определенному устройству чтобы обрабатывать подозрения на угон аккаунта
// мягкое удаление пользователей
// зашить роль в токене, чтобы сразу отрезать пользователей от админки без запроса к бд, но при этом еще и отзывать токены уметь, чтобы при смене роли токен сразу отзывался, а не работал еще 15 минут
// добавить валидацию на входе для всех исходя из ограничений базы данных
// добавить кастомный обработчик ошибок из валидатора, чтобы ловить все все ошибки и отслыать их на клиент с разными сообщениями

// пофиксить баги admin_service 821 line 
// пофиксить баги repository 588 line 

func main() {
	cfg := config.LoadConfig()  // загружаем .env и все его значения
	
	logger.InitLog()            // инициализируем логер, чтобы нормально записывать в файл
	
	validate := validator.New() // создаем валидатор, чтобы потом передать в хэндлеры
	validate.RegisterValidation("fio", tools.ValidateFio)


	pool, err := pgxpool.New(context.Background(), cfg.DBDSN) // создаем пул соединений
	defer func() { pool.Close() }() // перед завершением работы закрываем соединение с базой данных
	if err != nil {
		slog.Error("Не удалось создать пул соединений с базой данных:", "err", err)
	}

	adminRepo := adminRepository.NewAdminRepository(pool)
	adminServ := adminService.NewAdminService(adminRepo, cfg)
	adminTrans := adminTransport.NewAdminTransport(adminServ, validate)

	authRepo := authRepository.NewAuthRepo(pool)          // возвращает указатель на репозиторий с указателем на подключение к базе данных и соответсвенно методы работы с бд
	authServ := authService.NewAuthService(authRepo, cfg) // передаем методы работы с бд в бизнес логику, возвращает методы работы бизнес логики
	authTrans := authTransport.NewAuthTransport(authServ, validate) // передаем нашу методы из бизнес логики и созданный валидатор

	TasksRepo := tasksRepository.NewTasksRepo(pool)          
	TasksServ := tasksService.NewTasksService(TasksRepo, cfg)
	TasksTrans := tasksTransport.NewTasksTransport(TasksServ, validate, cfg)


	if err := server.StartNewChiServer(authTrans, TasksTrans, adminTrans, cfg); err != nil {
		slog.Error("Ошибка при работе HTTP сервера:", "err", err)
	} else {
		slog.Info("Сервер завершился успешно")
	}
}
