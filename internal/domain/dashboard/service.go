package dashboard

import (
	"context"
	"strconv"
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

	expense, err := s.Repo.ExpenseTotals(ctx, prevStart, currStart, nextStart)
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
		TotalExpense:     newMetricCard(expense.Current, expense.Previous),
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

	now := s.now()
	currStart := firstOfMonth(now)
	nextStart := currStart.AddDate(0, 1, 0)
	sixStart := currStart.AddDate(0, -5, 0) // 6 months including the current one

	impact, err := s.Repo.ImpactSummary(ctx, now, currStart, nextStart)
	if err != nil {
		return nil, err
	}

	trendRows, err := s.Repo.HomeTrends(ctx, sixStart, nextStart)
	if err != nil {
		return nil, err
	}
	byMonth := map[string]map[string]float64{
		trendSourceVolunteer: {},
		trendSourceDonor:     {},
		trendSourceDonation:  {},
	}
	for _, row := range trendRows {
		if bucket, ok := byMonth[row.Source]; ok {
			bucket[row.YM] = row.Total
		}
	}

	return &HomeResponse{
		OngoingActivities: orEmpty(ongoing),
		LatestDonations:   orEmpty(latest),
		TopVolunteers:     orEmpty(top),
		ImpactSummary:     impact,
		VolunteerTrend:    buildTrend(sixStart, byMonth[trendSourceVolunteer]),
		DonorTrend:        buildTrend(sixStart, byMonth[trendSourceDonor]),
		DonationTrend:     buildTrend(sixStart, byMonth[trendSourceDonation]),
	}, nil
}

// buildTrend produces exactly 6 ascending monthly points starting at sixStart,
// filling months without data with zero.
func buildTrend(sixStart time.Time, totals map[string]float64) []TrendPoint {
	points := make([]TrendPoint, 0, 6)
	for i := 0; i < 6; i++ {
		key := sixStart.AddDate(0, i, 0).Format("2006-01")
		points = append(points, TrendPoint{Month: key, Total: totals[key]})
	}
	return points
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

// GetVolunteerDashboard builds the personal dashboard for the authenticated
// user. The scalar cards are the current month; the chart is the last 6 months
// (oldest first). Total queries: resolve(1) + activity(1) + donation(1) +
// donors(1) = 4 aggregate round trips, all server-timezone month-bounded.
func (s *Service) GetVolunteerDashboard(ctx context.Context, userID int) (*VolunteerDashboardResponse, error) {
	volunteerID, err := s.Repo.ResolveVolunteerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Donor/group tables store the volunteer link as volunteers.id in text form.
	volunteerKey := strconv.Itoa(volunteerID)

	now := s.now()
	currStart := firstOfMonth(now)
	nextStart := currStart.AddDate(0, 1, 0)
	sixStart := currStart.AddDate(0, -5, 0) // 6 months including the current one

	activity, err := s.Repo.MonthlyActivityStats(ctx, volunteerID, sixStart, nextStart)
	if err != nil {
		return nil, err
	}
	donation, err := s.Repo.MonthlyDonationStats(ctx, volunteerKey, sixStart, nextStart)
	if err != nil {
		return nil, err
	}
	donors, err := s.Repo.CountMyDonors(ctx, volunteerKey)
	if err != nil {
		return nil, err
	}

	donorByMonthRows, err := s.Repo.MonthlyDonorCounts(ctx, volunteerKey, sixStart, nextStart)
	if err != nil {
		return nil, err
	}

	expenseRows, err := s.Repo.MonthlyExpenseStats(ctx, volunteerID, sixStart, nextStart)
	if err != nil {
		return nil, err
	}

	actByMonth := make(map[string]monthlyActivityRow, len(activity))
	for _, row := range activity {
		actByMonth[row.YM] = row
	}
	donByMonth := make(map[string]monthlyDonationRow, len(donation))
	for _, row := range donation {
		donByMonth[row.YM] = row
	}
	donorByMonth := make(map[string]monthlyDonorRow, len(donorByMonthRows))
	for _, row := range donorByMonthRows {
		donorByMonth[row.YM] = row
	}
	expenseByMonth := make(map[string]monthlyExpenseRow, len(expenseRows))
	for _, row := range expenseRows {
		expenseByMonth[row.YM] = row
	}

	currentKey := currStart.Format("2006-01")
	chart := make([]VolunteerMonthlyStat, 0, 6)
	var current VolunteerMonthlyStat
	for i := 0; i < 6; i++ {
		monthStart := sixStart.AddDate(0, i, 0)
		key := monthStart.Format("2006-01")
		act := actByMonth[key] // zero value when the month has no rows
		don := donByMonth[key]
		stat := VolunteerMonthlyStat{
			Month:             monthStart.Format("Jan"),
			ActivityHours:     act.ActivityHours,
			PhilosophyHours:   act.PhilosophyHours,
			MissionHours:      act.MissionHours,
			MyDonation:        don.MyDonation,
			CollectedDonation: don.CollectedDonation,
			DonorCount:        donorByMonth[key].DonorCount,
			ExpenseAmount:     expenseByMonth[key].ExpenseAmount,
		}
		chart = append(chart, stat)
		if key == currentKey {
			current = stat
		}
	}

	return &VolunteerDashboardResponse{
		TotalActivityHours:     current.ActivityHours,
		PhilosophyHours:        current.PhilosophyHours,
		MissionHours:           current.MissionHours,
		TotalDonors:            donors,
		TotalMyDonation:        current.MyDonation,
		TotalCollectedDonation: current.CollectedDonation,
		TotalMyExpense:         current.ExpenseAmount,
		MonthlyChart:           chart,
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
