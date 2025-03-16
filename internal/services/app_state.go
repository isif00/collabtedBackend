package services

import (
	"context"
	"fmt"

	"github.com/CollabTED/CollabTed-Backend/prisma"
	db "github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type AppStateService struct{}

func NewAppStateService() *AppStateService {
	return &AppStateService{}
}

func (s *AppStateService) CreateAppState(userWorkspaceId string) (*db.AppStateModel, error) {
	appState, err := prisma.Client.AppState.CreateOne(
		db.AppState.UserWorkspaceID.Set(userWorkspaceId),
		db.AppState.UnreadChannels.Set([]string{}),
	).Exec(
		context.Background(),
	)
	return appState, err
}

func (s *AppStateService) GetAppState(userWorkspaceId string) (*db.AppStateModel, error) {
	appState, err := prisma.Client.AppState.FindFirst(
		db.AppState.UserWorkspaceID.Equals(userWorkspaceId),
	).Exec(context.Background())
	return appState, err
}

func (s *AppStateService) UpdateAppState(userWorkspaceId, action, value string) (*db.AppStateModel, error) {
	ctx := context.Background()

	switch action {
	case "append":
		existing, err := prisma.Client.AppState.FindFirst(
			db.AppState.UserWorkspaceID.Equals(userWorkspaceId),
		).Exec(ctx)
		if err != nil {
			return nil, err
		}

		uniqueChannels := make(map[string]struct{})

		if existing != nil {
			for _, ch := range existing.UnreadChannels {
				uniqueChannels[ch] = struct{}{}
			}
		}

		uniqueChannels[value] = struct{}{}

		newChannels := make([]string, 0, len(uniqueChannels))
		for ch := range uniqueChannels {
			newChannels = append(newChannels, ch)
		}

		appState, err := prisma.Client.AppState.FindUnique(
			db.AppState.UserWorkspaceID.Equals(userWorkspaceId),
		).Update(
			db.AppState.UnreadChannels.Set(newChannels),
		).Exec(ctx)
		return appState, err

	case "clear":
		existing, err := prisma.Client.AppState.FindUnique(
			db.AppState.UserWorkspaceID.Equals(userWorkspaceId),
		).Exec(ctx)
		if err != nil {
			return nil, err
		}

		filtered := make([]string, 0, len(existing.UnreadChannels))
		for _, ch := range existing.UnreadChannels {
			if ch != value {
				filtered = append(filtered, ch)
			}
		}

		updated, err := prisma.Client.AppState.FindUnique(
			db.AppState.UserWorkspaceID.Equals(userWorkspaceId),
		).Update(
			db.AppState.UnreadChannels.Set(filtered),
		).Exec(ctx)
		return updated, err

	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}
}
