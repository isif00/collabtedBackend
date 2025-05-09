package services

import (
	"context"

	"fmt"

	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type ChannelService struct{}

func NewChannelService() *ChannelService {
	return &ChannelService{}
}

// CreateChannel creates a new channel in a workspace and adds existing participants.
func (s *ChannelService) CreateChannel(data types.ChannelD) (*db.ChannelModel, error) {
	// Create a new channel
	result, err := prisma.Client.Channel.CreateOne(
		db.Channel.Name.Set(data.Name),
		db.Channel.CreatorID.Set(data.CreatorID),
		db.Channel.Workspace.Link(
			db.Workspace.ID.Equals(data.WorkspaceID),
		),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}

	usr, err := prisma.Client.UserWorkspace.FindUnique(
		db.UserWorkspace.ID.Equals(data.CreatorID),
	).Exec(context.Background())
	if err != nil {
		fmt.Println("user workspace not found")
		return nil, err
	}
	_, err = prisma.Client.Channel.FindUnique(
		db.Channel.ID.Equals(result.ID),
	).Update(
		db.Channel.Participants.Link(
			db.UserWorkspace.ID.Equals(usr.ID),
		),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetChannelById retrieves a channel by its ID.
func (s *ChannelService) GetChannelById(channelID string) (*db.ChannelModel, error) {
	channel, err := prisma.Client.Channel.FindUnique(
		db.Channel.ID.Equals(channelID),
	).With(db.Channel.Participants.Fetch()).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (s *ChannelService) DeleteChannel(channelID string) (*db.ChannelModel, error) {
	channel, err := prisma.Client.Channel.FindUnique(
		db.Channel.ID.Equals(channelID),
	).Delete().Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (s *ChannelService) EditChannelName(channelID, name string) (*db.ChannelModel, error) {
	channel, err := prisma.Client.Channel.FindUnique(
		db.Channel.ID.Equals(channelID),
	).Update(
		db.Channel.Name.Set(name),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}

	return channel, nil
}

// ListChannelsByWorkspace lists all channels in a workspace.
func (s *ChannelService) ListChannelsByWorkspace(workspaceID string) ([]db.ChannelModel, error) {
	channels, err := prisma.Client.Channel.FindMany(
		db.Channel.WorkspaceID.Equals(workspaceID),
	).With(db.Channel.Participants.Fetch()).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (s *ChannelService) AddParticipants(workspaceID string, channelID string, userIDs []string) ([]*db.UserWorkspaceModel, error) {
	ctx := context.Background()
	var addedUsers []*db.UserWorkspaceModel

	for _, userID := range userIDs {
		// Ensure userID is already a valid ObjectID as string
		user, err := prisma.Client.UserWorkspace.FindFirst(
			db.UserWorkspace.UserID.Equals(userID), // Expect userID to be a valid string
			db.UserWorkspace.WorkspaceID.Equals(workspaceID),
		).Exec(ctx)
		if err != nil {
			return nil, err
		}

		// Add the user to the channel's participants
		c, err := prisma.Client.Channel.FindUnique(
			db.Channel.ID.Equals(channelID),
		).With(db.Channel.Participants.Fetch()).Update(
			db.Channel.Participants.Link(db.UserWorkspace.ID.Equals(user.ID)),
		).Exec(ctx)
		if err != nil {
			return nil, err
		}

		fmt.Println("added", user.ID)
		fmt.Println("channel", channelID)

		fmt.Println(c.Participants()[0].JoinedAt)
		fmt.Println(c.ID)

		logger.LogDebug().Msg("Added user to channel")
		addedUsers = append(addedUsers, user)
	}

	return addedUsers, nil
}
