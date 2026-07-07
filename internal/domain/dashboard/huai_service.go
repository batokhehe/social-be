package dashboard

import (
	"context"
	"time"
)

// HuAiService owns the Hu Ai dashboard KPI business logic. It is intentionally
// separate from XieLiService (they may diverge); both reuse the ScopeResolver
// and the AreaRepository aggregate helpers.
type HuAiService struct {
	Resolver *ScopeResolver
	Repo     AreaRepository
	now      func() time.Time
}

func NewHuAiService(resolver *ScopeResolver, repo AreaRepository) *HuAiService {
	return &HuAiService{Resolver: resolver, Repo: repo, now: time.Now}
}

func (s *HuAiService) GetDashboard(ctx context.Context, userID int) (*AreaDashboardResponse, error) {
	ids, err := s.Resolver.ResolveHuAiVolunteerIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	currStart := firstOfMonth(now)
	sixStart := currStart.AddDate(0, -5, 0)

	chart := make([]AreaMonthlyStat, 0, 6)
	var current AreaMonthlyStat
	for i := 0; i < 6; i++ {
		monthStart := sixStart.AddDate(0, i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		activities, err := s.Repo.CountActivities(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		hours, err := s.Repo.SumVolunteerHours(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		activeVols, err := s.Repo.CountActiveVolunteersInRange(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		donors, err := s.Repo.CountDonors(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		donations, err := s.Repo.SumDonations(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		expenses, err := s.Repo.SumExpenses(ctx, ids, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}

		stat := AreaMonthlyStat{
			Month:            monthStart.Format("Jan"),
			Activities:       activities,
			VolunteerHours:   hours,
			ActiveVolunteers: activeVols,
			TotalDonors:      donors,
			TotalDonations:   donations,
			TotalExpenses:    expenses,
		}
		chart = append(chart, stat)
		if i == 5 {
			current = stat
		}
	}

	activeStatus, total, err := s.Repo.VolunteerStatusCounts(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &AreaDashboardResponse{
		Summary: AreaSummary{
			TotalActivities:     current.Activities,
			TotalVolunteerHours: current.VolunteerHours,
			ActiveVolunteers:    activeStatus,
			TotalVolunteers:     total,
			TotalDonors:         current.TotalDonors,
			TotalDonations:      current.TotalDonations,
			TotalExpenses:       current.TotalExpenses,
		},
		MonthlyChart: chart,
	}, nil
}
