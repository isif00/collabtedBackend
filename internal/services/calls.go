package services

import (
	"github.com/CollabTED/CollabTed-Backend/config"
	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
)

type CallService struct{}

func NewCallService() *CallService {
	return &CallService{}
}

func (s *CallService) GetGlobalJoinToken(participantName string, workspaceId string) (string, string, error) {
	API_KEY := config.LIVEKIT_API_KEY
	API_SECRET := config.LIVEKIT_API_SECRET

	at := auth.NewAccessToken(API_KEY, API_SECRET)

	var roomId string = "global:" + workspaceId
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomId,
	}

	at.AddGrant(grant).SetIdentity(participantName)

	globalRoomJoinToken, err := at.ToJWT()
	if err != nil {
		return "", "", err
	}

	logger.LogDebug().Msg(globalRoomJoinToken)

	return globalRoomJoinToken, roomId, nil
}

func (s *CallService) GetPrivateJoinToken(participantName string, workspaceId string) (string, string, error) {
	API_KEY := config.LIVEKIT_API_KEY
	API_SECRET := config.LIVEKIT_API_SECRET

	at := auth.NewAccessToken(API_KEY, API_SECRET)

	var roomId string = uuid.NewString() + workspaceId
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomId,
	}

	at.AddGrant(grant).SetIdentity(participantName)

	privateRoomJoinToken, err := at.ToJWT()
	if err != nil {
		return "", "", err
	}

	return privateRoomJoinToken, roomId, nil
}

func (s *CallService) JoinRoomToken(roomId string, participantName string) (string, error) {
	API_KEY := config.LIVEKIT_API_KEY
	API_SECRET := config.LIVEKIT_API_SECRET

	at := auth.NewAccessToken(API_KEY, API_SECRET)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomId,
	}

	at.AddGrant(grant).SetIdentity(participantName)

	privateRoomJoinToken, err := at.ToJWT()
	if err != nil {
		return "", err
	}

	return privateRoomJoinToken, nil
}
