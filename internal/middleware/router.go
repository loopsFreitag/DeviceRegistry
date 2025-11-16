package middleware

import (
	"github.com/gorilla/mux"
	"github.com/loopsFreitag/DeviceRegistry/internal/controller"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"github.com/loopsFreitag/DeviceRegistry/internal/service"

	_ "github.com/loopsFreitag/DeviceRegistry/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewAppRouter() *mux.Router {
	router := mux.NewRouter()

	userRepo := repository.NewUserRepository(model.DBX())
	authService := service.NewAuthService(userRepo)

	authMiddleware := NewAuthMiddleware(authService)

	// Public routes
	controller.NewHealthCheck(controller.WithDBChecker()).SetRoutes(router)
	controller.NewAuthController(authService).SetRoutes(router)

	// Protected routes
	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(authMiddleware.RequireAuth)
	controller.NewDeviceController().SetRoutes(protectedRouter)

	// Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return router
}
