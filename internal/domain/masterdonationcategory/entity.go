package masterdonationcategory

type MasterDonationCategory struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedBy *int   `json:"created_by,omitempty"`
	UpdatedBy *int   `json:"updated_by,omitempty"`
	DeletedBy *int   `json:"deleted_by,omitempty"`
}
