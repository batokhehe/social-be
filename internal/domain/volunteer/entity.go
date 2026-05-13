package volunteer

type FreeTime struct {
	DayOfWeek string `json:"day_of_week"`
	Period    string `json:"period"`
}

type Volunteer struct {
	ID                    int        `json:"id"`
	UserID                int        `json:"user_id"`
	TzuhiAppID            string     `json:"tzuhi_app_id"`
	VISID                 string     `json:"vis_id"`
	IndonesianName        string     `json:"indonesian_name"`
	MandarinName          string     `json:"mandarin_name"`
	BirthPlace            string     `json:"birth_place"`
	BirthDate             string     `json:"birth_date"`
	MasterAreaID          int        `json:"master_area_id"`
	LevelVolunteerID      int        `json:"level_volunteer_id"`
	AttributeVolunteerID  int        `json:"attribute_volunteer_id"`
	ReligionID            int        `json:"religion_id"`
	BloodType             string     `json:"blood_type"`
	Rhesus                string     `json:"rhesus"`
	LastEducation         string     `json:"last_education"`
	MaritalStatus         string     `json:"marital_status"`
	Profession            string     `json:"profession"`
	Field                 string     `json:"field"`
	ResidentialAddress    string     `json:"residential_address"`
	PostalCode            string     `json:"postal_code"`
	HomePhone             string     `json:"home_phone"`
	OfficePhone           string     `json:"office_phone"`
	Phone                 string     `json:"phone"`
	Email                 string     `json:"email"`
	Languages             string     `json:"languages"`
	PrivateVehicle        bool       `json:"private_vehicle"`
	RegularDonor          bool       `json:"regular_donor"`
	InterestedActivityIDs []int      `json:"interested_activity_ids,omitempty"`
	FreeTimes             []FreeTime `json:"free_times,omitempty"`
	Status                string     `json:"status"`
	CreatedBy             *int       `json:"created_by,omitempty"`
	UpdatedBy             *int       `json:"updated_by,omitempty"`
	DeletedBy             *int       `json:"deleted_by,omitempty"`
}
