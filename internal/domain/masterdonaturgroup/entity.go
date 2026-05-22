package masterdonaturgroup

type MasterDonaturGroup struct {
	ID        int    `json:"id"`
	GroupID   string `json:"id_group_donatur"`
	Name      string `json:"name"`
	PICName   string `json:"pic_name"`
	PICPhone  string `json:"pic_phone"`
	Status    string `json:"status"`
	CreatedBy *int   `json:"created_by,omitempty"`
	UpdatedBy *int   `json:"updated_by,omitempty"`
	DeletedBy *int   `json:"deleted_by,omitempty"`
}
