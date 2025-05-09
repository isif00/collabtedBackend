package services

import (
	"context"
	"errors"

	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type StatusService struct{}

func NewStatusService() *StatusService {
	return &StatusService{}
}

func (s *StatusService) CreateStatus(data types.StatusD) (*db.StatusModel, error) {
	// Create a new status
	result, err := prisma.Client.Status.CreateOne(
		db.Status.Project.Link(
			db.Project.ID.Equals(data.ProjectID),
		),
		db.Status.Title.Set(data.Name),
		db.Status.Color.Set(data.Color),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *StatusService) EditStatus(statusId, userId string, data types.StatusD) (*db.StatusModel, error) {
	result, err := prisma.Client.Status.FindUnique(
		db.Status.ID.Equals(statusId),
	).Update(
		db.Status.Title.Set(data.Name),
		db.Status.Color.Set(data.Color),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return result, err
}

func (s *StatusService) GetStatusesByProject(projectID string, userID string) ([]db.StatusModel, error) {
	// // Check if the user is an assignee of the project
	// assignee, err := s.isAssigneeOfProject(projectID, userID)
	// if err != nil {
	// 	return nil, err
	// }
	// if !assignee {
	// 	return nil, errors.New("only assignees of the project can view statuses")
	// }

	// Get all statuses for the project
	statuses, err := prisma.Client.Status.FindMany(
		db.Status.ProjectID.Equals(projectID),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

func (s *StatusService) GetStatusByID(statusID string, userID string) (*db.StatusModel, error) {
	// Get the project ID of the status
	projectID, err := s.getProjectIDOfStatus(statusID)
	if err != nil {
		return nil, err
	}

	// Check if the user is an assignee of the project
	assignee, err := s.isAssigneeOfProject(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !assignee {
		return nil, errors.New("only assignees of the project can view statuses")
	}

	// Get the status by ID
	status, err := prisma.Client.Status.FindFirst(
		db.Status.ID.Equals(statusID),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (s *StatusService) DeleteStatus(statusID string, userID string) error {
	// Get the project ID of the status
	_, err := s.getProjectIDOfStatus(statusID)
	if err != nil {
		return err
	}

	// Delete the status by ID
	_, err = prisma.Client.Status.FindUnique(
		db.Status.ID.Equals(statusID),
	).Delete().Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// func (s *StatusService) isLeadOfProject(projectID string, userID string) (bool, error) {
// 	// Get the project by ID
// 	project, err := prisma.Client.Project.FindFirst(
// 		db.Project.ID.Equals(projectID),
// 	).Exec(context.Background())
// 	if err != nil {
// 		return false, err
// 	}
// 	//get userworkspaceID
// 	userwrk, err := prisma.Client.UserWorkspace.FindMany(
// 		db.UserWorkspace.WorkspaceID.Equals(project.WorkspaceID),
// 		db.UserWorkspace.UserID.Equals(userID),
// 	).Exec(context.Background())
// 	if err != nil {
// 		return false, err
// 	}
// 	// Check if the user is the lead of the project
// 	return project.LeadID == userwrk[0].ID, nil
// }

func (s *StatusService) isAssigneeOfProject(projectID string, userID string) (bool, error) {
	// Get the project by ID
	project, err := prisma.Client.Project.FindUnique(
		db.Project.ID.Equals(projectID),
	).With(
		db.Project.Assignees.Fetch(),
	).Exec(context.Background())
	if err != nil {
		return false, err
	}

	// Check if the user is in the list of assignees
	assignees := project.Assignees()
	for _, assignee := range assignees {
		if assignee.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *StatusService) getProjectIDOfStatus(statusID string) (string, error) {
	// Get the status by ID
	status, err := prisma.Client.Status.FindFirst(
		db.Status.ID.Equals(statusID),
	).Exec(context.Background())
	if err != nil {
		return "", err
	}

	// Get the project ID of the status
	return status.ProjectID, nil
}
