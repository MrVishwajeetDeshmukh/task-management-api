package handler

import (
	"strconv"

	"task-management-api/internal/domain"
	"task-management-api/internal/service"
	"task-management-api/internal/util"

	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	taskService   service.TaskService
	workerService service.WorkerService
}

func NewTaskHandler(taskService service.TaskService, workerService service.WorkerService) *TaskHandler {
	return &TaskHandler{
		taskService:   taskService,
		workerService: workerService,
	}
}

func (h *TaskHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := util.ValidateTaskTitle(req.Title); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	userID := c.Locals("userID").(string)

	task, err := h.taskService.Create(req, userID)
	if err != nil {
		return util.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	// Enqueue task for auto-completion
	h.workerService.EnqueueTask(task.ID)

	return util.SendSuccess(c, fiber.StatusCreated, task)
}

func (h *TaskHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userRole := c.Locals("userRole").(string)
	isAdmin := userRole == string(domain.RoleAdmin)

	filter := domain.TaskFilter{
		Limit:  50,
		Offset: 0,
	}

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		taskStatus := domain.TaskStatus(status)
		if !taskStatus.IsValid() {
			return util.SendError(c, fiber.StatusBadRequest, "invalid status parameter")
		}
		filter.Status = &taskStatus
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	tasks, err := h.taskService.List(filter, userID, isAdmin)
	if err != nil {
		return util.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return util.SendSuccess(c, fiber.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)
	userRole := c.Locals("userRole").(string)
	isAdmin := userRole == string(domain.RoleAdmin)

	task, err := h.taskService.GetByID(id, userID, isAdmin)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "task not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized access" {
			status = fiber.StatusForbidden
		}
		return util.SendError(c, status, err.Error())
	}

	return util.SendSuccess(c, fiber.StatusOK, task)
}

func (h *TaskHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)
	userRole := c.Locals("userRole").(string)
	isAdmin := userRole == string(domain.RoleAdmin)

	var req domain.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Title != nil {
		if err := util.ValidateTaskTitle(*req.Title); err != nil {
			return util.SendError(c, fiber.StatusBadRequest, err.Error())
		}
	}

	task, err := h.taskService.Update(id, req, userID, isAdmin)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "task not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized access" {
			status = fiber.StatusForbidden
		} else if err.Error() == "invalid status" {
			status = fiber.StatusBadRequest
		}
		return util.SendError(c, status, err.Error())
	}

	return util.SendSuccess(c, fiber.StatusOK, task)
}

func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)
	userRole := c.Locals("userRole").(string)
	isAdmin := userRole == string(domain.RoleAdmin)

	err := h.taskService.Delete(id, userID, isAdmin)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "task not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized access" {
			status = fiber.StatusForbidden
		}
		return util.SendError(c, status, err.Error())
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
