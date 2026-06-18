package masterdonatur

type CreateRequest struct {
	DonaturID      string `json:"id_donatur" validate:"required,max=50"`
	Phone          string `json:"telepon" validate:"max=20"`
	TzuChiAppID    string `json:"id_tzu_chi_app" validate:"omitempty,max=50"`
	VisVolunteerID string `json:"id_vis_volunteer" validate:"omitempty,max=50"`
	DonaturGroupID *int   `json:"id_group_donatur" validate:"omitempty,gt=0"`
	Name           string `json:"name" validate:"required,max=100"`
	Status         string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	DonaturID      string `json:"id_donatur" validate:"required,max=50"`
	Phone          string `json:"telepon" validate:"max=20"`
	TzuChiAppID    string `json:"id_tzu_chi_app" validate:"omitempty,max=50"`
	VisVolunteerID string `json:"id_vis_volunteer" validate:"omitempty,max=50"`
	DonaturGroupID *int   `json:"id_group_donatur" validate:"omitempty,gt=0"`
	Name           string `json:"name" validate:"required,max=100"`
	Status         string `json:"status" validate:"omitempty,oneof=active inactive"`
}
