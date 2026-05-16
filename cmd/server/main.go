package main

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/logger"
	"accelerator/internal/core/server"
	adminRepository "accelerator/internal/features/admin/repository"
	authRepository "accelerator/internal/features/auth/repository"
	patternsRepository "accelerator/internal/features/patterns/repository"
	tasksRepository "accelerator/internal/features/tasks/repository"
	"accelerator/internal/tools"

	adminService "accelerator/internal/features/admin/service"
	authService "accelerator/internal/features/auth/service"
	patternsService "accelerator/internal/features/patterns/service"
	tasksService "accelerator/internal/features/tasks/service"

	adminTransport "accelerator/internal/features/admin/transport"
	authTransport "accelerator/internal/features/auth/transport"
	patternsTransport "accelerator/internal/features/patterns/transport"
	tasksTransport "accelerator/internal/features/tasks/transport"
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"github.com/jackc/pgx/v5/pgxpool"
)

// что можно добавить в будущем?

// хранение в сессии еще и fingerprint, чтобы привязывать сессию к определенному устройству чтобы обрабатывать подозрения на угон аккаунта
// мягкое удаление пользователей
// зашить роль в токене, чтобы сразу отрезать пользователей от админки без запроса к бд, но при этом еще и отзывать токены уметь, чтобы при смене роли токен сразу отзывался, а не работал еще 15 минут
// добавить валидацию на входе для всех исходя из ограничений базы данных
// добавить кастомный обработчик ошибок из валидатора, чтобы ловить все все ошибки и отслыать их на клиент с разными сообщениями


// во всех read-write методах сделать транзакции, обновить методы репозитория через executor и добавить во все SELECT запросы for update





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

	patternsRepo := patternsRepository.NewPatternsRepository(pool)          
	patternsServ := patternsService.NewPatternsService(patternsRepo)
	patternsTrans := patternsTransport.NewPatternsTransport(patternsServ, validate)

	tasksRepo := tasksRepository.NewTasksRepo(pool)          
	tasksServ := tasksService.NewTasksService(tasksRepo, cfg)
	tasksTrans := tasksTransport.NewTasksTransport(tasksServ, validate, cfg)


	if err := server.StartNewChiServer(adminTrans, authTrans, patternsTrans, tasksTrans, cfg); err != nil {
		slog.Error("Ошибка при работе HTTP сервера:", "err", err)
	} else {
		slog.Info("Сервер завершился успешно")
	}
}
