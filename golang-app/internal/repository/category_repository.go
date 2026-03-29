package repository

import (
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *entity.Category) error
	GetByID(id uint) (*entity.Category, error)
	GetAll(status string) ([]entity.Category, error)
	GetByUserID(userID uint) ([]entity.Category, error)
	UpdateStatus(id uint, status string, rejectionReason string) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(category *entity.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepo) GetByID(id uint) (*entity.Category, error) {
	var category entity.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepo) GetAll(status string) ([]entity.Category, error) {
	var categories []entity.Category
	query := r.db.Model(&entity.Category{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) GetByUserID(userID uint) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Where("created_by = ?", userID).Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) UpdateStatus(id uint, status string, rejectionReason string) error {
	return r.db.Model(&entity.Category{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           status,
		"rejection_reason": rejectionReason,
	}).Error
}
