package router

import (
	"github.com/CollabTED/CollabTed-Backend/internal/handlers"
	middlewares "github.com/CollabTED/CollabTed-Backend/internal/middlewares/rest"
	"github.com/labstack/echo/v4"
)

func UserRequestRouter(e *echo.Group) {
	h := handlers.NewUserRequestHandler()
	UserRequest := e.Group("/user-requests")
	UserRequest.GET("/", h.GetRequests, middlewares.AuthMiddleware)
	UserRequest.POST("/", h.SaveRequest)
}
