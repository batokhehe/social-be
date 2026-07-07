package dashboard

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// ScopeVolunteer is the authenticated volunteer plus the leadership flags used
// for authorization. Flags are COALESCEd to false (NULL = not assigned).
type ScopeVolunteer struct {
	ID            int  `gorm:"column:id"`
	MasterAreaID  int  `gorm:"column:master_area_id"`
	IsHuAiLeader  bool `gorm:"column:is_hu_ai_leader"`
	IsHuAiDeputy  bool `gorm:"column:is_hu_ai_deputy"`
	IsXieLiLeader bool `gorm:"column:is_xie_li_leader"`
	IsXieLiDeputy bool `gorm:"column:is_xie_li_deputy"`
}

// AreaRepository holds the reusable, scope-agnostic aggregate helpers shared by
// the Hu Ai and Xie Li dashboards. Every aggregate takes []volunteerID and
// aggregates with WHERE ... IN (...) -- never one query per volunteer.
type AreaRepository interface {
	// GetScopeVolunteer resolves the authenticated user's volunteer + flags.
	GetScopeVolunteer(ctx context.Context, userID int) (*ScopeVolunteer, error)
	// AreaVolunteerIDs returns all volunteer ids whose master_area is the given
	// area or any descendant of it (recursive over master_areas.parent_id).
	AreaVolunteerIDs(ctx context.Context, masterAreaID int) ([]int, error)

	CountActivities(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error)
	SumVolunteerHours(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error)
	CountActiveVolunteersInRange(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error)
	VolunteerStatusCounts(ctx context.Context, volunteerIDs []int) (active int64, total int64, err error)
	CountDonors(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error)
	SumDonations(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error)
	// SumExpenses sums expenses (status != cancelled) for the scope volunteers,
	// by expense_date over [from, to).
	SumExpenses(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error)
}

type AreaGormRepository struct {
	DB *gorm.DB
}

func NewAreaGormRepository(db *gorm.DB) AreaRepository {
	return &AreaGormRepository{DB: db}
}

func (r *AreaGormRepository) GetScopeVolunteer(ctx context.Context, userID int) (*ScopeVolunteer, error) {
	const q = `
SELECT id, master_area_id,
    COALESCE(is_hu_ai_leader, false)  AS is_hu_ai_leader,
    COALESCE(is_hu_ai_deputy, false)  AS is_hu_ai_deputy,
    COALESCE(is_xie_li_leader, false) AS is_xie_li_leader,
    COALESCE(is_xie_li_deputy, false) AS is_xie_li_deputy
FROM volunteers
WHERE user_id = ? AND deleted_at IS NULL`

	var rows []ScopeVolunteer
	if err := r.DB.WithContext(ctx).Raw(q, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrVolunteerNotFound
	}
	return &rows[0], nil
}

// AreaVolunteerIDs walks the master_areas subtree rooted at masterAreaID and
// returns every volunteer under it. One recursive query, no N+1.
func (r *AreaGormRepository) AreaVolunteerIDs(ctx context.Context, masterAreaID int) ([]int, error) {
	const q = `
WITH RECURSIVE subtree AS (
    SELECT id FROM master_areas WHERE id = ? AND deleted_at IS NULL
    UNION ALL
    SELECT ma.id FROM master_areas ma
    JOIN subtree s ON ma.parent_id = s.id
    WHERE ma.deleted_at IS NULL
)
SELECT v.id
FROM volunteers v
JOIN subtree s ON s.id = v.master_area_id
WHERE v.deleted_at IS NULL`

	var ids []int
	if err := r.DB.WithContext(ctx).Raw(q, masterAreaID).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *AreaGormRepository) CountActivities(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COUNT(DISTINCT ea.event_id)
FROM event_attendances ea
WHERE ea.deleted_at IS NULL
    AND ea.volunteer_id IN ?
    AND ea.checkin_at IS NOT NULL
    AND ea.checkin_at >= ? AND ea.checkin_at < ?`

	var count int64
	err := r.DB.WithContext(ctx).Raw(q, volunteerIDs, from, to).Scan(&count).Error
	return count, err
}

func (r *AreaGormRepository) SumVolunteerHours(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COALESCE(SUM(total_hours), 0)
FROM event_attendances
WHERE deleted_at IS NULL
    AND volunteer_id IN ?
    AND checkout_at IS NOT NULL
    AND checkout_at >= ? AND checkout_at < ?`

	var sum float64
	err := r.DB.WithContext(ctx).Raw(q, volunteerIDs, from, to).Scan(&sum).Error
	return sum, err
}

func (r *AreaGormRepository) CountActiveVolunteersInRange(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COUNT(DISTINCT volunteer_id)
FROM event_attendances
WHERE deleted_at IS NULL
    AND volunteer_id IN ?
    AND checkin_at IS NOT NULL
    AND checkin_at >= ? AND checkin_at < ?`

	var count int64
	err := r.DB.WithContext(ctx).Raw(q, volunteerIDs, from, to).Scan(&count).Error
	return count, err
}

func (r *AreaGormRepository) VolunteerStatusCounts(ctx context.Context, volunteerIDs []int) (int64, int64, error) {
	if len(volunteerIDs) == 0 {
		return 0, 0, nil
	}
	const q = `
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'active') AS active
FROM volunteers
WHERE id IN ? AND deleted_at IS NULL`

	var row struct {
		Total  int64 `gorm:"column:total"`
		Active int64 `gorm:"column:active"`
	}
	if err := r.DB.WithContext(ctx).Raw(q, volunteerIDs).Scan(&row).Error; err != nil {
		return 0, 0, err
	}
	return row.Active, row.Total, nil
}

func (r *AreaGormRepository) CountDonors(ctx context.Context, volunteerIDs []int, from, to time.Time) (int64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COUNT(*)
FROM master_donaturs md
JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
WHERE md.deleted_at IS NULL
    AND mg.deleted_at IS NULL
    AND btrim(mg.volunteer_id) IN ?
    AND md.created_at >= ? AND md.created_at < ?`

	var count int64
	err := r.DB.WithContext(ctx).Raw(q, volunteerKeys(volunteerIDs), from, to).Scan(&count).Error
	return count, err
}

func (r *AreaGormRepository) SumDonations(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COALESCE(SUM(d.amount), 0)
FROM donations d
WHERE d.deleted_at IS NULL
    AND d.type = ?
    AND d.created_at >= ? AND d.created_at < ?
    AND d.donatur_id IN (
        SELECT md.id
        FROM master_donaturs md
        JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
        WHERE md.deleted_at IS NULL
            AND mg.deleted_at IS NULL
            AND btrim(mg.volunteer_id) IN ?
    )`

	var sum float64
	err := r.DB.WithContext(ctx).Raw(q, donationTypeMoney, from, to, volunteerKeys(volunteerIDs)).Scan(&sum).Error
	return sum, err
}

func (r *AreaGormRepository) SumExpenses(ctx context.Context, volunteerIDs []int, from, to time.Time) (float64, error) {
	if len(volunteerIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COALESCE(SUM(amount), 0)
FROM expenses
WHERE deleted_at IS NULL
    AND status <> 'cancelled'
    AND volunteer_id IN ?
    AND expense_date >= ? AND expense_date < ?`

	var sum float64
	err := r.DB.WithContext(ctx).Raw(q, volunteerIDs, from, to).Scan(&sum).Error
	return sum, err
}

// volunteerKeys converts volunteer ids to their text form, since the donor/group
// tables store the volunteer link as volunteers.id in text.
func volunteerKeys(ids []int) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = strconv.Itoa(id)
	}
	return keys
}
