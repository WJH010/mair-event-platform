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

	filectr "event-platform/internal/file/controller"
	filerepo "event-platform/internal/file/repository"
	filesvc "event-platform/internal/file/service"

	msgctr "event-platform/internal/message/controller"
	msgrepo "event-platform/internal/message/repository"
	msgsvc "event-platform/internal/message/service"

	eventctr "event-platform/internal/event/controller"
	eventmodel "event-platform/internal/event/model"
	eventrepo "event-platform/internal/event/repository"
	eventsvc "event-platform/internal/event/service"
	eventstock "event-platform/internal/event/stock"

	fieldctr "event-platform/internal/field/controller"
	fieldrepo "event-platform/internal/field/repository"
	fieldsvc "event-platform/internal/field/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 注册中间件和路由
func SetupRoutes(cfg *config.Config, router *gin.Engine, minioRepo filerepo.MinIORepository) {
	db := database.GetDB()

	// 初始化依赖
	// 初始化仓库
	fileRepo := filerepo.NewFileRepository(db)
	userRepo := userrepo.NewUserRepository(db)
	industryRepo := userrepo.NewIndustryRepository(db)
	msgRepo := msgrepo.NewMessageRepository(db)
	eventRepo := eventrepo.NewEventRepository(db)
	eventUserInfoRepo := eventrepo.NewEventUserInfoRepository(db)
	msgGroupRepo := msgrepo.NewMsgGroupRepository(db, msgRepo)
	userRoleRepo := userrepo.NewUserRoleRepository(db)
	fieldRepo := fieldrepo.NewFieldRepository(db)

	// 初始化服务
	fileService := filesvc.NewFileService(minioRepo, fileRepo)
	msgService := msgsvc.NewMessageService(msgRepo, msgGroupRepo)
	msgGroupService := msgsvc.NewMsgGroupService(msgGroupRepo, msgRepo)
	userService := usersvc.NewUserService(userRepo, msgGroupService, cfg)
	industryService := usersvc.NewIndustryService(industryRepo)
	eventService := eventsvc.NewEventService(eventRepo, eventUserInfoRepo, userRepo, fileRepo, eventstock.NewStockService(), cache.New[int, *eventmodel.Event](3*time.Second))
	eventUserInfoService := eventsvc.NewEventUserInfoService(eventUserInfoRepo)
	userRoleService := usersvc.NewUserRoleService(userRoleRepo)
	fieldService := fieldsvc.NewFieldService(fieldRepo)

	// 初始化控制器
	fileController := filectr.NewFileController(fileService)
	userController := userctr.NewUserController(userService)
	industryController := userctr.NewIndustryController(industryService)
	msgController := msgctr.NewMessageController(msgService)
	eventController := eventctr.NewEventController(eventService)
	eventUserInfoController := eventctr.NewEventUserInfoController(eventUserInfoService)
	msgGroupController := msgctr.NewMsgGroupController(msgGroupService)
	userRoleController := userctr.NewUserRoleController(userRoleService)
	fieldController := fieldctr.NewFieldController(fieldService)

	// API分组
	api := router.Group("/api")
	{
		// 用户相关路由
		user := api.Group("/user")
		{
			// 公开接口 - 无需认证
			// 注册用户
			user.POST("/register", userController.RegisterUser)
			// 登录
			user.POST("/Login", userController.Login)
			// 刷新token接口
			user.POST("/refreshToken", userController.RefreshToken)
			// 需要认证的用户接口
			authUser := user.Group("")
			authUser.Use(middleware.AuthMiddleware(cfg))
			{
				// 更新当前用户信息
				authUser.PUT("/update", userController.UpdateUserInfo)
				// 修改密码
				authUser.PUT("/changePassword", userController.ChangePassword)
				// 获取当前用户信息
				authUser.GET("/info", userController.GetUserInfo)
				// 退出登录
				authUser.POST("/logout", userController.Logout)
				// 管理员接口 - 在认证基础上增加角色校验
				adminUser := authUser.Group("")
				adminUser.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					// 分页查询系统用户列表
					adminUser.GET("/listAll", userController.ListAllUsers)
					// 禁用/启用用户接口
					adminUser.PUT("/updateStatus/:id", userController.UpdateUserStatus)
				}
				// 超级管理员接口
				superAdmin := authUser.Group("")
				superAdmin.Use(middleware.RoleMiddleware(utils.RoleSuperAdmin))
				{
					// 更新用户角色
					superAdmin.PUT("/updateRole/:id", userController.UpdateUserRole)
				}
			}
		}
		// 行业路由
		industry := api.Group("/industry")
		{
			// 公开接口 - 无需认证
			// 获取行业列表（不分页）
			industry.GET("", industryController.ListIndustries)
			// 需要认证的用户接口
			authIndustry := industry.Group("")
			authIndustry.Use(middleware.AuthMiddleware(cfg))
			{
				// 管理员接口 - 在认证基础上增加角色校验
				adminIndustry := authIndustry.Group("")
				adminIndustry.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					// 创建行业
					adminIndustry.POST("/create", industryController.CreateIndustry)
					// 更新行业
					adminIndustry.PUT("/update/:id", industryController.UpdateIndustry)
				}
			}
		}
		// 用户角色路由
		userRole := api.Group("/userRole")
		{
			// 获取用户角色列表（不分页）
			userRole.GET("", middleware.RoleMiddleware(utils.RoleAdmin), userRoleController.List)
		}
		// 文件上传路由
		file := api.Group("/file")
		file.Use(middleware.AuthMiddleware(cfg))
		{
			// 上传文件
			file.POST("/upload", fileController.UploadFile)
			// 删除文件
			file.DELETE("/deleteImage/:id", fileController.DeleteImage)

		}
		// 消息相关路由
		message := api.Group("/message")
		message.Use(middleware.AuthMiddleware(cfg))
		{
			// 获取消息详情
			message.GET("/:id", msgController.GetMessageContent)
			// 查询是否有未读消息
			message.GET("/hasUnreadMessages", msgController.HasUnreadMessages)
			// 标记所有消息为已读
			message.PUT("/markAllAsRead", msgController.MarkAllMessagesAsRead)
			// 分页查询用户消息群组列表
			message.GET("/userMessageGroups", msgController.ListUserMessageGroups)
			// 分页查询组内消息列表
			message.GET("/byGroups/:id", msgController.ListMsgByGroups)
			// 消息群组管理，仅管理员可操作
			adminMessage := message.Group("")
			adminMessage.Use(middleware.RoleMiddleware(utils.RoleAdmin))
			{
				// 根据消息群组ID查询消息列表
				adminMessage.GET("/allByGroupID/:id", msgController.ListMessagesByGroupID)
				// 分页获取消息群组列表
				adminMessage.GET("/messageGroups", msgGroupController.ListMsgGroups)
				// 分页获取指定群组内用户列表
				adminMessage.GET("/groupUsers/:id", msgGroupController.ListGroupsUsers)
				// 分页获取不在指定组内的用户列表
				adminMessage.GET("/notIngroupUsers/:id", msgGroupController.ListNotInGroupUsers)
				// 根据id获取指定消息群组信息
				adminMessage.GET("/groupDetail/:id", msgGroupController.GetMsgGroupByID)
				// 创建消息群组
				adminMessage.POST("/createGroup", msgGroupController.CreateMsgGroup)
				// 将用户添加到指定群组
				adminMessage.POST("/addUserToGroup/:id", msgGroupController.AddUserToGroup)
				// 向指定群组发送消息
				adminMessage.POST("/sendMessage/:id", msgController.SendMessage)
				// 更新指定群组信息
				adminMessage.PUT("/updateGroup/:id", msgGroupController.UpdateMsgGroup)
				// 撤销指定消息
				adminMessage.DELETE("/revokeMessage/:id", msgController.RevokeGroupMessage)
				// 从指定群组移除用户
				adminMessage.DELETE("/removeUserFromGroup/:id", msgGroupController.DeleteUserFromGroup)
				// 删除群组
				adminMessage.DELETE("/deleteGroup/:id", msgGroupController.DeleteMsgGroup)
			}
		}
		// 活动相关路由
		event := api.Group("/event")
		{
			// 公开接口 - 无需认证
			// 分页查询活动列表
			event.GET("", eventController.ListEvent)
			// 获取指定活动详情
			event.GET("/:id", eventController.GetEventDetail)

			// 需要认证的用户接口
			authEvent := event.Group("")
			authEvent.Use(middleware.AuthMiddleware(cfg))
			{
				// 报名活动
				authEvent.POST("/registration", eventController.RegistrationEvent)
				// 查询当前用户是否报名指定活动
				authEvent.GET("/isUserRegistered/:id", eventController.IsUserRegistered)
				// 取消报名活动
				authEvent.DELETE("/cancelRegistration/:id", eventController.CancelRegistrationEvent)
				// 获取当前用户已报名的活动列表
				authEvent.GET("/userRegisteredEvents", eventController.ListUserRegisteredEvents)

				// 管理员接口 - 在认证基础上增加角色校验
				adminEvent := authEvent.Group("")
				adminEvent.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					// 创建活动
					adminEvent.POST("/create", eventController.CreateEvent)
					// 更新活动
					adminEvent.PUT("/update/:id", eventController.UpdateEvent)
					// 删除活动
					adminEvent.DELETE("/delete/:id", eventController.DeleteEvent)
					// 分页查询报名指定活动的用户列表
					adminEvent.GET("/regUsers/:id", eventController.ListEventRegisteredUsers)
					// 创建用户信息字段
					adminEvent.POST("/createUserInfo", eventUserInfoController.Create)
					// 更新用户信息字段
					adminEvent.PUT("/updateUserInfo/:id", eventUserInfoController.Update)
					// 更新用户信息字段状态
					adminEvent.PUT("/updateUserInfoStatus/:id", eventUserInfoController.UpdateStatus)
					// 查询用户信息字段列表
					adminEvent.GET("/userInfo", eventUserInfoController.List)
				}
			}
		}
		// 领域路由
		field := api.Group("/field")
		{
			// 公开接口 - 无需认证
			// 分页查询领域列表
			field.GET("", fieldController.ListFields)
			authField := field.Group("")
			authField.Use(middleware.AuthMiddleware(cfg))
			{
				// 管理员接口 - 在认证基础上增加角色校验
				adminField := authField.Group("")
				adminField.Use(middleware.RoleMiddleware(utils.RoleAdmin))
				{
					// 创建领域
					adminField.POST("/create", fieldController.CreateField)
					// 更新领域
					adminField.PUT("/update/:id", fieldController.UpdateField)
					// 删除领域
					adminField.DELETE("/delete/:id", fieldController.DeleteField)
					// 更新领域状态
					adminField.PUT("/updateStatus/:id", fieldController.UpdateFieldStatus)
				}
			}
		}
	}
}
