package levelvolunteer

type LevelVolunteer struct {
	ID          int    `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedBy   *int   `json:"created_by,omitempty"`
	UpdatedBy   *int   `json:"updated_by,omitempty"`
	DeletedBy   *int   `json:"deleted_by,omitempty"`
}
