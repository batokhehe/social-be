package masterdonaturgroup

type CreateRequest struct {
	GroupID     string `json:"id_group_donatur" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=100"`
	VolunteerID string `json:"volunteer_id" validate:"required,max=100"`
	PICPhone    string `json:"pic_phone" validate:"required,max=20"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	GroupID     string `json:"id_group_donatur" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=100"`
	VolunteerID string `json:"volunteer_id" validate:"required,max=100"`
	PICPhone    string `json:"pic_phone" validate:"required,max=20"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}
