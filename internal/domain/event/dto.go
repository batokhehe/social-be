package event

type CreateRequest struct {
	Name               string `json:"name" validate:"required,max=100"`
	StartAt            string `json:"start_at" validate:"required"`
	EndAt              string `json:"end_at" validate:"required"`
	CategoryActivityID int    `json:"category_activity_id" validate:"required,gt=0"`
	ActivityID         int    `json:"activity_id" validate:"required,gt=0"`
	PicUserID          int    `json:"pic_user_id" validate:"required,gt=0"`
	Status             string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	Name               string `json:"name" validate:"required,max=100"`
	StartAt            string `json:"start_at" validate:"required"`
	EndAt              string `json:"end_at" validate:"required"`
	CategoryActivityID int    `json:"category_activity_id" validate:"required,gt=0"`
	ActivityID         int    `json:"activity_id" validate:"required,gt=0"`
	PicUserID          int    `json:"pic_user_id" validate:"required,gt=0"`
	Status             string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type CreateAttachmentRequest struct {
	Description string `form:"description" json:"description,omitempty"`
}
