package main

import (
	"fmt"
	"golang-app/internal/repository"
	"golang-app/internal/model/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=password dbname=golang_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	repo := repository.NewAuctionRepository(db)
	var dummy entity.Auction // Force usage
	_ = dummy
	auctions, total, err := repo.GetAuctions("active", "", "", 1, 10)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("TOTAL: %d\n", total)
	for _, a := range auctions {
		sellerName := "Unknown"
		if a.Product != nil && a.Product.Seller != nil {
			sellerName = a.Product.Seller.FullName
		}
		fmt.Printf("- Auction #%d: Product: %s, Seller: %s\n", a.ID, a.Product.Name, sellerName)
	}
}
