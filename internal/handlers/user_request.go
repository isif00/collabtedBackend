package handlers

import (
	"net/http"

	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/labstack/echo/v4"
)

type UserRequestHandler struct {
	UserRequestService *services.UserRequestService
}

func NewUserRequestHandler() *UserRequestHandler {
	return &UserRequestHandler{
		UserRequestService: services.NewUserRequestService(),
	}
}

func (u *UserRequestHandler) GetRequests(c echo.Context) error {
	requests, err := u.UserRequestService.GetRequests()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, requests)

}

func (u *UserRequestHandler) SaveRequest(c echo.Context) error {
	var data types.UserRequest
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	request, err := u.UserRequestService.SaveRequest(data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, request)
}
