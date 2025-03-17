package router

import (
	"github.com/CollabTED/CollabTed-Backend/internal/handlers"
	middlewares "github.com/CollabTED/CollabTed-Backend/internal/middlewares/rest"
	"github.com/labstack/echo/v4"
)

func AppStateRouter(e *echo.Group) {
	h := handlers.NewAppStateHandler()
	appState := e.Group("/app-state", middlewares.AuthMiddleware)
	appState.GET("/:userworkspaceId", h.GetAppState)
	appState.PATCH("/:userworkspaceId", h.UpdateAppState)
}
