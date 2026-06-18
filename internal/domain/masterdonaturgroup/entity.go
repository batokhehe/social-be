package masterdonaturgroup

type MasterDonaturGroup struct {
	ID          int    `json:"id"`
	GroupID     string `json:"id_group_donatur"`
	Name        string `json:"name"`
	VolunteerID string `json:"volunteer_id"`
	PIC         *PIC   `json:"pic"` // volunteer resolved from volunteer_id (volunteers.id); null if not found
	PICPhone    string `json:"pic_phone"`
	Status      string `json:"status"`
	CreatedBy   *int   `json:"created_by,omitempty"`
	UpdatedBy   *int   `json:"updated_by,omitempty"`
	DeletedBy   *int   `json:"deleted_by,omitempty"`
}

// PIC is the volunteer in charge of a donatur group, looked up from the
// volunteers table by volunteer_id (matched against volunteers.id).
type PIC struct {
	ID             int    `json:"id"`
	IndonesianName string `json:"indonesian_name"`
	VISID          string `json:"vis_id"`
	Phone          string `json:"phone"`
}
