package server

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/server/middleware"
	adminTransport "accelerator/internal/features/admin/transport"
	authTransport "accelerator/internal/features/auth/transport"
	tasksTransport "accelerator/internal/features/tasks/transport"
	"accelerator/internal/tools"
	"errors"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func StartNewChiServer(
		authTrans *authTransport.AuthTransport, 
		tasksTrans *tasksTransport.TasksTransport, 
		adminTrans *adminTransport.AdminTransport, 
		cfg *config.Config,
	) error {
	router := chi.NewRouter() // используем chi, он легковесный , в нем есть встроенные обработчики переменных в паттерне и нормальный роутинг

	router.Use(chiMiddleware.RequestID) // генерирует для каждого запроса уникальный ID
	router.Use(chiMiddleware.RealIP)    // позволяет видеть реальный IP пользователя для логера
	router.Use(chiMiddleware.Logger)    // Потом, чтобы логировать всё, включая айдишник

	// создаем группу api с префиксом
	router.Route("/api/v1", func(router chi.Router) {
		// эндпоинт, который доступен только при первом запуске приложения с пустой таблицей users
		router.Post("/creator/registration", adminTrans.AddCreatorHandle)

		// создаем группу c admin префиксом
		router.Route("/admin", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg)) // проверяет токен и загружает в контекст ID
			router.Post("/users", adminTrans.RegisterNewUserHandle)
			router.Get("/users", adminTrans.GetUsersHandle) 
			router.Put("/users/{userID}", adminTrans.EditUserHandle)
			router.Post("/users/{userID}/reset-password", adminTrans.ResetPasswordHandle)
			router.Delete("/users/{userID}", adminTrans.DeleteUserHandle)

			router.Post("/groups", adminTrans.CreateGroupHandle)
			router.Get("/groups/{groupID}", adminTrans.GetMembersGroupHandle)
			router.Get("/groups", adminTrans.GetGroupsHandle)
			router.Put("/groups/{groupID}", adminTrans.EditGroupHandle)
			router.Post("/groups/{groupID}/members/{userID}", adminTrans.AddUserGroupHandle)
			router.Delete("/groups/{groupID}/members/{userID}", adminTrans.DeleteUserGroupHandle)
			router.Delete("/groups/{groupID}", adminTrans.DeleteGroupHandle)
		})

		// авторизация и обновление токенов
		router.Route("/auth", func(router chi.Router) {
			router.Post("/login", authTrans.LoginHandle)
			router.Post("/refresh", authTrans.RefreshHandle)
		})

		// основная работа с задачами
		router.Route("/tasks", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg))
			router.Post("/upload", tasksTrans.UploadHandle)
		})

		// работа со своим аккаунтом
		router.Route("/users/me", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg))
			router.Post("/change-temp-password", authTrans.ChangeTempPasswordHandle)
		})
	})

	

	// делаем кастомное сообщение, такое же как и при запрете доступа к админке, чтобы нельзя было его опознать
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		tools.WriteError(w, error_type.NewNotFound("Страница не найдена"))
	})


	err := http.ListenAndServe(cfg.ServerPort, router)

	if errors.Is(err, http.ErrServerClosed) { // если ошибка при нормальном завершении сервера. то все ок
		return nil
	} else {
		return err
	}
	
}
