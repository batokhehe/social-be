package volunteer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"social-be/internal/pkg/mailer"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/security"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]Volunteer, int64, error)
	GetByID(ctx context.Context, id int) (*Volunteer, error)
	GetByUserID(ctx context.Context, userID int) (*Volunteer, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Volunteer, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
	ValidateResetToken(ctx context.Context, token string) (*ResetPasswordValidation, error)
	ResetPassword(ctx context.Context, token, password string) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type volunteerModel struct {
	ID                   int        `gorm:"column:id"`
	UserID               int        `gorm:"column:user_id"`
	TzuhiAppID           string     `gorm:"column:tzuhi_app_id"`
	VISID                string     `gorm:"column:vis_id"`
	IndonesianName       string     `gorm:"column:indonesian_name"`
	MandarinName         string     `gorm:"column:mandarin_name"`
	BirthPlace           string     `gorm:"column:birth_place"`
	BirthDate            time.Time  `gorm:"column:birth_date"`
	MasterAreaID         int        `gorm:"column:master_area_id"`
	LevelVolunteerID     int        `gorm:"column:level_volunteer_id"`
	AttributeVolunteerID *int       `gorm:"column:attribute_volunteer_id"`
	ReligionID           int        `gorm:"column:religion_id"`
	BloodType            string     `gorm:"column:blood_type"`
	Rhesus               string     `gorm:"column:rhesus"`
	LastEducation        string     `gorm:"column:last_education"`
	MaritalStatus        string     `gorm:"column:marital_status"`
	Profession           string     `gorm:"column:profession"`
	Field                string     `gorm:"column:field"`
	ResidentialAddress   string     `gorm:"column:residential_address"`
	PostalCode           string     `gorm:"column:postal_code"`
	HomePhone            string     `gorm:"column:home_phone"`
	OfficePhone          string     `gorm:"column:office_phone"`
	Phone                string     `gorm:"column:phone"`
	Email                string     `gorm:"column:email"`
	Languages            string     `gorm:"column:languages"`
	PrivateVehicle       bool       `gorm:"column:private_vehicle"`
	RegularDonor         bool       `gorm:"column:regular_donor"`
	Status               string     `gorm:"column:status"`
	CreatedBy            *int       `gorm:"column:created_by"`
	UpdatedBy            *int       `gorm:"column:updated_by"`
	DeletedBy            *int       `gorm:"column:deleted_by"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
}

func (volunteerModel) TableName() string { return "volunteers" }

type userModel struct {
	ID           int    `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	Email        string `gorm:"column:email"`
	PasswordHash string `gorm:"column:password_hash"`
	Role         int    `gorm:"column:role"`
	Status       int    `gorm:"column:status"`
	CreatedBy    *int   `gorm:"column:created_by"`
	UpdatedBy    *int   `gorm:"column:updated_by"`
}

func (userModel) TableName() string { return "users" }

type passwordResetTokenModel struct {
	UserID    int       `gorm:"column:user_id"`
	Token     string    `gorm:"column:token"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
}

func (passwordResetTokenModel) TableName() string { return "password_reset_tokens" }

type volunteerActivityModel struct {
	VolunteerID      int `gorm:"column:volunteer_id"`
	MasterActivityID int `gorm:"column:master_activity_id"`
}

func (volunteerActivityModel) TableName() string {
	return "volunteer_activities"
}

type volunteerFreeTimeModel struct {
	VolunteerID int    `gorm:"column:volunteer_id"`
	DayOfWeek   string `gorm:"column:day_of_week"`
	Period      string `gorm:"column:period"`
}

func (volunteerFreeTimeModel) TableName() string {
	return "volunteer_free_times"
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error) {
	var id int
	var resetLink string
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := buildModel(req, actorID)
		if err != nil {
			return err
		}
		userID, link, err := createVolunteerUserAndResetToken(tx, req, actorID)
		if err != nil {
			return err
		}
		row.UserID = userID
		resetLink = link
		code, err := nextTzuhiAppID(tx)
		if err != nil {
			return err
		}
		row.TzuhiAppID = code
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		id = row.ID
		return replaceChildren(tx, id, req.InterestedActivityIDs, req.FreeTimes)
	})
	if err != nil {
		r.Logger.WithError(err).Error("Create Volunteer failed")
		return nil, fmt.Errorf("failed create volunteer: %w", err)
	}
	if err := mailer.SendPasswordReset(ctx, req.Email, req.IndonesianName, resetLink); err != nil {
		r.Logger.WithError(err).WithField("email", req.Email).Error("failed send password reset email")
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]Volunteer, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&volunteerModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []volunteerModel
	if err := baseQuery.Order("id DESC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]Volunteer, 0, len(rows))
	for _, row := range rows {
		item := toEntity(row)
		if err := r.loadChildren(ctx, &item); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Volunteer, error) {
	var row volunteerModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, err
	}
	item := toEntity(row)
	if err := r.loadChildren(ctx, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) GetByUserID(ctx context.Context, userID int) (*Volunteer, error) {
	var row volunteerModel
	if err := r.DB.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Take(&row).Error; err != nil {
		return nil, err
	}
	item := toEntity(row)
	if err := r.loadChildren(ctx, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) ValidateResetToken(ctx context.Context, token string) (*ResetPasswordValidation, error) {
	var row passwordResetTokenModel
	if err := r.DB.WithContext(ctx).Where("token = ? AND used_at IS NULL AND expires_at > ?", token, time.Now()).Take(&row).Error; err != nil {
		return nil, err
	}

	var user userModel
	if err := r.DB.WithContext(ctx).Where("id = ?", row.UserID).Take(&user).Error; err != nil {
		return nil, err
	}

	return &ResetPasswordValidation{Email: user.Email, Valid: true, ExpiresAt: row.ExpiresAt}, nil
}

func (r *GormRepository) ResetPassword(ctx context.Context, token, password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row passwordResetTokenModel
		if err := tx.Where("token = ? AND used_at IS NULL AND expires_at > ?", token, time.Now()).Take(&row).Error; err != nil {
			return err
		}
		if err := tx.Model(&userModel{}).Where("id = ?", row.UserID).Update("password_hash", hash).Error; err != nil {
			return err
		}
		return tx.Model(&passwordResetTokenModel{}).Where("token = ?", token).Update("used_at", time.Now()).Error
	})
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Volunteer, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates, err := buildModel(req, actorID)
		if err != nil {
			return err
		}
		result := tx.Model(&volunteerModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
			"vis_id": req.VISID, "indonesian_name": req.IndonesianName, "mandarin_name": req.MandarinName,
			"birth_place": updates.BirthPlace, "birth_date": updates.BirthDate, "master_area_id": req.MasterAreaID,
			"level_volunteer_id": req.LevelVolunteerID, "attribute_volunteer_id": req.AttributeVolunteerID,
			"religion_id": req.ReligionID, "blood_type": req.BloodType, "rhesus": req.Rhesus,
			"last_education": req.LastEducation, "marital_status": req.MaritalStatus, "profession": req.Profession,
			"field": req.Field, "residential_address": req.ResidentialAddress, "postal_code": req.PostalCode,
			"home_phone": req.HomePhone, "office_phone": req.OfficePhone, "phone": req.Phone, "email": req.Email,
			"languages": req.Languages, "private_vehicle": req.PrivateVehicle, "regular_donor": req.RegularDonor,
			"updated_by": actorID, "updated_at": time.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return replaceChildren(tx, id, req.InterestedActivityIDs, req.FreeTimes)
	})
	if err != nil {
		return nil, fmt.Errorf("failed update volunteer: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&volunteerModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"status": "inactive", "deleted_by": actorID, "deleted_at": time.Now(), "updated_by": actorID, "updated_at": time.Now(),
	})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func buildModel(req CreateRequest, actorID int) (volunteerModel, error) {
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return volunteerModel{}, fmt.Errorf("birth_date must use YYYY-MM-DD format: %w", err)
	}
	return volunteerModel{
		VISID: req.VISID, IndonesianName: req.IndonesianName, MandarinName: req.MandarinName,
		BirthPlace: req.BirthPlace, BirthDate: birthDate, MasterAreaID: req.MasterAreaID,
		LevelVolunteerID: req.LevelVolunteerID, AttributeVolunteerID: req.AttributeVolunteerID, ReligionID: req.ReligionID,
		BloodType: req.BloodType, Rhesus: req.Rhesus, LastEducation: req.LastEducation, MaritalStatus: req.MaritalStatus,
		Profession: req.Profession, Field: req.Field, ResidentialAddress: req.ResidentialAddress, PostalCode: req.PostalCode,
		HomePhone: req.HomePhone, OfficePhone: req.OfficePhone, Phone: req.Phone, Email: req.Email,
		Languages: req.Languages, PrivateVehicle: req.PrivateVehicle, RegularDonor: req.RegularDonor,
		Status: "active", CreatedBy: &actorID, UpdatedBy: &actorID,
	}, nil
}

func createVolunteerUserAndResetToken(tx *gorm.DB, req CreateRequest, actorID int) (int, string, error) {
	password := "TZ12345"
	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return 0, "", err
	}
	user := userModel{
		Name:         req.IndonesianName,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         2,
		Status:       1,
		CreatedBy:    &actorID,
		UpdatedBy:    &actorID,
	}
	if err := tx.Create(&user).Error; err != nil {
		return 0, "", err
	}

	token, err := randomToken(32)
	if err != nil {
		return 0, "", err
	}
	resetToken := passwordResetTokenModel{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := tx.Create(&resetToken).Error; err != nil {
		return 0, "", err
	}

	fmt.Println("USER ID:", user.ID)
	fmt.Println("EMAIL:", user.Email)
	fmt.Println("PASSWORD: TZ12345")
	return user.ID, buildResetLink(token), nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func buildResetLink(token string) string {
	baseURL := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL + "/reset-password?token=" + url.QueryEscape(token)
}

func nextTzuhiAppID(tx *gorm.DB) (string, error) {
	prefix := time.Now().Format("0601")
	var count int64
	if err := tx.Model(&volunteerModel{}).Where("tzuhi_app_id LIKE ?", prefix+"%").Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%05d", prefix, count+1), nil
}

func replaceChildren(tx *gorm.DB, id int, activityIDs []int, freeTimes []FreeTimeRequest) error {
	if err := tx.Where("volunteer_id = ?", id).Delete(&volunteerActivityModel{}).Error; err != nil {
		return err
	}
	for _, activityID := range activityIDs {
		if err := tx.Create(&volunteerActivityModel{VolunteerID: id, MasterActivityID: activityID}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("volunteer_id = ?", id).Delete(&volunteerFreeTimeModel{}).Error; err != nil {
		return err
	}
	for _, freeTime := range freeTimes {
		if err := tx.Create(&volunteerFreeTimeModel{VolunteerID: id, DayOfWeek: freeTime.DayOfWeek, Period: freeTime.Period}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GormRepository) loadChildren(ctx context.Context, item *Volunteer) error {
	var activities []volunteerActivityModel
	if err := r.DB.WithContext(ctx).Where("volunteer_id = ?", item.ID).Find(&activities).Error; err != nil {
		return err
	}
	for _, activity := range activities {
		item.InterestedActivityIDs = append(item.InterestedActivityIDs, activity.MasterActivityID)
	}
	var freeTimes []volunteerFreeTimeModel
	if err := r.DB.WithContext(ctx).Where("volunteer_id = ?", item.ID).Find(&freeTimes).Error; err != nil {
		return err
	}
	for _, freeTime := range freeTimes {
		item.FreeTimes = append(item.FreeTimes, FreeTime{DayOfWeek: freeTime.DayOfWeek, Period: freeTime.Period})
	}
	return nil
}

func toEntity(row volunteerModel) Volunteer {
	return Volunteer{
		ID: row.ID, UserID: row.UserID, TzuhiAppID: row.TzuhiAppID, VISID: row.VISID, IndonesianName: row.IndonesianName, MandarinName: row.MandarinName,
		BirthPlace: row.BirthPlace, BirthDate: row.BirthDate.Format("2006-01-02"), MasterAreaID: row.MasterAreaID,
		LevelVolunteerID: row.LevelVolunteerID, AttributeVolunteerID: row.AttributeVolunteerID, ReligionID: row.ReligionID,
		BloodType: row.BloodType, Rhesus: row.Rhesus, LastEducation: row.LastEducation, MaritalStatus: row.MaritalStatus,
		Profession: row.Profession, Field: row.Field, ResidentialAddress: row.ResidentialAddress, PostalCode: row.PostalCode,
		HomePhone: row.HomePhone, OfficePhone: row.OfficePhone, Phone: row.Phone, Email: row.Email, Languages: row.Languages,
		PrivateVehicle: row.PrivateVehicle, RegularDonor: row.RegularDonor, Status: row.Status,
		CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, DeletedBy: row.DeletedBy,
	}
}
