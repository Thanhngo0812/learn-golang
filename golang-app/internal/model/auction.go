package entity

import "time"

// Auction Entity
type Auction struct {
	ID                uint      `gorm:"primaryKey;column:id" json:"id"`
	ProductID         uint      `gorm:"column:product_id;unique;not null" json:"product_id"`
	StartPrice        float64   `gorm:"column:start_price;not null" json:"start_price"`
	StepPrice         float64   `gorm:"column:step_price;not null;default:10000" json:"step_price"`
	CurrentPrice      float64   `gorm:"column:current_price;not null;default:0" json:"current_price"`
	StartTime         time.Time `gorm:"column:start_time;not null" json:"start_time"`
	EndTime           time.Time `gorm:"column:end_time;not null" json:"end_time"`
	Status            string    `gorm:"column:status;default:'pending'" json:"status"`

	// Số đếm lượng bid (dùng để map vào lúc Select Count)
	BidCount int `gorm:"->;column:bid_count" json:"bid_count,omitempty"`

	// Relationship (Kỹ thuật 3: Phục vụ cho Eager Loading Preload)
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WinnerID          *uint     `gorm:"column:winner_id" json:"winner_id"`
	BuyNowPrice       *float64  `gorm:"column:buy_now_price" json:"buy_now_price"`
	IsAutoExtend      bool      `gorm:"column:is_auto_extend;default:true" json:"is_auto_extend"`
	ExtendTimeSeconds int       `gorm:"column:extend_time_seconds;default:300" json:"extend_time_seconds"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	SellerConfirmed   bool      `gorm:"column:seller_confirmed;default:false" json:"seller_confirmed"`
	BuyerConfirmed    bool      `gorm:"column:buyer_confirmed;default:false" json:"buyer_confirmed"`
	RejectionReason   string    `gorm:"column:rejection_reason" json:"rejection_reason"`
	Bids              []Bid     `gorm:"foreignKey:AuctionID" json:"bids,omitempty"`
}

func (Auction) TableName() string {
	return "auctions"
}
