package expense

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound   = errors.New("category not found")
	ErrCategoryInactive   = errors.New("category is inactive")
	ErrVolunteerNotFound  = errors.New("volunteer not found")
	ErrExpenseNotFound    = errors.New("expense not found")
	ErrInvalidExpenseDate = errors.New("expense_date must be a valid date/time")
)

// ExpenseSort is a validated (whitelisted) ORDER BY.
type ExpenseSort struct {
	Column string
	Order  string
}

func (s ExpenseSort) orderClause() string {
	// id tiebreaker keeps ordering stable across pages when the primary column
	// has duplicate values (e.g. many rows sharing an expense_date).
	return s.Column + " " + s.Order + ", id " + s.Order
}

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*Expense, error)
	GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort ExpenseSort) ([]ExpenseListItem, int64, error)
	GetByID(ctx context.Context, id int) (*Expense, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type expenseModel struct {
	ID          int        `gorm:"column:id"`
	ExpenseNo   string     `gorm:"column:expense_no"`
	ExpenseDate time.Time  `gorm:"column:expense_date"`
	CategoryID  int        `gorm:"column:category_id"`
	VolunteerID int        `gorm:"column:volunteer_id"`
	Amount      float64    `gorm:"column:amount"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (expenseModel) TableName() string { return "expenses" }

func toEntity(row expenseModel) Expense {
	return Expense{
		ID: row.ID, ExpenseNo: row.ExpenseNo, ExpenseDate: row.ExpenseDate,
		CategoryID: row.CategoryID, VolunteerID: row.VolunteerID, Amount: row.Amount,
		Description: row.Description, Status: row.Status,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// buildExpenseNo formats EXP-YYYYMM-00001. The month prefix comes from the
// expense date, so the running number resets per transaction month.
func buildExpenseNo(t time.Time, seq int) string {
	return fmt.Sprintf("EXP-%s-%05d", t.Format("200601"), seq)
}

func parseExpenseDate(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalidExpenseDate
}

// nextExpenseNo allocates the next running number for expenseDate's month inside
// the caller's transaction. A per-month advisory lock serialises concurrent
// allocations so no two expenses get the same number (the UNIQUE index is the
// final backstop). MAX(...) -- not COUNT -- so soft-deleted rows never cause a
// number to be reissued.
func nextExpenseNo(tx *gorm.DB, expenseDate time.Time) (string, error) {
	month := expenseDate.Format("200601")
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "expenses:"+month).Error; err != nil {
		return "", err
	}
	var maxSeq int
	if err := tx.Raw(
		`SELECT COALESCE(MAX(CAST(RIGHT(expense_no, 5) AS INTEGER)), 0) FROM expenses WHERE expense_no LIKE ?`,
		"EXP-"+month+"-%",
	).Scan(&maxSeq).Error; err != nil {
		return "", err
	}
	return buildExpenseNo(expenseDate, maxSeq+1), nil
}

// ensureCategory requires the category to exist, be live (deleted_at IS NULL)
// and be active. Inactive/retired categories are not selectable for expenses.
func ensureCategory(tx *gorm.DB, id int) error {
	var actives []bool
	if err := tx.Raw(
		`SELECT active FROM master_expense_categories WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&actives).Error; err != nil {
		return err
	}
	if len(actives) == 0 {
		return ErrCategoryNotFound
	}
	if !actives[0] {
		return ErrCategoryInactive
	}
	return nil
}

func ensureVolunteer(tx *gorm.DB, id int) error {
	var exists bool
	if err := tx.Raw(
		`SELECT EXISTS(SELECT 1 FROM volunteers WHERE id = ? AND deleted_at IS NULL)`, id,
	).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return ErrVolunteerNotFound
	}
	return nil
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
	expenseDate, err := parseExpenseDate(req.ExpenseDate)
	if err != nil {
		return nil, err
	}

	var newID int
	err = r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureCategory(tx, req.CategoryID); err != nil {
			return err
		}
		if err := ensureVolunteer(tx, req.VolunteerID); err != nil {
			return err
		}
		expenseNo, err := nextExpenseNo(tx, expenseDate)
		if err != nil {
			return err
		}
		row := expenseModel{
			ExpenseNo: expenseNo, ExpenseDate: expenseDate,
			CategoryID: req.CategoryID, VolunteerID: req.VolunteerID,
			Amount: req.Amount, Description: req.Description, Status: req.Status,
			CreatedBy: &actorID, UpdatedBy: &actorID,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		newID = row.ID
		return nil
	})
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) && !errors.Is(err, ErrCategoryInactive) && !errors.Is(err, ErrVolunteerNotFound) {
			r.Logger.WithError(err).Error("Create Expense failed")
		}
		return nil, err
	}
	return r.GetByID(ctx, newID)
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort ExpenseSort) ([]ExpenseListItem, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Model(&expenseModel{}).Where("deleted_at IS NULL")
	baseQuery = query.ApplyFilters(baseQuery, filters)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count Expense failed")
		return nil, 0, fmt.Errorf("failed count expense: %w", err)
	}

	var rows []expenseModel
	if err := baseQuery.
		Order(sort.orderClause()).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated Expense failed")
		return nil, 0, fmt.Errorf("failed get paginated expense: %w", err)
	}

	items := make([]ExpenseListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ExpenseListItem{
			ID: row.ID, ExpenseNo: row.ExpenseNo, ExpenseDate: row.ExpenseDate,
			Amount: row.Amount, Status: row.Status, CreatedAt: row.CreatedAt,
		})
	}

	if err := r.fillListNames(ctx, rows, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// fillListNames resolves category + volunteer names in two batched queries
// (no N+1), then maps them onto the list items by id.
func (r *GormRepository) fillListNames(ctx context.Context, rows []expenseModel, items []ExpenseListItem) error {
	if len(rows) == 0 {
		return nil
	}

	categoryIDs := make([]int, 0, len(rows))
	volunteerIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		categoryIDs = append(categoryIDs, row.CategoryID)
		volunteerIDs = append(volunteerIDs, row.VolunteerID)
	}

	categoryNames, err := r.namesByID(ctx, "master_expense_categories", "name", categoryIDs)
	if err != nil {
		return err
	}
	volunteerNames, err := r.namesByID(ctx, "volunteers", "indonesian_name", volunteerIDs)
	if err != nil {
		return err
	}

	for i, row := range rows {
		items[i].CategoryName = categoryNames[row.CategoryID]
		items[i].VolunteerName = volunteerNames[row.VolunteerID]
	}
	return nil
}

// namesByID resolves id -> display name from a reference table in one query.
// Soft-deleted rows are intentionally included so a historical expense keeps
// showing its category/volunteer even after that record is deactivated.
func (r *GormRepository) namesByID(ctx context.Context, table, nameColumn string, ids []int) (map[int]string, error) {
	type row struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var rows []row
	if err := r.DB.WithContext(ctx).
		Table(table).
		Select("id", nameColumn+" AS name").
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed load %s names: %w", table, err)
	}
	out := make(map[int]string, len(rows))
	for _, x := range rows {
		out[x.ID] = x.Name
	}
	return out, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Expense, error) {
	var row expenseModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}

	item := toEntity(row)
	if err := r.loadCategory(ctx, &item); err != nil {
		return nil, err
	}
	if err := r.loadVolunteer(ctx, &item); err != nil {
		return nil, err
	}
	if err := r.loadAuditUsers(ctx, &item, row); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) loadCategory(ctx context.Context, item *Expense) error {
	var cat CategoryInfo
	if err := r.DB.WithContext(ctx).
		Table("master_expense_categories").
		Select("id", "code", "name").
		Where("id = ?", item.CategoryID).
		Take(&cat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed load expense category: %w", err)
	}
	item.Category = &cat
	return nil
}

func (r *GormRepository) loadVolunteer(ctx context.Context, item *Expense) error {
	var vol VolunteerInfo
	// No deleted_at filter (see namesByID): keep resolving the volunteer on a
	// historical expense even after the volunteer is deactivated.
	if err := r.DB.WithContext(ctx).
		Table("volunteers").
		Select("id", "indonesian_name", "master_area_id").
		Where("id = ?", item.VolunteerID).
		Take(&vol).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed load expense volunteer: %w", err)
	}
	item.Volunteer = &vol
	return nil
}

// loadAuditUsers resolves created/updated/deleted by user names in one query.
func (r *GormRepository) loadAuditUsers(ctx context.Context, item *Expense, row expenseModel) error {
	idSet := make(map[int]struct{})
	for _, p := range []*int{row.CreatedBy, row.UpdatedBy, row.DeletedBy} {
		if p != nil {
			idSet[*p] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return nil
	}
	ids := make([]int, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	var users []UserInfo
	if err := r.DB.WithContext(ctx).
		Table("users").
		Select("id", "name").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return fmt.Errorf("failed load expense audit users: %w", err)
	}
	byID := make(map[int]UserInfo, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}

	assign := func(actorID *int, dst **UserInfo) {
		if actorID == nil {
			return
		}
		if u, ok := byID[*actorID]; ok {
			ref := u
			*dst = &ref
		}
	}
	assign(row.CreatedBy, &item.CreatedBy)
	assign(row.UpdatedBy, &item.UpdatedBy)
	assign(row.DeletedBy, &item.DeletedBy)
	return nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
	expenseDate, err := parseExpenseDate(req.ExpenseDate)
	if err != nil {
		return nil, err
	}

	err = r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureCategory(tx, req.CategoryID); err != nil {
			return err
		}
		if err := ensureVolunteer(tx, req.VolunteerID); err != nil {
			return err
		}
		// expense_no is immutable and intentionally not updated.
		result := tx.Model(&expenseModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
			"expense_date": expenseDate,
			"category_id":  req.CategoryID,
			"volunteer_id": req.VolunteerID,
			"amount":       req.Amount,
			"description":  req.Description,
			"status":       req.Status,
			"updated_by":   actorID,
			"updated_at":   time.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrExpenseNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	result := r.DB.WithContext(ctx).Model(&expenseModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": actorID,
			"updated_by": actorID,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExpenseNotFound
	}
	return nil
}
