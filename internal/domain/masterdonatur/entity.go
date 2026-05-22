package masterdonatur

type MasterDonatur struct {
	ID             int    `json:"id"`
	DonaturID      string `json:"id_donatur"`
	Phone          string `json:"telepon"`
	TzuChiAppID    string `json:"id_tzu_chi_app"`
	VisVolunteerID string `json:"id_vis_relawan"`
	DonaturGroupID *int   `json:"id_group_donatur,omitempty"`
	Area           string `json:"area"`
	Status         string `json:"status"`
	CreatedBy      *int   `json:"created_by,omitempty"`
	UpdatedBy      *int   `json:"updated_by,omitempty"`
	DeletedBy      *int   `json:"deleted_by,omitempty"`
}
