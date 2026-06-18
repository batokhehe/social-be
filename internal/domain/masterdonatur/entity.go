package masterdonatur

type MasterDonatur struct {
	ID             int               `json:"id"`
	DonaturID      string            `json:"id_donatur"`
	Phone          string            `json:"telepon"`
	TzuChiAppID    string            `json:"id_tzu_chi_app"`
	VisVolunteerID string            `json:"id_vis_volunteer"`
	DonaturGroupID *int              `json:"id_group_donatur,omitempty"`
	Name           string            `json:"name"`
	Status         string            `json:"status"`
	Volunteer      *DonaturVolunteer `json:"volunteer"`      // resolved from id_vis_volunteer -> volunteers.id; null if unset/not found
	DonaturGroup   *DonaturGroupInfo `json:"donatur_group"`  // resolved from id_group_donatur -> master_donatur_groups.id; null if unset/not found
	CreatedBy      *int              `json:"created_by,omitempty"`
	UpdatedBy      *int              `json:"updated_by,omitempty"`
	DeletedBy      *int              `json:"deleted_by,omitempty"`
}

// DonaturGroupInfo is the group resolved from id_group_donatur (master_donatur_groups.id).
type DonaturGroupInfo struct {
	ID      int    `json:"id"`
	GroupID string `json:"id_group_donatur"`
	Name    string `json:"name"`
	Status  string `json:"status"`
}

// DonaturVolunteer is the volunteer behind a donatur, with its area nested.
type DonaturVolunteer struct {
	ID             int             `json:"id"`
	VISID          string          `json:"vis_id"`
	IndonesianName string          `json:"indonesian_name"`
	MasterAreaID   int             `json:"master_area_id"`
	MasterArea     *MasterAreaInfo `json:"master_area"` // resolved from master_area_id -> master_areas.id
}

// MasterAreaInfo is the volunteer's area (volunteers.master_area_id -> master_areas.id).
type MasterAreaInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}
