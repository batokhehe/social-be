package levelarea

type LevelArea struct {
	IDLevelArea   int    `json:"id_level_area"`
	Level         int    `json:"level"`
	TingkatanArea string `json:"tingkatan_area"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	CreatedBy     *int   `json:"created_by,omitempty"`
	UpdatedBy     *int   `json:"updated_by,omitempty"`
	DeletedBy     *int   `json:"deleted_by,omitempty"`
}
