package di

import (
	"database/sql"
	accountHandler "fluencybe/internal/app/handler/account"
	accountRepo "fluencybe/internal/app/repository/account"
	accountSer "fluencybe/internal/app/service/account"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
)

type AccountModule struct {
	UserHandler      *accountHandler.UserHandler
	DeveloperHandler *accountHandler.DeveloperHandler
}

func ProvideAccountModule(
	dbConn *sql.DB,
	redisClient *cache.RedisClient,
	log *logger.PrettyLogger,
) *AccountModule {
	// Repositories
	userRepo := accountRepo.NewUserRepository(dbConn, redisClient, log)
	developerRepo := accountRepo.NewDeveloperRepository(dbConn, redisClient)

	// Services
	userService := accountSer.NewUserService(userRepo)
	developerService := accountSer.NewDeveloperService(developerRepo)

	// Handlers
	userHandler := accountHandler.NewUserHandler(userService)
	developerHandler := accountHandler.NewDeveloperHandler(developerService)

	return &AccountModule{
		UserHandler:      userHandler,
		DeveloperHandler: developerHandler,
	}
}
