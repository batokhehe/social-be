package volunteer

import "time"

type FreeTimeRequest struct {
	DayOfWeek string `json:"day_of_week" validate:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	Period    string `json:"period" validate:"required,oneof=morning afternoon evening"`
}

type ResetPasswordVerifyRequest struct {
	Token string `json:"token" validate:"required,max=128"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,max=128"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type ResetPasswordValidation struct {
	Email     string    `json:"email"`
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CreateRequest struct {
	VISID                 string            `json:"vis_id" validate:"required,max=50"`
	IndonesianName        string            `json:"indonesian_name" validate:"required,max=100"`
	MandarinName          string            `json:"mandarin_name" validate:"omitempty,max=100"`
	BirthPlace            string            `json:"birth_place" validate:"required,max=100"`
	BirthDate             string            `json:"birth_date" validate:"required"`
	MasterAreaID          int               `json:"master_area_id" validate:"required,min=1"`
	LevelVolunteerID      int               `json:"level_volunteer_id" validate:"required,min=1"`
	AttributeVolunteerID  int               `json:"attribute_volunteer_id" validate:"required,min=1"`
	ReligionID            int               `json:"religion_id" validate:"required,min=1"`
	BloodType             string            `json:"blood_type" validate:"required,oneof=A B AB O"`
	Rhesus                string            `json:"rhesus" validate:"required,oneof=positive negative unknown"`
	LastEducation         string            `json:"last_education" validate:"required,oneof=SD SMP SMA D1 D2 D3 D4 S1 S2 S3"`
	MaritalStatus         string            `json:"marital_status" validate:"required,oneof=single married divorced widowed"`
	Profession            string            `json:"profession" validate:"required,max=100"`
	Field                 string            `json:"field" validate:"required,max=100"`
	ResidentialAddress    string            `json:"residential_address" validate:"required,max=255"`
	PostalCode            string            `json:"postal_code" validate:"required,max=20"`
	HomePhone             string            `json:"home_phone" validate:"omitempty,max=30"`
	OfficePhone           string            `json:"office_phone" validate:"omitempty,max=30"`
	Phone                 string            `json:"phone" validate:"required,max=30"`
	Email                 string            `json:"email" validate:"required,email,max=100"`
	InterestedActivityIDs []int             `json:"interested_activity_ids" validate:"required,min=1,dive,min=1"`
	Languages             string            `json:"languages" validate:"required,max=255"`
	PrivateVehicle        bool              `json:"private_vehicle"`
	RegularDonor          bool              `json:"regular_donor"`
	FreeTimes             []FreeTimeRequest `json:"free_times" validate:"required,min=1,dive"`
}

type UpdateRequest = CreateRequest
