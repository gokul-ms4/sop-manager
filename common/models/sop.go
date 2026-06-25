package models

import "time"

type SopHeading struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Heading     string    `json:"heading"`
	Description string    `json:"description"`
	SopItems    []SopItem `json:"sop_items" gorm:"foreignKey:HeadingID"`
	CreatedBy   int       `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type SopItem struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	HeadingID int       `json:"heading_id" gorm:"constraint:OnDelete:CASCADE"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Position  int      `json:"position"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type SopChunk struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SopHeadingID uint      `json:"sop_heading_id"`
	SopItemID    uint      `json:"sop_item_id"`
	ChunkText    string    `gorm:"type:text" json:"chunk_text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Embedding    string    `gorm:"type:vector(3072)" json:"embedding"`
}
