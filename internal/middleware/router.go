package middleware

import (
	"github.com/gorilla/mux"
	"github.com/loopsFreitag/DeviceRegistry/internal/controller"
)

func NewAppRouter() *mux.Router {
	router := mux.NewRouter()

	controller.NewHealthCheck(controller.WithDBChecker()).SetRoutes(router)

	return router
}
