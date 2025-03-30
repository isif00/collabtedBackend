package router

import (
	"github.com/CollabTED/CollabTed-Backend/internal/handlers"
	middlewares "github.com/CollabTED/CollabTed-Backend/internal/middlewares/rest"
	"github.com/labstack/echo/v4"
)

func SubscriptionRoutes(e *echo.Group) {
	h := handlers.NewSubscriptionHandler()
	Subscriptions := e.Group("/subscription", middlewares.AuthMiddleware)
	Subscriptions.GET("/:email", h.GetUserSubscription)
}
