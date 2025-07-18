package handlers

import (
	"net/http"

	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/labstack/echo/v4"
)

// TaskHandler struct with services as dependencies
type TaskHandler struct {
	TaskService    *services.TaskService
	ProjectService *services.ProjectService
}

// NewTaskHandler creates a new TaskHandler instance
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		TaskService:    services.NewTaskService(),
		ProjectService: services.NewProjectService(),
	}
}

// CreateTaskHandler handles task creation
func (h *TaskHandler) CreateTaskHandler(c echo.Context) error {
	var taskData types.TaskD

	// Parse input
	if err := c.Bind(&taskData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create the task
	task, err := h.TaskService.CreateTask(taskData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) ChangeTaskStatus(c echo.Context) error {
	taskId := c.Param("taskId")
	statusId := c.Param("statusId")
	task, err := h.TaskService.ChangeTaskStatus(taskId, statusId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, task)
}

// GetTaskByIdHandler retrieves a task by ID, ensuring the user is a member of the project.
func (h *TaskHandler) GetTaskByIdHandler(c echo.Context) error {
	taskID := c.Param("id")
	// claims := c.Get("user").(*types.Claims) // Assume userId is extracted from JWT middleware
	// workspaceId := c.QueryParam("workspaceId")

	// // Ensure the user is a member of the project
	// isMember, err := h.ProjectService.IsUserMemberOfProject(claims.ID, workspaceId, taskID)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	// }

	// if !isMember {
	// 	return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not a member of this project"})
	// }

	// Retrieve the task
	task, err := h.TaskService.GetTaskById(taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, task)
}

// ListTasksByProjectHandler lists tasks in a project, ensuring the user is a member of the project.
func (h *TaskHandler) ListTasksByProjectHandler(c echo.Context) error {
	projectID := c.Param("projectId")
	// claims := c.Get("user").(*types.Claims) // Assume userId is extracted from JWT middleware
	// workspaceId := c.Param("workspaceId")

	// // Ensure the user is a member of the project
	// isMember, err := h.ProjectService.IsUserMemberOfProject(claims.ID, workspaceId, projectID)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	// }

	// if !isMember {
	// 	return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not a member of this project"})
	// }

	// List tasks in the project
	tasks, err := h.TaskService.ListTasksByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) ListTasksCountByProjectHandler(c echo.Context) error {
	projectID := c.Param("projectId")
	// claims := c.Get("user").(*types.Claims) // Assume userId is extracted from JWT middleware
	// workspaceId := c.Param("workspaceId")

	// // Ensure the user is a member of the project
	// isMember, err := h.ProjectService.IsUserMemberOfProject(claims.ID, workspaceId, projectID)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	// }

	// if !isMember {
	// 	return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not a member of this project"})
	// }

	// List tasks in the project
	tasks, err := h.TaskService.ListTasksByProject(projectID)
	tasksCount := len(tasks)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tasksCount)
}

// Update Task Description
func (h *TaskHandler) UpdateDescription(c echo.Context) error {
	var taskId = c.Param("taskId")

	var request types.TaskD
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.TaskService.UpdateTask(request, taskId, "description")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) UpdateTaskTitle(c echo.Context) error {
	var taskId = c.Param("taskId")

	var request types.TaskD
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.TaskService.UpdateTask(request, taskId, "title")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) UpdateTaskPriority(c echo.Context) error {
	var taskId = c.Param("taskId")

	var request types.TaskD
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.TaskService.UpdateTask(request, taskId, "priority")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) UpdateTaskDeadline(c echo.Context) error {
	var taskId = c.Param("taskId")

	var request types.TaskD
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.TaskService.UpdateTask(request, taskId, "deadline")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// AddAssigneeToTaskHandler adds an assignee to a task
func (h *TaskHandler) AddAssigneeToTaskHandler(c echo.Context) error {
	var requestData struct {
		UserIDs []string `json:"userId" binding:"required"`
	}

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	taskID := c.Param("id")
	claims := c.Get("user").(*types.Claims) // Assume userId is extracted from JWT middleware
	workspaceId := c.QueryParam("workspaceId")

	// Check if user has permissions to add assignees (manager or project lead)
	canPerform, err := h.TaskService.CanUserPerformAction(claims.ID, workspaceId, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if !canPerform {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to add assignees to this task"})
	}

	// Add the assignee to the task
	userWorkspace, err := h.TaskService.AddAssignees(workspaceId, taskID, requestData.UserIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, userWorkspace)
}

func (h *TaskHandler) RemoveAssigneeToTaskHandler(c echo.Context) error {
	var requestData struct {
		UserIDs []string `json:"userIds" binding:"required"`
	}

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	taskID := c.Param("id")
	claims := c.Get("user").(*types.Claims) // Assume userId is extracted from JWT middleware
	workspaceId := c.QueryParam("workspaceId")

	// Check if user has permissions to add assignees (manager or project lead)
	canPerform, err := h.TaskService.CanUserPerformAction(claims.ID, workspaceId, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if !canPerform {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to add assignees to this task"})
	}

	// Add the assignee to the task
	userWorkspace, err := h.TaskService.RemoveAssignees(workspaceId, taskID, requestData.UserIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, userWorkspace)
}

func (h *TaskHandler) DeleteTaskHandler(c echo.Context) error {
	taskId := c.Param("taskId")
	err := h.TaskService.DeleteTask(taskId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, "Task deleted successfully")
}
