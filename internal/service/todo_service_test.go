package service

import (
	"testing"
	"time"

	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/google/uuid"
)

// In-memory mock for TodoRepository
type mockTodoRepository struct {
	todos map[uuid.UUID]*model.Todo
}

func newMockTodoRepository() *mockTodoRepository {
	return &mockTodoRepository{
		todos: make(map[uuid.UUID]*model.Todo),
	}
}

func (m *mockTodoRepository) Create(todo *model.Todo) error {
	if todo.ID == uuid.Nil {
		todo.ID = uuid.New()
	}
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepository) FindByUserID(userID uuid.UUID, limit int) ([]model.Todo, error) {
	var result []model.Todo
	for _, t := range m.todos {
		if t.UserID == userID {
			result = append(result, *t)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockTodoRepository) FindByIDAndUserID(id, userID uuid.UUID) (*model.Todo, error) {
	if todo, ok := m.todos[id]; ok {
		if todo.UserID == userID {
			return todo, nil
		}
	}
	return nil, nil
}

func (m *mockTodoRepository) Update(todo *model.Todo) error {
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepository) Delete(id, userID uuid.UUID) error {
	if todo, ok := m.todos[id]; ok {
		if todo.UserID == userID {
			delete(m.todos, id)
		}
	}
	return nil
}

func TestTodoService_CreateAndGet(t *testing.T) {
	repo := newMockTodoRepository()
	todoSvc := NewTodoService(repo)

	user1 := uuid.New()
	user2 := uuid.New()

	// 1. Create Todo for User 1
	todo, err := todoSvc.Create(user1, "Buy milk", "2 cartons of fresh milk")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if todo.ID == uuid.Nil || todo.Title != "Buy milk" {
		t.Fatalf("unexpected todo created: %+v", todo)
	}

	// 2. User 1 gets own Todo
	fetched, err := todoSvc.Get(user1, todo.ID)
	if err != nil {
		t.Fatalf("Get failed for owner: %v", err)
	}
	if fetched.ID != todo.ID {
		t.Errorf("expected ID %v, got %v", todo.ID, fetched.ID)
	}

	// 3. User 2 attempts to get User 1's Todo (Access control / Isolation)
	_, err = todoSvc.Get(user2, todo.ID)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound when accessed by non-owner, got %v", err)
	}

	// 4. Get non-existent Todo ID
	_, err = todoSvc.Get(user1, uuid.New())
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound for non-existent ID, got %v", err)
	}
}

func TestTodoService_List(t *testing.T) {
	repo := newMockTodoRepository()
	todoSvc := NewTodoService(repo)

	user1 := uuid.New()
	user2 := uuid.New()

	// Create 3 todos for user1, 1 for user2
	_, _ = todoSvc.Create(user1, "Task 1", "Desc 1")
	_, _ = todoSvc.Create(user1, "Task 2", "Desc 2")
	_, _ = todoSvc.Create(user1, "Task 3", "Desc 3")
	_, _ = todoSvc.Create(user2, "User 2 Task", "Desc")

	// List user1 todos with limit 2
	todos, err := todoSvc.List(user1, 2)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("expected 2 todos due to limit, got %d", len(todos))
	}

	// List user1 todos with limit 10
	allUser1Todos, err := todoSvc.List(user1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(allUser1Todos) != 3 {
		t.Errorf("expected 3 todos for user1, got %d", len(allUser1Todos))
	}

	// List user2 todos
	user2Todos, err := todoSvc.List(user2, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(user2Todos) != 1 {
		t.Errorf("expected 1 todo for user2, got %d", len(user2Todos))
	}
}

func TestTodoService_Update(t *testing.T) {
	repo := newMockTodoRepository()
	todoSvc := NewTodoService(repo)

	user1 := uuid.New()
	user2 := uuid.New()

	todo, _ := todoSvc.Create(user1, "Original Title", "Original Desc")

	// 1. User 1 updates title and completed status
	newTitle := "Updated Title"
	completed := true
	updated, err := todoSvc.Update(user1, todo.ID, &newTitle, nil, &completed)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected updated title %q, got %q", "Updated Title", updated.Title)
	}
	if !updated.Completed {
		t.Error("expected completed to be true")
	}
	if updated.Description != "Original Desc" {
		t.Errorf("expected description to remain untouched, got %q", updated.Description)
	}

	// 2. User 2 attempts to update User 1's todo
	hackedTitle := "Hacked Title"
	_, err = todoSvc.Update(user2, todo.ID, &hackedTitle, nil, nil)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound on cross-user update, got %v", err)
	}
}

func TestTodoService_Delete(t *testing.T) {
	repo := newMockTodoRepository()
	todoSvc := NewTodoService(repo)

	user1 := uuid.New()
	user2 := uuid.New()

	todo, _ := todoSvc.Create(user1, "Delete Me", "Desc")

	// 1. User 2 cannot delete User 1's todo
	err := todoSvc.Delete(user2, todo.ID)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound when non-owner deletes, got %v", err)
	}

	// Ensure todo still exists
	_, err = todoSvc.Get(user1, todo.ID)
	if err != nil {
		t.Errorf("todo should still exist after unauthorized delete attempt: %v", err)
	}

	// 2. User 1 deletes own todo
	err = todoSvc.Delete(user1, todo.ID)
	if err != nil {
		t.Fatalf("Delete failed for owner: %v", err)
	}

	// Ensure todo is gone
	_, err = todoSvc.Get(user1, todo.ID)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound after deletion, got %v", err)
	}
}
