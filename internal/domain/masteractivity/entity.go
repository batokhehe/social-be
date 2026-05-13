package masteractivity

type MasterActivity struct {
	ID                 int    `json:"id"`
	CategoryActivityID int    `json:"category_activity_id"`
	Name               string `json:"name"`
	Target             int    `json:"target"`
	Description        string `json:"description"`
	Status             string `json:"status"`
	CreatedBy          *int   `json:"created_by,omitempty"`
	UpdatedBy          *int   `json:"updated_by,omitempty"`
	DeletedBy          *int   `json:"deleted_by,omitempty"`
}
