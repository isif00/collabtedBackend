package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
	prismaTypes "github.com/steebchen/prisma-client-go/runtime/types"
)

type BoardService struct{}

func NewBoardService() *BoardService {
	return &BoardService{}
}

func (s *BoardService) UpdateBoard(data types.BoardD, boardId string) (*db.BoardModel, error) {
	jsonElements, err := json.Marshal(data.Elements)
	if err != nil {
		log.Fatalf("Error marshaling elements: %v", err)
	}

	jsonAppState, err := json.Marshal(data.AppState)
	if err != nil {
		log.Fatalf("Error marshaling appState: %v", err)
	}

	jsonFiles, err := json.Marshal(data.Files)
	if err != nil {
		log.Fatalf("Error marshaling files: %v", err)
	}

	updatedBoard, err := prisma.Client.Board.FindUnique(
		db.Board.ID.Equals(boardId),
	).Update(
		db.Board.Elements.Set(jsonElements),
		db.Board.AppState.Set(jsonAppState),
		db.Board.Files.Set(jsonFiles),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return updatedBoard, nil
}

func (s *BoardService) SaveBoard(data types.BoardD) (*db.BoardModel, error) {

	jsonElements, err := json.Marshal(data.Elements)
	if err != nil {
		log.Fatalf("Error marshaling elements: %v", err)
	}

	var appState, files *db.JSON

	if data.AppState != nil {
		tmpAppState := db.JSON(data.AppState)
		appState = &tmpAppState
	}

	if data.Files != nil {
		tmpFiles := db.JSON(data.Files)
		files = &tmpFiles
	}

	result, err := prisma.Client.Board.CreateOne(
		db.Board.WorkspaceID.Set(data.WorkspaceID),
		db.Board.Elements.Set(prismaTypes.JSON(jsonElements)),
		db.Board.AppState.SetIfPresent(appState),
		db.Board.Files.SetIfPresent(files),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetBoard retrieves the board state for a given workspace
func (s *BoardService) GetBoard(workspaceId string) (*db.BoardModel, error) {
	board, err := prisma.Client.Board.FindFirst(
		db.Board.WorkspaceID.Equals(workspaceId),
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return board, nil
}
