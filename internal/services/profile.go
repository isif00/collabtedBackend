package services

import (
	"context"
	"fmt"

	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type ProfileService struct{}

func NewProfileService() *ProfileService {
	return &ProfileService{}
}

func (s *ProfileService) GetUser(id string) (*db.UserModel, error) {
	logger.LogInfo().Fields(map[string]interface{}{
		"query":  "get profile",
		"params": id,
	}).Msg("DB Query")
	ctx := context.Background()
	user, err := prisma.Client.User.FindUnique(
		db.User.ID.Equals(id),
	).Omit(db.User.Password.Field()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *ProfileService) GetUserByEmail(email string) (*db.UserModel, error) {
	logger.LogInfo().Fields(map[string]interface{}{
		"query":  "search profile",
		"params": email,
	}).Msg("DB Query")

	ctx := context.Background()
	user, err := prisma.Client.User.FindUnique(
		db.User.Email.Equals(email),
	).Omit(db.User.Password.Field()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *ProfileService) SearchByName(name string) ([]db.UserModel, error) {
	logger.LogInfo().Fields(map[string]interface{}{
		"query":  "search profile",
		"params": name,
	}).Msg("DB Query")

	ctx := context.Background()
	users, err := prisma.Client.User.FindMany(
		db.User.Name.Contains(name),
	).Omit(db.User.Password.Field()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *ProfileService) UpdateUser(id string, payload types.ProfileUpdate) (*db.UserModel, error) {
	logger.LogInfo().Fields(map[string]interface{}{
		"query":  "update profile",
		"id":     id,
		"params": payload,
	}).Msg("DB Query")

	ctx := context.Background()
	users, err := prisma.Client.User.FindUnique(
		db.User.ID.Equals(id),
	).Omit(db.User.Password.Field()).Update(
		db.User.Email.Set(payload.Email),
		db.User.Name.Set(payload.Name),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (s *ProfileService) DeleteUser(id string) (string, error) {
	logger.LogInfo().Fields(map[string]interface{}{
		"query":  "delete profile",
		"params": id,
	}).Msg("DB Query")

	ctx := context.Background()
	deleted, err := prisma.Client.User.FindUnique(
		db.User.ID.Equals(id),
	).Delete().Exec(ctx)
	if err != nil {
		return "", nil
	}
	fmt.Println(deleted.ID)
	return deleted.ID, nil
}
