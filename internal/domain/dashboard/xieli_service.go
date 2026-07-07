package dashboard

import (
	"context"
	"time"
)

// XieLiService owns the Xie Li dashboard KPI business logic. Kept separate from
// HuAiService by design; today the KPIs match, but each dashboard can evolve
// independently. Both reuse the ScopeResolver and AreaRepository helpers.
type XieLiService struct {
	Resolver *ScopeResolver
	Repo     AreaRepository
	now      func() time.Time
}

func NewXieLiService(resolver *ScopeResolver, repo AreaRepository) *XieLiService {
	return &XieLiService{Resolver: resolver, Repo: repo, now: time.Now}
}

func (s *XieLiService) GetDashboard(ctx context.Context, userID int) (*AreaDashboardResponse, error) {
	ids, err := s.Resolver.ResolveXieLiVolunteerIDs(ctx, userID)
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
