package server

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/error_type"
	"accelerator/internal/core/server/middleware"
	adminTransport "accelerator/internal/features/admin/transport"
	authTransport "accelerator/internal/features/auth/transport"
	patternsTransport "accelerator/internal/features/patterns/transport"
	tasksTransport "accelerator/internal/features/tasks/transport"
	"accelerator/internal/tools"
	"errors"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func StartNewChiServer(
	adminTrans *adminTransport.AdminTransport,
	authTrans *authTransport.AuthTransport,
	patternsTrans *patternsTransport.PatternsTransport,
	tasksTrans *tasksTransport.TasksTransport,

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
			router.Route("/users", func(router chi.Router) {
				// регистрация нового пользователя, креатор може всех, админ только user
				router.Post("/", adminTrans.RegisterNewUserHandle)
				// получить пользователей, креатор видит всех, кроме себя, админ только user
				router.Get("/", adminTrans.GetUsersHandle)
				// изменить пользователя, креатор можетменять всех и повышать user до admin, а админ может менять только user и не может менять роль
				router.Put("/{userID}", adminTrans.EditUserHandle)
				// сбрасывает пароль, креатор может сбросить кому угодно, кроме себя, а админ только user
				router.Post("/{userID}/reset-password", adminTrans.ResetPasswordHandle)
				// удалить пользователя, креатор всех кроме себя, админ только user
				router.Delete("/{userID}", adminTrans.DeleteUserHandle)
			})
			router.Route("/groups", func(router chi.Router) {
				// создание группы, доступно только администратору
				router.Post("/", adminTrans.CreateGroupHandle)
				// Получение списка групп (каждый видит свои)
				router.Get("/", adminTrans.GetGroupsHandle)
				// Информация о группе + участники, креатор видит всех, админ только user, себя не видит
				router.Get("/{groupID}", adminTrans.GetMembersGroupHandle)
				// Редактирование группы (права проверяются в сервисе: creator – любую, admin – только свою)
				router.Put("/{groupID}", adminTrans.EditGroupHandle)
				// Удаление группы (аналогично)
				router.Delete("/{groupID}", adminTrans.DeleteGroupHandle)
				// Добавление участника в группу, админ только user, креатор тоже
				router.Post("/{groupID}/members/{userID}", adminTrans.AddUserGroupHandle)
				// Удаление участника из группы
				router.Delete("/{groupID}/members/{userID}", adminTrans.DeleteUserGroupHandle)
			})
		})

		// авторизация и обновление токенов
		router.Route("/auth", func(router chi.Router) {
			router.Post("/login", authTrans.LoginHandle)
			router.Post("/refresh", authTrans.RefreshHandle)
		})

		// основная работа с задачами
		router.Route("/tasks", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg))
			// загрузка аудио и прочей информации для транскрибации
			router.Post("/upload", tasksTrans.UploadHandle)
		})

		// работа с шаблонами
		router.Route("/patterns", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg))
			// создание шаблона
			router.Post("/", patternsTrans.CreatePatternHandler)
			// получение информации о шаблоне по ID
			router.Get("/{patternID}", patternsTrans.GetPattern)
			// получение доступных шаблонов для группы по ID
			router.Get("/{groupID}", patternsTrans.GetGroupPatterns)
			// получение созданных шаблонов для креатора
			router.Get("/global", patternsTrans.GetCreatorPatterns)
			// получение всех шаблонов по группам для креатора
			router.Get("/all", patternsTrans.GetAllPatternsInGroups)
			// изменение шаблона
			router.Put("/{patternID}", patternsTrans.EditPattern)
			// удаление шаблона
			router.Delete("/{patternID}", patternsTrans.DeletePattern)
		})

		// работа со своим аккаунтом
		router.Route("/users/me", func(router chi.Router) {
			router.Use(middleware.AuthMiddleware(cfg))
			// изменение временного пароля
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
