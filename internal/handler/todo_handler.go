package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/example/go-gin-jwt-auth/internal/middleware"
	"github.com/example/go-gin-jwt-auth/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const defaultTodoLimit = 50
const maxTodoLimit = 100

type TodoHandler struct {
	todoService service.TodoService
}

func NewTodoHandler(todoService service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

func (h *TodoHandler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := h.todoService.Create(userID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

func (h *TodoHandler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	todos, err := h.todoService.List(userID, parseTodoLimit(c.Query("limit")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todos": todos})
}

func (h *TodoHandler) Get(c *gin.Context) {
	userID, todoID, ok := idsFromRequest(c)
	if !ok {
		return
	}

	todo, err := h.todoService.Get(userID, todoID)
	if err != nil {
		writeTodoError(c, err)
		return
	}

	c.JSON(http.StatusOK, todo)
}

func (h *TodoHandler) Update(c *gin.Context) {
	userID, todoID, ok := idsFromRequest(c)
	if !ok {
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := h.todoService.Update(userID, todoID, req.Title, req.Description, req.Completed)
	if err != nil {
		writeTodoError(c, err)
		return
	}

	c.JSON(http.StatusOK, todo)
}

func (h *TodoHandler) Delete(c *gin.Context) {
	userID, todoID, ok := idsFromRequest(c)
	if !ok {
		return
	}

	if err := h.todoService.Delete(userID, todoID); err != nil {
		writeTodoError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
		return uuid.Nil, false
	}
	return userID, true
}

func idsFromRequest(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := userIDFromContext(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	todoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return uuid.Nil, uuid.Nil, false
	}

	return userID, todoID, true
}

func parseTodoLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultTodoLimit
	}
	if limit > maxTodoLimit {
		return maxTodoLimit
	}
	return limit
}

func writeTodoError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrTodoNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Todo request failed"})
}
