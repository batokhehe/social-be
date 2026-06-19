package donation

import (
	"time"

	"gorm.io/gorm"
)

// Donation type values.
const (
	DonationTypeMoney = 0 // money donation
	DonationTypeGoods = 1 // goods donation
)

type Donation struct {
	ID                 int            `json:"id"`
	DonaturID          int            `json:"donatur_id"`
	DonaturGroupID     *int           `json:"donatur_group_id"` // nullable: Excel imports carry no group
	AreaID             *int           `json:"area_id"`          // nullable: Excel imports carry no area
	DonationCategoryID int            `json:"donation_category_id"`
	Type               int            `json:"type"`             // 0 = money, 1 = goods
	Period             *time.Time     `json:"period"`           // donation period (datetime), nullable
	Currency           string         `json:"currency"`
	Amount             float64        `json:"amount"`
	OtherItems         *string        `json:"other_items"` // NULL for money donations (type 0)
	ImportBatchID      *string        `json:"import_batch_id,omitempty"`
	CreatedBy          *int           `json:"created_by,omitempty"`
	UpdatedBy          *int           `json:"updated_by,omitempty"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedBy          *int           `json:"deleted_by,omitempty"`
	DeletedAt          gorm.DeletedAt `json:"-"` // enables GORM soft delete + auto query filtering
	// Resolved references (read-only enrichment, populated in the repository --
	// not DB columns, so GORM must ignore them). Null when the referenced row is
	// missing or soft-deleted, or when donatur_group_id is NULL.
	Donatur      *DonaturInfo      `json:"donatur" gorm:"-"`
	DonaturGroup *DonaturGroupInfo `json:"donatur_group" gorm:"-"`
}

// DonaturInfo is the donor resolved from donations.donatur_id (master_donaturs.id).
type DonaturInfo struct {
	ID             int               `json:"id"`
	DonaturID      string            `json:"id_donatur"`
	Name           string            `json:"name"`
	Phone          string            `json:"phone"`
	Status         string            `json:"status"`
	VisVolunteerID string            `json:"id_vis_volunteer"`
	Volunteer      *DonaturVolunteer `json:"volunteer"` // resolved from id_vis_volunteer -> volunteers.id
}

// DonaturVolunteer is the volunteer behind a donor (master_donaturs.id_vis_volunteer
// -> volunteers.id), with its area nested.
type DonaturVolunteer struct {
	ID             int                `json:"id"`
	VISID          string             `json:"vis_id"`
	IndonesianName string             `json:"indonesian_name"`
	MasterAreaID   int                `json:"master_area_id"`
	MasterArea     *DonaturMasterArea `json:"master_area"` // resolved from master_area_id -> master_areas.id
}

// DonaturMasterArea is the volunteer's area (volunteers.master_area_id -> master_areas.id).
type DonaturMasterArea struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// DonaturGroupInfo is the group resolved from donations.donatur_group_id
// (master_donatur_groups.id).
type DonaturGroupInfo struct {
	ID      int    `json:"id"`
	GroupID string `json:"id_group_donatur"`
	Name    string `json:"name"`
	Status  string `json:"status"`
}
