package donation

type CreateRequest struct {
	DonaturID          int     `json:"donatur_id" validate:"required,min=1"`
	DonaturGroupID     int     `json:"donatur_group_id" validate:"required,min=1"`
	AreaID             int     `json:"area_id" validate:"required,min=1"`
	DonationCategoryID int     `json:"donation_category_id" validate:"required,min=1"`
	Currency           string  `json:"currency" validate:"required,max=10"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OtherItems         string  `json:"other_items" validate:"omitempty,max=500"`
}

type UpdateRequest struct {
	DonaturID          int     `json:"donatur_id" validate:"required,min=1"`
	DonaturGroupID     int     `json:"donatur_group_id" validate:"required,min=1"`
	AreaID             int     `json:"area_id" validate:"required,min=1"`
	DonationCategoryID int     `json:"donation_category_id" validate:"required,min=1"`
	Currency           string  `json:"currency" validate:"required,max=10"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OtherItems         string  `json:"other_items" validate:"omitempty,max=500"`
}
