package masterarea

type MasterArea struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LevelAreaID string `json:"level_area_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ParentID    *int   `json:"parent_id,omitempty"`
	Location    string `json:"location"`
	CreatedBy   *int   `json:"created_by,omitempty"`
	UpdatedBy   *int   `json:"updated_by,omitempty"`
	DeletedBy   *int   `json:"deleted_by,omitempty"`
}
