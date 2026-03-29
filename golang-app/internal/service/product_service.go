package service

import (
	"io"
	"log"
	"mime/multipart"
	"golang-app/internal/model/dto"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/internal/worker"
	"golang-app/pkg/apperror"
)

type ProductService interface {
	CreateProduct(req dto.CreateProductRequest, sellerID uint) (*entity.Product, error)
	GetProductByID(id uint) (*entity.Product, error)
	GetAllProducts(sellerID uint, categoryID uint, status string, page int, limit int) ([]entity.Product, int64, error)
	ApproveProduct(id uint, status string, reason string) error
	UpdateProduct(id uint, req dto.CreateProductRequest, userID uint, isAdmin bool) error
	DeleteProduct(id uint, userID uint, isAdmin bool) error
	ToggleBanned(id uint, isAdmin bool, reason string) error
	DeleteProductImage(productID uint, imageID uint, userID uint, isAdmin bool) error
}

type productService struct {
	repo         repository.ProductRepository
	auctionRepo  repository.AuctionRepository
	uploadWorker *worker.UploadWorker
}

func NewProductService(repo repository.ProductRepository, auctionRepo repository.AuctionRepository, uploadWorker *worker.UploadWorker) ProductService {
	return &productService{
		repo:         repo,
		auctionRepo:  auctionRepo,
		uploadWorker: uploadWorker,
	}
}

func (s *productService) CreateProduct(req dto.CreateProductRequest, sellerID uint) (*entity.Product, error) {
	product := &entity.Product{
		SellerID:    sellerID,
		CategoryID:  &req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending", // Chờ duyệt? Tùy logic, ở đây giả sử pending
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	// Xử lý ảnh
	log.Printf("Service: Processing %d images for product %d", len(req.Images), product.ID)
	for i, fileHeader := range req.Images {
		func(index int, header *multipart.FileHeader) {
			file, err := header.Open()
			if err != nil {
				log.Printf("Service error: failed to open image %d: %v", index, err)
				return
			}
			defer file.Close()

			fileData, err := io.ReadAll(file)
			if err != nil {
				log.Printf("Service error: failed to read image data %d: %v", index, err)
				return
			}

			// Tạo bản ghi ảnh trước (chưa có URL)
			img := &entity.ProductImage{
				ProductID:    product.ID,
				ImageURL:     "", // Sẽ được cập nhật từ worker
				IsPrimary:    index == 0,
				DisplayOrder: index,
			}

			if err := s.repo.CreateImage(img); err != nil {
				log.Printf("Service error: failed to create image record %d: %v", index, err)
				return
			}

			log.Printf("Service: Created image record ID %d, FileData length: %d bytes. Sending to worker", img.ID, len(fileData))
			// Gửi task vào worker
			s.uploadWorker.AddTask(worker.UploadTask{
				ProductID: product.ID,
				ImageID:   img.ID,
				FileData:  fileData,
				Folder:    "products",
			})
		}(i, fileHeader)
	}

	return product, nil
}

func (s *productService) GetProductByID(id uint) (*entity.Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) GetAllProducts(sellerID uint, categoryID uint, status string, page int, limit int) ([]entity.Product, int64, error) {
	return s.repo.GetAll(sellerID, categoryID, status, page, limit)
}

func (s *productService) ApproveProduct(id uint, status string, reason string) error {
	return s.repo.UpdateStatus(id, status, reason)
}

func (s *productService) UpdateProduct(id uint, req dto.CreateProductRequest, userID uint, isAdmin bool) error {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !isAdmin && product.SellerID != userID {
		return apperror.NewForbidden(nil, "Bạn không có quyền sửa sản phẩm này")
	}

	// Check for active auction
	hasActive, err := s.auctionRepo.CheckActiveAuction(id)
	if err != nil {
		return err
	}
	if hasActive {
		return apperror.NewBadRequest(nil, "Không thể sửa sản phẩm khi đang có phiên đấu giá hoạt động")
	}

	product.CategoryID = &req.CategoryID
	product.Name = req.Name
	product.Description = req.Description
	if !isAdmin {
		product.Status = "pending"
	}

	if err := s.repo.Update(product); err != nil {
		return err
	}

	// Handle new images if any
	if len(req.Images) > 0 {
		log.Printf("Service: Processing %d new images for product %d", len(req.Images), product.ID)
		for i, fileHeader := range req.Images {
			func(index int, header *multipart.FileHeader) {
				file, err := header.Open()
				if err != nil {
					log.Printf("Service error: failed to open image: %v", err)
					return
				}
				defer file.Close()

				fileData, err := io.ReadAll(file)
				if err != nil {
					log.Printf("Service error: failed to read image data: %v", err)
					return
				}

				img := &entity.ProductImage{
					ProductID:    product.ID,
					ImageURL:     "",
					IsPrimary:    false, // Default to false for updates
					DisplayOrder: 100 + index, // High order to put them at the end
				}

				if err := s.repo.CreateImage(img); err != nil {
					log.Printf("Service error: failed to create image record: %v", err)
					return
				}

				s.uploadWorker.AddTask(worker.UploadTask{
					ProductID: product.ID,
					ImageID:   img.ID,
					FileData:  fileData,
					Folder:    "products",
				})
			}(i, fileHeader)
		}
	}

	return nil
}

func (s *productService) DeleteProduct(id uint, userID uint, isAdmin bool) error {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !isAdmin && product.SellerID != userID {
		return apperror.NewForbidden(nil, "Bạn không có quyền xóa sản phẩm này")
	}

	// Logic: can only delete if NO auctions ever created
	if len(product.Auctions) > 0 {
		return apperror.NewBadRequest(nil, "Không thể xóa sản phẩm đã từng tham gia đấu giá. Hãy sử dụng chức năng Khóa.")
	}

	return s.repo.Delete(id)
}

func (s *productService) ToggleBanned(id uint, isAdmin bool, reason string) error {
	if !isAdmin {
		return apperror.NewForbidden(nil, "Chỉ Admin mới có quyền khóa/mở khóa sản phẩm")
	}

	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Check if active auction exists
	hasActive, err := s.auctionRepo.CheckActiveAuction(id)
	if err != nil {
		return err
	}
	if hasActive && product.Status != "banned" {
		return apperror.NewBadRequest(nil, "Không thể khóa sản phẩm khi đang có phiên đấu giá hoạt động")
	}

	if product.Status == "banned" {
		product.Status = "approved" // Or back to pending?
		product.RejectionReason = "" 
	} else {
		product.Status = "banned"
		product.RejectionReason = reason
	}

	return s.repo.Update(product)
}

func (s *productService) DeleteProductImage(productID uint, imageID uint, userID uint, isAdmin bool) error {
	product, err := s.repo.GetByID(productID)
	if err != nil {
		return err
	}

	if !isAdmin && product.SellerID != userID {
		return apperror.NewForbidden(nil, "Bạn không có quyền xóa ảnh của sản phẩm này")
	}

	return s.repo.DeleteImage(imageID)
}
