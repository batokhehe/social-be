package dashboard

import (
	"context"
	"time"

	"social-be/internal/pkg/helper"
)

type Service struct {
	Repo Repository
	// now is injectable so month boundaries are deterministic in tests.
	now func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo, now: time.Now}
}

// GetSummary aggregates the four KPI cards. Month boundaries are computed once
// and shared across the per-card queries (4 round trips total).
func (s *Service) GetSummary(ctx context.Context) (*SummaryResponse, error) {
	now := s.now()
	currStart := firstOfMonth(now)
	nextStart := currStart.AddDate(0, 1, 0)
	prevStart := currStart.AddDate(0, -1, 0)

	donation, err := s.Repo.DonationTotals(ctx, prevStart, currStart, nextStart)
	if err != nil {
		return nil, err
	}

	donors, err := s.Repo.ActiveDonorCounts(ctx, prevStart, currStart, nextStart)
	if err != nil {
		return nil, err
	}

	volunteers, err := s.Repo.ActiveVolunteerCounts(ctx, prevStart, currStart, nextStart)
	if err != nil {
		return nil, err
	}

	upcoming, err := s.Repo.UpcomingEventCount(ctx, now)
	if err != nil {
		return nil, err
	}

	return &SummaryResponse{
		TotalDonation:    newMetricCard(donation.Current, donation.Previous),
		ActiveDonors:     newMetricCard(float64(donors.Current), float64(donors.Previous)),
		ActiveVolunteers: newMetricCard(float64(volunteers.Current), float64(volunteers.Previous)),
		UpcomingEvents:   UpcomingEventsCard{Count: upcoming},
	}, nil
}

// GetHome assembles the homepage widgets. Four independent aggregate queries,
// no N+1: ongoing events, latest donations (single join), top volunteers
// (single grouped query), and the impact counts (single round trip).
func (s *Service) GetHome(ctx context.Context) (*HomeResponse, error) {
	ongoing, err := s.Repo.OngoingActivities(ctx, s.now())
	if err != nil {
		return nil, err
	}

	latest, err := s.Repo.LatestDonations(ctx)
	if err != nil {
		return nil, err
	}

	top, err := s.Repo.TopVolunteers(ctx)
	if err != nil {
		return nil, err
	}

	impact, err := s.Repo.ImpactSummary(ctx, s.now())
	if err != nil {
		return nil, err
	}

	return &HomeResponse{
		OngoingActivities: orEmpty(ongoing),
		LatestDonations:   orEmpty(latest),
		TopVolunteers:     orEmpty(top),
		ImpactSummary:     impact,
	}, nil
}

// GetDonationsByCategory returns the current-month donation-by-category pie
// chart: one slice per category with its summed amount, plus the grand total.
// Percentages are left to the frontend. The grand total is summed from the
// slices so no extra query is needed.
func (s *Service) GetDonationsByCategory(ctx context.Context) (*DonationByCategoryResponse, error) {
	currStart := firstOfMonth(s.now())
	nextStart := currStart.AddDate(0, 1, 0)

	slices, err := s.Repo.DonationsByCategory(ctx, currStart, nextStart)
	if err != nil {
		return nil, err
	}

	var total float64
	for _, slice := range slices {
		total += slice.Total
	}

	return &DonationByCategoryResponse{
		TotalAmount: total,
		Categories:  orEmpty(slices),
	}, nil
}

// orEmpty guarantees a non-nil slice so the JSON response serializes [] (not null).
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func newMetricCard(current, previous float64) MetricCard {
	return MetricCard{
		Current:    current,
		Previous:   previous,
		Percentage: helper.CalculatePercentage(current, previous),
	}
}

// firstOfMonth returns midnight on the first day of t's month, in t's location.
func firstOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}
