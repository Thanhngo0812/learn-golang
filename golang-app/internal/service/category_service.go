package service

import (
	"golang-app/internal/model/dto"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"
	"strings"
)

type CategoryService interface {
	CreateCategory(req dto.CreateCategoryRequest, userID uint, role string) (*entity.Category, error)
	GetAllCategories(status string) ([]entity.Category, error)
	GetMyCategories(userID uint) ([]entity.Category, error)
	ApproveCategory(id uint, status string, reason string) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(req dto.CreateCategoryRequest, userID uint, role string) (*entity.Category, error) {
	status := "pending"
	if role == "admin" {
		status = "active"
	}

	category := &entity.Category{
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		CreatedBy:   &userID,
	}

	err := s.repo.Create(category)
	return category, err
}

func (s *categoryService) GetAllCategories(status string) ([]entity.Category, error) {
	return s.repo.GetAll(status)
}

func (s *categoryService) GetMyCategories(userID uint) ([]entity.Category, error) {
	return s.repo.GetByUserID(userID)
}

func (s *categoryService) ApproveCategory(id uint, status string, reason string) error {
	normalized := strings.ToLower(strings.TrimSpace(status))

	// Keep backward compatibility for clients sending product-style statuses.
	if normalized == "approve" || normalized == "approved" {
		normalized = "active"
	}
	if normalized == "reject" {
		normalized = "rejected"
	}

	if normalized != "active" && normalized != "rejected" && normalized != "pending" {
		return apperror.NewBadRequest(nil, "Trạng thái category không hợp lệ. Chỉ chấp nhận: active, rejected, pending")
	}

	if normalized != "rejected" {
		reason = ""
	}

	return s.repo.UpdateStatus(id, normalized, reason)
}
