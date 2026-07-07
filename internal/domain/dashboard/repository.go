package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

// donationTypeMoney mirrors donation.DonationTypeMoney. It is duplicated here as
// a private constant so the dashboard read-model does not depend on the donation
// domain package.
const donationTypeMoney = 0

// Category activity ids used by the personal volunteer dashboard.
const (
	categoryActivityMission    = 2 // survey, case meeting, ...
	categoryActivityPhilosophy = 4
)

// ErrVolunteerNotFound is returned when the authenticated user has no volunteer
// profile (so the personal dashboard cannot be built).
var ErrVolunteerNotFound = errors.New("volunteer profile not found")

// monthlyActivityRow is one grouped month of attendance hours.
type monthlyActivityRow struct {
	YM              string  `gorm:"column:ym"`
	ActivityHours   float64 `gorm:"column:activity_hours"`
	PhilosophyHours float64 `gorm:"column:philosophy_hours"`
	MissionHours    float64 `gorm:"column:mission_hours"`
}

// monthlyDonationRow is one grouped month of money donations for the volunteer.
type monthlyDonationRow struct {
	YM                string  `gorm:"column:ym"`
	MyDonation        float64 `gorm:"column:my_donation"`
	CollectedDonation float64 `gorm:"column:collected_donation"`
}

// monthlyDonorRow is one grouped month of donors acquired under the volunteer's
// groups (by master_donaturs.created_at).
type monthlyDonorRow struct {
	YM         string `gorm:"column:ym"`
	DonorCount int64  `gorm:"column:donor_count"`
}

// monthlyExpenseRow is one grouped month of expenses (by expenses.expense_date).
type monthlyExpenseRow struct {
	YM            string  `gorm:"column:ym"`
	ExpenseAmount float64 `gorm:"column:expense_amount"`
}

// AmountPair holds a current-month / previous-month monetary aggregate.
type AmountPair struct {
	Current  float64
	Previous float64
}

// CountPair holds a current-month / previous-month row count.
type CountPair struct {
	Current  int64
	Previous int64
}

// Repository exposes aggregate-only reads for the dashboard. Each method issues
// a single round trip and returns only aggregated scalars (never full rows).
type Repository interface {
	// DonationTotals returns SUM(amount) for money donations (type = 0, not
	// soft-deleted) for the current and previous month in one query.
	DonationTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error)
	// ExpenseTotals returns org-wide expense sums (status != cancelled) for the
	// current and previous month (by expense_date) in one query.
	ExpenseTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error)
	// ActiveDonorCounts returns COUNT(*) of active master_donaturs created in the
	// current and previous month in one query.
	ActiveDonorCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error)
	// ActiveVolunteerCounts returns COUNT(*) of active volunteers created in the
	// current and previous month in one query.
	ActiveVolunteerCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error)
	// UpcomingEventCount returns COUNT(*) of active, not-soft-deleted events
	// starting at or after now.
	UpcomingEventCount(ctx context.Context, now time.Time) (int64, error)

	// OngoingActivities returns up to 5 active events whose [start_at, end_at]
	// window contains now, ordered by start_at ASC.
	OngoingActivities(ctx context.Context, now time.Time) ([]OngoingActivity, error)
	// LatestDonations returns the 5 most recent (not soft-deleted) donations with
	// the donor name resolved in a single join.
	LatestDonations(ctx context.Context) ([]LatestDonation, error)
	// TopVolunteers returns the top 5 volunteers by total contribution hours,
	// aggregated from completed check-in/check-out pairs in one grouped query.
	TopVolunteers(ctx context.Context) ([]TopVolunteer, error)
	// ImpactSummary returns active-volunteer and completed-activity counts in a
	// single round trip. completed activities = events that have already ended
	// (end_at < now), hence now is passed for a consistent, testable clock.
	// ImpactSummary returns active volunteers, completed activities, and the
	// current-month [currStart, nextStart) expense total in one query.
	ImpactSummary(ctx context.Context, now, currStart, nextStart time.Time) (ImpactSummary, error)

	// DonationsByCategory returns, per donation category, the summed money-donation
	// amount for the current month [currStart, nextStart) -- a single grouped
	// query (no N+1). Percentages are left to the frontend.
	DonationsByCategory(ctx context.Context, currStart, nextStart time.Time) ([]DonationCategorySlice, error)

	// --- personal volunteer dashboard ---

	// ResolveVolunteerID maps the authenticated user to their volunteers.id.
	// Returns ErrVolunteerNotFound when the user has no volunteer profile.
	ResolveVolunteerID(ctx context.Context, userID int) (int, error)
	// MonthlyActivityStats returns attendance hours (total, philosophy, mission)
	// grouped by month over [from, to), for the volunteer (by volunteers.id).
	MonthlyActivityStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyActivityRow, error)
	// MonthlyExpenseStats returns the volunteer's own expense totals (status !=
	// cancelled) grouped by month over [from, to).
	MonthlyExpenseStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyExpenseRow, error)
	// MonthlyDonationStats returns money-donation totals (my own / collected via
	// my groups) grouped by month over [from, to). volunteerKey is the
	// volunteers.id as text (how the donor/group tables store the link).
	MonthlyDonationStats(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonationRow, error)
	// CountMyDonors counts donors in groups owned by the volunteer (all-time).
	CountMyDonors(ctx context.Context, volunteerKey string) (int64, error)
	// MonthlyDonorCounts returns the number of donors acquired (by created_at)
	// under the volunteer's groups, grouped by month over [from, to).
	MonthlyDonorCounts(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonorRow, error)
}

type GormRepository struct {
	DB *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{DB: db}
}

// DonationTotals: conditional aggregation computes both months in a single scan.
// The outer created_at bound lets the planner restrict the scan to the two
// months of interest.
func (r *GormRepository) DonationTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error) {
	const q = `
SELECT
    COALESCE(SUM(amount) FILTER (WHERE created_at >= ? AND created_at < ?), 0) AS current,
    COALESCE(SUM(amount) FILTER (WHERE created_at >= ? AND created_at < ?), 0) AS previous
FROM donations
WHERE type = ?
    AND deleted_at IS NULL
    AND created_at >= ? AND created_at < ?`

	var out AmountPair
	if err := r.DB.WithContext(ctx).Raw(q,
		currStart, nextStart,
		prevStart, currStart,
		donationTypeMoney,
		prevStart, nextStart,
	).Scan(&out).Error; err != nil {
		return AmountPair{}, err
	}
	return out, nil
}

// ExpenseTotals mirrors DonationTotals on the expenses table (month by
// expense_date, cancelled excluded).
func (r *GormRepository) ExpenseTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error) {
	const q = `
SELECT
    COALESCE(SUM(amount) FILTER (WHERE expense_date >= ? AND expense_date < ?), 0) AS current,
    COALESCE(SUM(amount) FILTER (WHERE expense_date >= ? AND expense_date < ?), 0) AS previous
FROM expenses
WHERE status <> 'cancelled'
    AND deleted_at IS NULL
    AND expense_date >= ? AND expense_date < ?`

	var out AmountPair
	if err := r.DB.WithContext(ctx).Raw(q,
		currStart, nextStart,
		prevStart, currStart,
		prevStart, nextStart,
	).Scan(&out).Error; err != nil {
		return AmountPair{}, err
	}
	return out, nil
}

func (r *GormRepository) ActiveDonorCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error) {
	const q = `
SELECT
    COUNT(*) FILTER (WHERE created_at >= ? AND created_at < ?) AS current,
    COUNT(*) FILTER (WHERE created_at >= ? AND created_at < ?) AS previous
FROM master_donaturs
WHERE status = 'active'
    AND created_at >= ? AND created_at < ?`

	var out CountPair
	if err := r.DB.WithContext(ctx).Raw(q,
		currStart, nextStart,
		prevStart, currStart,
		prevStart, nextStart,
	).Scan(&out).Error; err != nil {
		return CountPair{}, err
	}
	return out, nil
}

func (r *GormRepository) ActiveVolunteerCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error) {
	const q = `
SELECT
    COUNT(*) FILTER (WHERE created_at >= ? AND created_at < ?) AS current,
    COUNT(*) FILTER (WHERE created_at >= ? AND created_at < ?) AS previous
FROM volunteers
WHERE status = 'active'
    AND created_at >= ? AND created_at < ?`

	var out CountPair
	if err := r.DB.WithContext(ctx).Raw(q,
		currStart, nextStart,
		prevStart, currStart,
		prevStart, nextStart,
	).Scan(&out).Error; err != nil {
		return CountPair{}, err
	}
	return out, nil
}

func (r *GormRepository) UpcomingEventCount(ctx context.Context, now time.Time) (int64, error) {
	const q = `
SELECT COUNT(*)
FROM events
WHERE status = 'active'
    AND start_at >= ?
    AND deleted_at IS NULL`

	var count int64
	if err := r.DB.WithContext(ctx).Raw(q, now).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// OngoingActivities: now is passed (rather than CURRENT_TIMESTAMP) so the clock
// is consistent with the rest of the dashboard and testable.
func (r *GormRepository) OngoingActivities(ctx context.Context, now time.Time) ([]OngoingActivity, error) {
	const q = `
SELECT id, name, start_at, end_at, status
FROM events
WHERE status = 'active'
    AND ? BETWEEN start_at AND end_at
    AND deleted_at IS NULL
ORDER BY start_at ASC
LIMIT 5`

	var out []OngoingActivity
	if err := r.DB.WithContext(ctx).Raw(q, now).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// LatestDonations resolves the donor name via a single LEFT JOIN (so donations
// whose donor was soft-deleted still appear, with an empty name).
func (r *GormRepository) LatestDonations(ctx context.Context) ([]LatestDonation, error) {
	const q = `
SELECT d.id,
       COALESCE(md.name, '') AS donator_name,
       d.amount,
       d.created_at
FROM donations d
LEFT JOIN master_donaturs md ON md.id = d.donatur_id
WHERE d.deleted_at IS NULL
ORDER BY d.created_at DESC
LIMIT 5`

	var out []LatestDonation
	if err := r.DB.WithContext(ctx).Raw(q).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// TopVolunteers aggregates contribution hours from completed attendance pairs.
// total_hours = SUM(checkout_at - checkin_at) converted to hours, rounded to 1dp.
func (r *GormRepository) TopVolunteers(ctx context.Context) ([]TopVolunteer, error) {
	const q = `
SELECT a.volunteer_id,
       v.indonesian_name AS name,
       ROUND(EXTRACT(EPOCH FROM SUM(a.checkout_at - a.checkin_at)) / 3600.0, 1) AS total_hours
FROM event_attendances a
JOIN volunteers v ON v.id = a.volunteer_id
WHERE a.checkin_at IS NOT NULL
    AND a.checkout_at IS NOT NULL
    AND a.deleted_at IS NULL
GROUP BY a.volunteer_id, v.indonesian_name
ORDER BY total_hours DESC
LIMIT 5`

	var out []TopVolunteer
	if err := r.DB.WithContext(ctx).Raw(q).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ImpactSummary returns both counts in one round trip via scalar subqueries.
//
// completed_activities counts events that have already finished (end_at < now).
// The application's event lifecycle only uses status 'active'/'inactive' and
// never writes 'completed', so completion is derived from the schedule, not a
// status flag. Soft-deleted events are excluded.
func (r *GormRepository) ImpactSummary(ctx context.Context, now, currStart, nextStart time.Time) (ImpactSummary, error) {
	const q = `
SELECT
    (SELECT COUNT(*) FROM volunteers WHERE status = 'active' AND deleted_at IS NULL) AS active_volunteers,
    (SELECT COUNT(*) FROM events WHERE end_at < @now AND deleted_at IS NULL) AS completed_activities,
    (SELECT COALESCE(SUM(amount), 0) FROM expenses
        WHERE status <> 'cancelled' AND deleted_at IS NULL
        AND expense_date >= @from AND expense_date < @to) AS current_month_expense`

	var out ImpactSummary
	if err := r.DB.WithContext(ctx).Raw(q,
		sql.Named("now", now),
		sql.Named("from", currStart),
		sql.Named("to", nextStart),
	).Scan(&out).Error; err != nil {
		return ImpactSummary{}, err
	}
	return out, nil
}

// DonationsByCategory aggregates current-month money donations (type = 0, not
// soft-deleted, created_at in [currStart, nextStart)) by category, summing the
// amount per category in one grouped query. Categories with no donations this
// month are naturally absent (no empty slices). A category that was later
// soft-deleted still appears, because the donation rows reference it by id.
func (r *GormRepository) DonationsByCategory(ctx context.Context, currStart, nextStart time.Time) ([]DonationCategorySlice, error) {
	const q = `
SELECT
    dc.id   AS category_id,
    dc.name AS category_name,
    COALESCE(SUM(d.amount), 0) AS total
FROM donations d
JOIN master_donation_categories dc ON dc.id = d.donation_category_id
WHERE d.type = ?
    AND d.deleted_at IS NULL
    AND d.created_at >= ? AND d.created_at < ?
GROUP BY dc.id, dc.name
ORDER BY total DESC`

	var out []DonationCategorySlice
	if err := r.DB.WithContext(ctx).Raw(q, donationTypeMoney, currStart, nextStart).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *GormRepository) ResolveVolunteerID(ctx context.Context, userID int) (int, error) {
	var id int
	if err := r.DB.WithContext(ctx).
		Raw(`SELECT id FROM volunteers WHERE user_id = ? AND deleted_at IS NULL`, userID).
		Scan(&id).Error; err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, ErrVolunteerNotFound
	}
	return id, nil
}

// MonthlyActivityStats: one grouped query over attendance hours. total_hours is
// reused (not recomputed). Only checked-out rows have total_hours; the explicit
// checkout_at IS NOT NULL keeps the month bucketing well-defined.
func (r *GormRepository) MonthlyActivityStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyActivityRow, error) {
	const q = `
SELECT
    to_char(date_trunc('month', ea.checkout_at), 'YYYY-MM') AS ym,
    COALESCE(SUM(ea.total_hours), 0) AS activity_hours,
    COALESCE(SUM(ea.total_hours) FILTER (WHERE e.category_activity_id = @philosophy), 0) AS philosophy_hours,
    COALESCE(SUM(ea.total_hours) FILTER (WHERE e.category_activity_id = @mission), 0) AS mission_hours
FROM event_attendances ea
JOIN events e ON e.id = ea.event_id
WHERE ea.volunteer_id = @vol
    AND ea.deleted_at IS NULL
    AND ea.checkout_at IS NOT NULL
    AND ea.checkout_at >= @from AND ea.checkout_at < @to
GROUP BY 1`

	var rows []monthlyActivityRow
	err := r.DB.WithContext(ctx).Raw(q,
		sql.Named("philosophy", categoryActivityPhilosophy),
		sql.Named("mission", categoryActivityMission),
		sql.Named("vol", volunteerID),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&rows).Error
	return rows, err
}

// MonthlyExpenseStats: one grouped query of the volunteer's own expenses by
// expense_date month (cancelled excluded).
func (r *GormRepository) MonthlyExpenseStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyExpenseRow, error) {
	const q = `
SELECT
    to_char(date_trunc('month', expense_date), 'YYYY-MM') AS ym,
    COALESCE(SUM(amount), 0) AS expense_amount
FROM expenses
WHERE deleted_at IS NULL
    AND status <> 'cancelled'
    AND volunteer_id = ?
    AND expense_date >= ? AND expense_date < ?
GROUP BY 1`

	var rows []monthlyExpenseRow
	err := r.DB.WithContext(ctx).Raw(q, volunteerID, from, to).Scan(&rows).Error
	return rows, err
}

// MonthlyDonationStats: one grouped query computing both "my donations" (I am
// the donor: master_donaturs.id_vis_volunteer = me) and "collected donations"
// (donors in groups I own: master_donatur_groups.volunteer_id = me). The donor
// id columns store volunteers.id as text, hence btrim(...) = @vol. The outer OR
// restricts the scan to the relevant donations before the FILTER split.
func (r *GormRepository) MonthlyDonationStats(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonationRow, error) {
	const q = `
SELECT
    to_char(date_trunc('month', d.created_at), 'YYYY-MM') AS ym,
    COALESCE(SUM(d.amount) FILTER (WHERE d.donatur_id IN (
        SELECT id FROM master_donaturs
        WHERE deleted_at IS NULL AND id_vis_volunteer = @vol
    )), 0) AS my_donation,
    COALESCE(SUM(d.amount) FILTER (WHERE d.donatur_id IN (
        SELECT md.id FROM master_donaturs md
        JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
        WHERE md.deleted_at IS NULL AND mg.deleted_at IS NULL AND mg.volunteer_id = @vol
    )), 0) AS collected_donation
FROM donations d
WHERE d.deleted_at IS NULL
    AND d.type = @money
    AND d.created_at >= @from AND d.created_at < @to
    AND (
        d.donatur_id IN (
            SELECT id FROM master_donaturs
            WHERE deleted_at IS NULL AND id_vis_volunteer = @vol
        )
        OR d.donatur_id IN (
            SELECT md.id FROM master_donaturs md
            JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
            WHERE md.deleted_at IS NULL AND mg.deleted_at IS NULL AND mg.volunteer_id = @vol
        )
    )
GROUP BY 1`

	var rows []monthlyDonationRow
	err := r.DB.WithContext(ctx).Raw(q,
		sql.Named("vol", volunteerKey),
		sql.Named("money", donationTypeMoney),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&rows).Error
	return rows, err
}

// CountMyDonors counts donors belonging to groups the volunteer owns
// (volunteer -> donatur_group -> master_donaturs). All-time.
func (r *GormRepository) CountMyDonors(ctx context.Context, volunteerKey string) (int64, error) {
	const q = `
SELECT COUNT(*)
FROM master_donaturs md
JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
WHERE md.deleted_at IS NULL
    AND mg.deleted_at IS NULL
    AND mg.volunteer_id = @vol`

	var count int64
	err := r.DB.WithContext(ctx).Raw(q, sql.Named("vol", volunteerKey)).Scan(&count).Error
	return count, err
}

// MonthlyDonorCounts groups donors by the month they were created, restricted to
// the volunteer's groups. Uses the same group-ownership path as CountMyDonors.
func (r *GormRepository) MonthlyDonorCounts(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonorRow, error) {
	const q = `
SELECT
    to_char(date_trunc('month', md.created_at), 'YYYY-MM') AS ym,
    COUNT(*) AS donor_count
FROM master_donaturs md
JOIN master_donatur_groups mg ON mg.id = md.id_group_donatur
WHERE md.deleted_at IS NULL
    AND mg.deleted_at IS NULL
    AND mg.volunteer_id = @vol
    AND md.created_at >= @from AND md.created_at < @to
GROUP BY 1`

	var rows []monthlyDonorRow
	err := r.DB.WithContext(ctx).Raw(q,
		sql.Named("vol", volunteerKey),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&rows).Error
	return rows, err
}
