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
