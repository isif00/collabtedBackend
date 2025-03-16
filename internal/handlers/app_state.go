package handlers

import (
	"net/http"

	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/labstack/echo/v4"
)

type AppStateHandler struct {
	srv services.AppStateService
}

func NewAppStateHandler() *AppStateHandler {
	return &AppStateHandler{
		srv: *services.NewAppStateService(),
	}
}

func (h *AppStateHandler) GetAppState(c echo.Context) error {
	userWorkspaceId := c.Param("userworkspaceId")
	appState, err := h.srv.GetAppState(userWorkspaceId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, appState)
}

func (h *AppStateHandler) UpdateAppState(c echo.Context) error {
	userWorkspaceId := c.Param("userworkspaceId")

	var req types.AppStateUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request format")
	}

	appState, err := h.srv.UpdateAppState(userWorkspaceId, req.Action, req.Value)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, appState)
}
