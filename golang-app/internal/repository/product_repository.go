package repository

import (
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *entity.Product) error
	GetByID(id uint) (*entity.Product, error)
	GetAll(sellerID uint, categoryID uint, status string, page int, limit int) ([]entity.Product, int64, error)
	CreateImage(image *entity.ProductImage) error
	UpdateImageURL(imageID uint, imageURL string) error
	UpdateStatus(id uint, status string, reason string) error
	Update(product *entity.Product) error
	Delete(id uint) error
	DeleteImage(id uint) error
}

type productRepo struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(product *entity.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepo) GetByID(id uint) (*entity.Product, error) {
	var product entity.Product
	err := r.db.Preload("Images").Preload("Category").Preload("Seller").Preload("Auctions").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) GetAll(sellerID uint, categoryID uint, status string, page int, limit int) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64
	query := r.db.Model(&entity.Product{}).Preload("Images").Preload("Category").Preload("Seller").Preload("Auctions")
	if sellerID != 0 {
		query = query.Where("seller_id = ?", sellerID)
	}
	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&products).Error
	return products, total, err
}

func (r *productRepo) CreateImage(image *entity.ProductImage) error {
	return r.db.Create(image).Error
}

func (r *productRepo) UpdateImageURL(imageID uint, imageURL string) error {
	// Use UpdateColumn to avoid updating updated_at if it exists, or triggering hooks
	return r.db.Model(&entity.ProductImage{}).Where("id = ?", imageID).UpdateColumn("image_url", imageURL).Error
}

func (r *productRepo) UpdateStatus(id uint, status string, reason string) error {
	return r.db.Model(&entity.Product{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           status,
		"rejection_reason": reason,
	}).Error
}

func (r *productRepo) Update(product *entity.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&entity.Product{}, id).Error
}

func (r *productRepo) DeleteImage(id uint) error {
	return r.db.Delete(&entity.ProductImage{}, id).Error
}
