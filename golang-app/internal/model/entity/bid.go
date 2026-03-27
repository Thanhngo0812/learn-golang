package entity

import "time"

// Bid Entity maps to the `bids` table in the database
type Bid struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	AuctionID uint      `gorm:"column:auction_id;not null;index" json:"auction_id"` // Kỹ thuật 1: Đánh Index để tối ưu GROUP BY
	UserID    uint      `gorm:"column:user_id;not null" json:"user_id"`
	Amount    float64   `gorm:"column:amount;not null" json:"amount"`
	BidTime   time.Time `gorm:"column:bid_time;autoCreateTime" json:"bid_time"`

	// Relationships
	Auction *Auction `gorm:"foreignKey:AuctionID" json:"auction,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Bid) TableName() string {
	return "bids"
}
