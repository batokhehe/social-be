package donation

import "time"

type Donation struct {
	ID                 int       `json:"id"`
	DonaturID          int       `json:"donatur_id"`
	DonaturGroupID     int       `json:"donatur_group_id"`
	AreaID             int       `json:"area_id"`
	DonationCategoryID int       `json:"donation_category_id"`
	Currency           string    `json:"currency"`
	Amount             float64   `json:"amount"`
	OtherItems         string    `json:"other_items"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
