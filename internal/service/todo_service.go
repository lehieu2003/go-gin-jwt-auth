package service

import (
	"errors"

	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/example/go-gin-jwt-auth/internal/repository"
	"github.com/google/uuid"
)

var ErrTodoNotFound = errors.New("todo not found")

type TodoService interface {
	Create(userID uuid.UUID, title, description string) (*model.Todo, error)
	List(userID uuid.UUID, limit int) ([]model.Todo, error)
	Get(userID, id uuid.UUID) (*model.Todo, error)
	Update(userID, id uuid.UUID, title, description *string, completed *bool) (*model.Todo, error)
	Delete(userID, id uuid.UUID) error
}

type todoService struct {
	todoRepo repository.TodoRepository
}

func NewTodoService(todoRepo repository.TodoRepository) TodoService {
	return &todoService{todoRepo: todoRepo}
}

func (s *todoService) Create(userID uuid.UUID, title, description string) (*model.Todo, error) {
	todo := &model.Todo{
		UserID:      userID,
		Title:       title,
		Description: description,
	}
	if err := s.todoRepo.Create(todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func (s *todoService) List(userID uuid.UUID, limit int) ([]model.Todo, error) {
	return s.todoRepo.FindByUserID(userID, limit)
}

func (s *todoService) Get(userID, id uuid.UUID) (*model.Todo, error) {
	todo, err := s.todoRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, err
	}
	if todo == nil {
		return nil, ErrTodoNotFound
	}
	return todo, nil
}

func (s *todoService) Update(userID, id uuid.UUID, title, description *string, completed *bool) (*model.Todo, error) {
	todo, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if title != nil {
		todo.Title = *title
	}
	if description != nil {
		todo.Description = *description
	}
	if completed != nil {
		todo.Completed = *completed
	}
	return todo, s.todoRepo.Update(todo)
}

func (s *todoService) Delete(userID, id uuid.UUID) error {
	todo, err := s.Get(userID, id)
	if err != nil {
		return err
	}
	return s.todoRepo.Delete(todo.ID, userID)
}
