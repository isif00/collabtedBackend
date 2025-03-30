package handlers

import (
	"net/http"

	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	srv services.SubscriptionService
}

func NewSubscriptionHandler() *SubscriptionHandler {
	return &SubscriptionHandler{
		srv: *services.NewSubscriptionService(),
	}
}

func (s *SubscriptionHandler) GetUserSubscription(c echo.Context) error {
	var email = c.Param("email")

	subscription, err := s.srv.GetUserSubscription(email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, subscription)
}
