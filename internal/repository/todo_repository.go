package repository

import (
	"errors"

	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(todo *model.Todo) error
	FindByUserID(userID uuid.UUID, limit int) ([]model.Todo, error)
	FindByIDAndUserID(id, userID uuid.UUID) (*model.Todo, error)
	Update(todo *model.Todo) error
	Delete(id, userID uuid.UUID) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) FindByUserID(userID uuid.UUID, limit int) ([]model.Todo, error) {
	var todos []model.Todo
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&todos).Error
	return todos, err
}

func (r *todoRepository) FindByIDAndUserID(id, userID uuid.UUID) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&todo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) Update(todo *model.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Todo{}).Error
}
