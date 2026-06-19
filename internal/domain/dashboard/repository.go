package dashboard

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// donationTypeMoney mirrors donation.DonationTypeMoney. It is duplicated here as
// a private constant so the dashboard read-model does not depend on the donation
// domain package.
const donationTypeMoney = 0

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
	ImpactSummary(ctx context.Context, now time.Time) (ImpactSummary, error)

	// DonationsByCategory returns, per donation category, the summed money-donation
	// amount for the current month [currStart, nextStart) -- a single grouped
	// query (no N+1). Percentages are left to the frontend.
	DonationsByCategory(ctx context.Context, currStart, nextStart time.Time) ([]DonationCategorySlice, error)
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
func (r *GormRepository) ImpactSummary(ctx context.Context, now time.Time) (ImpactSummary, error) {
	const q = `
SELECT
    (SELECT COUNT(*) FROM volunteers WHERE status = 'active' AND deleted_at IS NULL) AS active_volunteers,
    (SELECT COUNT(*) FROM events WHERE end_at < ? AND deleted_at IS NULL) AS completed_activities`

	var out ImpactSummary
	if err := r.DB.WithContext(ctx).Raw(q, now).Scan(&out).Error; err != nil {
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
