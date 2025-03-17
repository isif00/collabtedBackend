package services

import (
	"context"

	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type UserRequestService struct {
}

func NewUserRequestService() *UserRequestService {
	return &UserRequestService{}
}

func (u *UserRequestService) GetRequests() ([]db.UserRequestModel, error) {
	requests, err := prisma.Client.UserRequest.FindMany().Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return requests, err
}

func (u *UserRequestService) SaveRequest(data types.UserRequest) (*db.UserRequestModel, error) {
	request, err := prisma.Client.UserRequest.CreateOne(
		db.UserRequest.Type.Set(db.RequestType(data.Type)),
		db.UserRequest.Email.Set(data.Email),
		db.UserRequest.Request.Set(data.Request),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return request, err
}
