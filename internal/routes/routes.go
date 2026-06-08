package routes

import (
	"event-platform/internal/cache"
	"event-platform/internal/config"
	"event-platform/internal/database"
	"event-platform/internal/middleware"
	"event-platform/internal/utils"
	"time"

	userctr "event-platform/internal/user/controller"
	userrepo "event-platform/internal/user/repository"
	usersvc "event-platform/internal/user/service"

	industryctr "event-platform/internal/industry/controller"
	industryrepo "event-platform/internal/industry/repository"
	industrysvc "event-platform/internal/industry/service"

	filectr "event-platform/internal/file/controller"
	filerepo "event-platform/internal/file/repository"
	filesvc "event-platform/internal/file/service"

	eventctr "event-platform/internal/event/controller"
	eventmodel "event-platform/internal/event/model"
	eventrepo "event-platform/internal/event/repository"
	eventsvc "event-platform/internal/event/service"
	eventstock "event-platform/internal/event/stock"

	fieldctr "event-platform/internal/field/controller"
	fieldrepo "event-platform/internal/field/repository"
	fieldsvc "event-platform/internal/field/service"

	messagectr "event-platform/internal/message/controller"
	messagerepo "event-platform/internal/message/repository"
	messagesvc "event-platform/internal/message/service"

	dashboardctr "event-platform/internal/dashboard/controller"
	dashboardrepo "event-platform/internal/dashboard/repository"
	dashboardsvc "event-platform/internal/dashboard/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 注册中间件和路由
func SetupRoutes(cfg *config.Config, router *gin.Engine, minioRepo filerepo.MinIORepository) {
	db := database.GetDB()

	// 初始化依赖
	// 初始化仓库
	fileRepo := filerepo.NewFileRepository(db)
	userRepo := userrepo.NewUserRepository(db)
	industryRepo := industryrepo.NewIndustryRepository(db)
	eventRepo := eventrepo.NewEventRepository(db)
	eventUserInfoRepo := eventrepo.NewEventUserInfoRepository(db)
	userRoleRepo := userrepo.NewUserRoleRepository(db)
	fieldRepo := fieldrepo.NewFieldRepository(db)
	messageRepo := messagerepo.NewMessageRepository(db)
	dashboardRepo := dashboardrepo.NewDashboardRepository(db)

	// 初始化服务
	fileService := filesvc.NewFileService(minioRepo, fileRepo)
	userService := usersvc.NewUserService(userRepo, cfg)
	industryService := industrysvc.NewIndustryService(industryRepo)
	eventService := eventsvc.NewEventService(eventRepo, eventUserInfoRepo, userRepo, fileRepo, eventstock.NewStockService(), cache.New[int, *eventmodel.Event](3*time.Second))
	eventUserInfoService := eventsvc.NewEventUserInfoService(eventUserInfoRepo)
	userRoleService := usersvc.NewUserRoleService(userRoleRepo)
	fieldService := fieldsvc.NewFieldService(fieldRepo)
	messageService := messagesvc.NewMessageService(messageRepo, userRepo)
	dashboardService := dashboardsvc.NewDashboardService(dashboardRepo, userRepo)

	// 初始化控制器
	fileController := filectr.NewFileController(fileService)
	userController := userctr.NewUserController(userService)
	industryController := industryctr.NewIndustryController(industryService)
	eventController := eventctr.NewEventController(eventService)
	eventUserInfoController := eventctr.NewEventUserInfoController(eventUserInfoService)
	userRoleController := userctr.NewUserRoleController(userRoleService)
	fieldController := fieldctr.NewFieldController(fieldService)
	messageController := messagectr.NewMessageController(messageService)
	dashboardController := dashboardctr.NewDashboardController(dashboardService)

	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("/sms/send", userController.SendSMS)
			users.POST("/sms/verify", userController.VerifySMS)
			users.POST("/register", userController.RegisterUser)
			users.POST("/login", userController.Login)
			users.POST("/login/sms", userController.SMSLogin)
			users.POST("/password/reset", userController.ResetPassword)
			users.POST("/token/refresh", userController.RefreshToken)

			authUsers := users.Group("")
			authUsers.Use(middleware.AuthMiddleware(cfg))
			{
				authUsers.POST("/logout", userController.Logout)
				authUsers.GET("/me", userController.GetUserInfo)
				authUsers.PUT("/me", userController.UpdateUserInfo)
				authUsers.PATCH("/me/password", userController.ChangePassword)

				adminUsers := authUsers.Group("")
				adminUsers.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					adminUsers.GET("", userController.ListAllUsers)
					adminUsers.PATCH("/:id/status", userController.UpdateUserStatus)
				}

				superAdminUsers := authUsers.Group("")
				superAdminUsers.Use(middleware.RoleMiddleware(utils.RoleSuperAdmin))
				{
					superAdminUsers.PATCH("/:id/role", userController.UpdateUserRole)
				}
			}
		}

		userRoles := api.Group("/user-roles")
		userRoles.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware(utils.RoleAdmin))
		{
			userRoles.GET("", userRoleController.List)
		}

		files := api.Group("/files")
		files.Use(middleware.AuthMiddleware(cfg))
		{
			files.POST("", fileController.UploadFile)
			files.DELETE("/:id", fileController.DeleteImage)
		}

		events := api.Group("/events")
		{
			authEvents := events.Group("")
			authEvents.Use(middleware.AuthMiddleware(cfg))
			{
				authEvents.GET("", eventController.ListEvent)
				authEvents.GET("/:id", eventController.GetEventDetail)
				authEvents.GET("/me/registrations", eventController.ListUserRegisteredEvents)
				authEvents.POST("/:id/registrations", eventController.RegistrationEvent)
				authEvents.GET("/:id/registrations/me", eventController.IsUserRegistered)
				authEvents.DELETE("/:id/registrations/me", eventController.CancelRegistrationEvent)

				adminEvents := authEvents.Group("")
				adminEvents.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					adminEvents.POST("", eventController.CreateEvent)
					adminEvents.PUT("/:id", eventController.UpdateEvent)
					adminEvents.DELETE("/:id", eventController.DeleteEvent)
					adminEvents.GET("/:id/registrations", eventController.ListEventRegisteredUsers)
				}
			}
		}

		userInfoFields := api.Group("/user-info-fields")
		userInfoFields.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware(utils.RoleAdmin))
		{
			userInfoFields.GET("", eventUserInfoController.List)
			userInfoFields.POST("", eventUserInfoController.Create)
			userInfoFields.PUT("/:id", eventUserInfoController.Update)
			userInfoFields.PATCH("/:id/status", eventUserInfoController.UpdateStatus)
		}

		fields := api.Group("/fields")
		{
			fields.GET("", fieldController.ListFields)
			authFields := fields.Group("")
			authFields.Use(middleware.AuthMiddleware(cfg))
			{
				adminFields := authFields.Group("")
				adminFields.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					adminFields.POST("", fieldController.CreateField)
					adminFields.PUT("/:id", fieldController.UpdateField)
					adminFields.DELETE("/:id", fieldController.DeleteField)
					adminFields.PATCH("/:id/status", fieldController.UpdateFieldStatus)
				}
			}
		}

		industries := api.Group("/industries")
		{
			industries.GET("", industryController.ListIndustries)
			authIndustries := industries.Group("")
			authIndustries.Use(middleware.AuthMiddleware(cfg))
			{
				adminIndustries := authIndustries.Group("")
				adminIndustries.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					adminIndustries.POST("", industryController.CreateIndustry)
					adminIndustries.PUT("/:id", industryController.UpdateIndustry)
					adminIndustries.DELETE("/:id", industryController.DeleteIndustry)
					adminIndustries.PATCH("/:id/status", industryController.UpdateIndustryStatus)
				}
			}
		}

		messages := api.Group("/messages")
		{
			authMessages := messages.Group("")
			authMessages.Use(middleware.AuthMiddleware(cfg))
			{
				authMessages.GET("/me", messageController.ListUserMessages)
				authMessages.GET("/me/unread-count", messageController.GetUnreadCount)

				adminMessages := authMessages.Group("")
				adminMessages.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					adminMessages.POST("", messageController.CreateMessage)
					adminMessages.DELETE("/:id", messageController.RevokeMessage)
					adminMessages.GET("", messageController.ListMessages)
				}
			}
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware(utils.RoleAdmin))
		{
			dashboard.GET("/users", dashboardController.ListDashboardUsers)
			dashboard.GET("/events", dashboardController.ListDashboardEvents)
			dashboard.GET("/events/:id", dashboardController.GetEventOverview)
			dashboard.GET("/events/:id/registrations", dashboardController.ListEventRegUsers)
			dashboard.GET("/events/:id/statistics", dashboardController.GetEventStatistics)
		}
	}
}
