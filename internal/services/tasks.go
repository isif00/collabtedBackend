package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

// TaskService handles the task operations
type TaskService struct{}

// NewTaskService creates a new TaskService instance
func NewTaskService() *TaskService {
	return &TaskService{}
}

// CreateTask creates a new task in a project and assigns assignees.
func (s *TaskService) CreateTask(data types.TaskD) (*db.TaskModel, error) {
	logger.LogDebug().Msg("Creating task..." + data.ProjectID)
	//Marshal description
	jsonDes, err := json.Marshal(data.Description)
	if err != nil {
		return nil, err
	}
	// Create a new task
	result, err := prisma.Client.Task.CreateOne(
		db.Task.Project.Link(
			db.Project.ID.Equals(data.ProjectID),
		),
		db.Task.Title.Set(data.Title),
		db.Task.Description.Set(jsonDes),
		db.Task.DueDate.Set(data.DueDate),
		db.Task.Priority.Set(db.Priority(data.Priority)),
		db.Task.Status.Link(
			db.Status.ID.Equals(data.StatusID),
		),
	).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	// Link assignees to the task
	for _, assigneeID := range data.AssigneesIDs {
		_, err = prisma.Client.Task.FindUnique(
			db.Task.ID.Equals(result.ID),
		).Update(db.Task.Assignees.Link(
			db.UserWorkspace.ID.Equals(assigneeID),
		)).Exec(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to add assignee with ID %s to the task: %v", assigneeID, err)
		}
	}

	return result, nil
}

// GetTaskById retrieves a task by its ID.
func (s *TaskService) GetTaskById(taskID string) (*db.TaskModel, error) {
	task, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(taskID),
	).With(db.Task.Assignees.Fetch(), db.Task.Project.Fetch(), db.Task.Status.Fetch()).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) UpdateTask(data types.TaskD, taskId string, field string) (*db.TaskModel, error) {

	fieldUpdaters := map[string]func(types.TaskD) (db.TaskSetParam, error){
		"description": func(data types.TaskD) (db.TaskSetParam, error) {
			jsonElements, err := json.Marshal(data.Description)
			if err != nil {
				return nil, fmt.Errorf("error marshaling description: %w", err)
			}
			return db.Task.Description.Set(jsonElements), nil
		},
		"title": func(data types.TaskD) (db.TaskSetParam, error) {
			return db.Task.Title.Set(data.Title), nil
		},
		"priority": func(data types.TaskD) (db.TaskSetParam, error) {
			return db.Task.Priority.Set(db.Priority(data.Priority)), nil
		},
		"deadline": func(data types.TaskD) (db.TaskSetParam, error) {
			return db.Task.DueDate.Set(data.DueDate), nil
		},
	}

	updater, exists := fieldUpdaters[field]
	if !exists {
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	updateParam, err := updater(data)
	if err != nil {
		return nil, err
	}

	updatedTask, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(taskId),
	).Update(
		updateParam,
	).Exec(context.Background())

	if err != nil {
		log.Printf("Error updating task: %v", err)
		return nil, err
	}

	return updatedTask, nil
}

// ListTasksByProject lists all tasks in a project.
func (s *TaskService) ListTasksByProject(projectID string) ([]db.TaskModel, error) {
	tasks, err := prisma.Client.Task.FindMany(
		db.Task.ProjectID.Equals(projectID),
	).With(db.Task.Assignees.Fetch(), db.Task.Status.Fetch()).Exec(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return tasks, nil
}

// AddAssignee adds a user to a task as an assignee.
func (s *TaskService) AddAssignees(workspaceID, taskID string, userID []string) ([]db.UserWorkspaceModel, error) {
	ctx := context.Background()
	users, err := prisma.Client.UserWorkspace.FindMany(
		db.UserWorkspace.UserID.In(userID),
		db.UserWorkspace.WorkspaceID.Equals(workspaceID),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		_, err = prisma.Client.Task.FindUnique(
			db.Task.ID.Equals(taskID),
		).Update(
			db.Task.Assignees.Link(db.UserWorkspace.ID.Equals(user.ID)),
		).Exec(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	return users, nil
}

func (s *TaskService) RemoveAssignees(workspaceID, taskID string, userID []string) ([]db.UserWorkspaceModel, error) {
	fmt.Println(userID)
	ctx := context.Background()
	users, err := prisma.Client.UserWorkspace.FindMany(
		db.UserWorkspace.UserID.In(userID),
		db.UserWorkspace.WorkspaceID.Equals(workspaceID),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(len(users))
	var usersIds []string

	for _, user := range users {
		_, err = prisma.Client.Task.FindUnique(
			db.Task.ID.Equals(taskID),
		).Update(
			db.Task.Assignees.Unlink(db.UserWorkspace.ID.Equals(user.ID)),
		).Exec(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	fmt.Println(usersIds)

	return users, nil
}

// CanUserPerformAction checks if a user has the required role to perform an action on a task.
func (s *TaskService) CanUserPerformAction(userId, workspaceId, taskId string) (bool, error) {
	// Find the UserWorkspace entry for the user in the workspace
	userWorkspace, err := prisma.Client.UserWorkspace.FindFirst(
		db.UserWorkspace.UserID.Equals(userId),
		db.UserWorkspace.WorkspaceID.Equals(workspaceId),
	).Exec(context.Background())

	if err != nil {
		return false, err
	}

	if userWorkspace == nil {
		return false, fmt.Errorf("user is not part of the workspace")
	}

	// Check if the user is a manager
	if userWorkspace.Role == db.UserRoleManager {
		return true, nil
	}

	// Fetch the task and check if the user is the lead of the associated project
	task, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(taskId),
	).With(db.Task.Project.Fetch()).Exec(context.Background())

	if err != nil {
		return false, err
	}

	// Check if the user is the lead of the project associated with the task
	if task.Project().LeadID == userWorkspace.ID {
		return true, nil
	}

	return false, nil
}

func (s *TaskService) ChangeTaskStatus(taskId, statusId string) (*db.TaskModel, error) {
	status, err := prisma.Client.Status.FindUnique(
		db.Status.ID.Equals(statusId),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	task, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(taskId),
	).Update(
		db.Task.Status.Link(
			db.Status.ID.Equals(status.ID),
		),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return task, nil

}

// AssignUserToTask assigns a single user to a task using the userWorkspaceID.
func (s *TaskService) AssignUserToTask(taskID, userWorkspaceID string) (*db.TaskModel, error) {
	ctx := context.Background()

	// Find the task and link the user as an assignee using userWorkspaceID
	task, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(taskID),
	).Update(
		db.Task.Assignees.Link(db.UserWorkspace.ID.Equals(userWorkspaceID)),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to assign userWorkspaceID %s to task: %v", userWorkspaceID, err)
	}

	return task, nil
}

func (s *ProjectService) IsUserMemberOfProject(userId, workspaceId, projectId string) (bool, error) {
	// Check if the user is part of the workspace and project
	fmt.Println(userId, workspaceId, projectId)
	userWorkspace, err := prisma.Client.UserWorkspace.FindFirst(
		db.UserWorkspace.UserID.Equals(userId),
		db.UserWorkspace.WorkspaceID.Equals(workspaceId),
	).With(db.UserWorkspace.Projects.Fetch()).Exec(context.Background())

	if err != nil {

		fmt.Println(err)
		return false, err
	}

	if userWorkspace == nil {
		return false, nil
	}

	// Check if the user is part of the specific project
	for _, project := range userWorkspace.Projects() {
		if project.ID == projectId {
			return true, nil
		}
	}

	return false, nil
}

func (s *TaskService) DeleteTask(projectId string) error {
	ctx := context.Background()
	_, err := prisma.Client.Task.FindUnique(
		db.Task.ID.Equals(projectId),
	).Delete().Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
