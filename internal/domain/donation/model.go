package donation

import (
	"time"

	"gorm.io/gorm"
)

type Donation struct {
	ID                 int            `json:"id"`
	DonaturID          int            `json:"donatur_id"`
	DonaturGroupID     *int           `json:"donatur_group_id"` // nullable: Excel imports carry no group
	AreaID             *int           `json:"area_id"`          // nullable: Excel imports carry no area
	DonationCategoryID int            `json:"donation_category_id"`
	Currency           string         `json:"currency"`
	Amount             float64        `json:"amount"`
	OtherItems         string         `json:"other_items"`
	ImportBatchID      *string        `json:"import_batch_id,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedBy          *int           `json:"deleted_by,omitempty"`
	DeletedAt          gorm.DeletedAt `json:"-"` // enables GORM soft delete + auto query filtering
}
