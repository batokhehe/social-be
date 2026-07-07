package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockAreaRepo lets tests drive scope + aggregates. Scalar aggregates are keyed
// by the month start (YYYY-MM) so the current-month/summary mapping is testable.
type mockAreaRepo struct {
	scope    *ScopeVolunteer
	scopeErr error

	areaIDs   []int
	gotAreaID int

	activitiesByMonth map[string]int64
	hoursByMonth      map[string]float64
	activeInRange     map[string]int64
	donorsByMonth     map[string]int64
	donationsByMonth  map[string]float64
	expensesByMonth   map[string]float64
	statusActive      int64
	statusTotal       int64

	gotIDs []int
}

func (m *mockAreaRepo) GetScopeVolunteer(ctx context.Context, userID int) (*ScopeVolunteer, error) {
	if m.scopeErr != nil {
		return nil, m.scopeErr
	}
	return m.scope, nil
}
func (m *mockAreaRepo) AreaVolunteerIDs(ctx context.Context, masterAreaID int) ([]int, error) {
	m.gotAreaID = masterAreaID
	return m.areaIDs, nil
}
func (m *mockAreaRepo) CountActivities(ctx context.Context, ids []int, from, to time.Time) (int64, error) {
	m.gotIDs = ids
	return m.activitiesByMonth[from.Format("2006-01")], nil
}
func (m *mockAreaRepo) SumVolunteerHours(ctx context.Context, ids []int, from, to time.Time) (float64, error) {
	return m.hoursByMonth[from.Format("2006-01")], nil
}
func (m *mockAreaRepo) CountActiveVolunteersInRange(ctx context.Context, ids []int, from, to time.Time) (int64, error) {
	return m.activeInRange[from.Format("2006-01")], nil
}
func (m *mockAreaRepo) VolunteerStatusCounts(ctx context.Context, ids []int) (int64, int64, error) {
	return m.statusActive, m.statusTotal, nil
}
func (m *mockAreaRepo) CountDonors(ctx context.Context, ids []int, from, to time.Time) (int64, error) {
	return m.donorsByMonth[from.Format("2006-01")], nil
}
func (m *mockAreaRepo) SumExpenses(ctx context.Context, ids []int, from, to time.Time) (float64, error) {
	return m.expensesByMonth[from.Format("2006-01")], nil
}
func (m *mockAreaRepo) SumDonations(ctx context.Context, ids []int, from, to time.Time) (float64, error) {
	return m.donationsByMonth[from.Format("2006-01")], nil
}

// --- scope resolver authorization ---

func TestResolveHuAi_ForbiddenWithoutFlag(t *testing.T) {
	repo := &mockAreaRepo{scope: &ScopeVolunteer{ID: 1, MasterAreaID: 10, IsXieLiLeader: true}} // xie li only
	resolver := NewScopeResolver(repo)

	_, err := resolver.ResolveHuAiVolunteerIDs(context.Background(), 99)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestResolveHuAi_AllowedForDeputy(t *testing.T) {
	repo := &mockAreaRepo{
		scope:   &ScopeVolunteer{ID: 1, MasterAreaID: 10, IsHuAiDeputy: true},
		areaIDs: []int{1, 2, 3},
	}
	resolver := NewScopeResolver(repo)

	ids, err := resolver.ResolveHuAiVolunteerIDs(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || repo.gotAreaID != 10 {
		t.Fatalf("ids=%v areaID=%d, want 3 ids from area 10", ids, repo.gotAreaID)
	}
}

func TestResolveXieLi_ForbiddenWithHuAiFlagOnly(t *testing.T) {
	repo := &mockAreaRepo{scope: &ScopeVolunteer{ID: 1, MasterAreaID: 10, IsHuAiLeader: true}}
	resolver := NewScopeResolver(repo)

	if _, err := resolver.ResolveXieLiVolunteerIDs(context.Background(), 99); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// --- service assembly ---

func TestHuAiService_GetDashboard(t *testing.T) {
	repo := &mockAreaRepo{
		scope:   &ScopeVolunteer{ID: 1, MasterAreaID: 10, IsHuAiLeader: true},
		areaIDs: []int{1, 2, 3},
		activitiesByMonth: map[string]int64{
			"2026-06": 12, // current month
			"2026-04": 5,
		},
		hoursByMonth:     map[string]float64{"2026-06": 96.5},
		activeInRange:    map[string]int64{"2026-06": 8},
		donorsByMonth:    map[string]int64{"2026-06": 4},
		donationsByMonth: map[string]float64{"2026-06": 5000000},
		expensesByMonth:  map[string]float64{"2026-06": 1200000, "2026-04": 300000},
		statusActive:     18,
		statusTotal:      24,
	}
	svc := &HuAiService{Resolver: NewScopeResolver(repo), Repo: repo, now: func() time.Time {
		return time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)
	}}

	got, err := svc.GetDashboard(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Summary = current month (June) + status-based active/total.
	if got.Summary.TotalActivities != 12 || got.Summary.TotalVolunteerHours != 96.5 {
		t.Fatalf("summary current-month wrong: %+v", got.Summary)
	}
	if got.Summary.ActiveVolunteers != 18 || got.Summary.TotalVolunteers != 24 {
		t.Fatalf("summary active/total = %d/%d, want 18/24", got.Summary.ActiveVolunteers, got.Summary.TotalVolunteers)
	}
	if got.Summary.TotalDonors != 4 || got.Summary.TotalDonations != 5000000 {
		t.Fatalf("summary donors/donations wrong: %+v", got.Summary)
	}
	// Expenses now resolved from the Expense module (current month = June).
	if got.Summary.TotalExpenses != 1200000 {
		t.Fatalf("summary expenses = %v, want 1200000", got.Summary.TotalExpenses)
	}
	if got.MonthlyChart[5].TotalExpenses != 1200000 || got.MonthlyChart[3].TotalExpenses != 300000 {
		t.Fatalf("chart expenses: jun=%v apr=%v", got.MonthlyChart[5].TotalExpenses, got.MonthlyChart[3].TotalExpenses)
	}

	// Chart: 6 months oldest-first, Jan..Jun.
	if len(got.MonthlyChart) != 6 {
		t.Fatalf("chart len = %d, want 6", len(got.MonthlyChart))
	}
	wantMonths := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	for i, m := range wantMonths {
		if got.MonthlyChart[i].Month != m {
			t.Fatalf("chart[%d].month = %q, want %q", i, got.MonthlyChart[i].Month, m)
		}
	}
	if got.MonthlyChart[3].Activities != 5 { // April
		t.Fatalf("April activities = %d, want 5", got.MonthlyChart[3].Activities)
	}
	if got.MonthlyChart[1].Activities != 0 { // Feb zero-filled
		t.Fatalf("Feb activities = %d, want 0", got.MonthlyChart[1].Activities)
	}
	if got.MonthlyChart[5].Activities != 12 || got.MonthlyChart[5].TotalDonations != 5000000 {
		t.Fatalf("June bucket wrong: %+v", got.MonthlyChart[5])
	}
	// Scope reached the aggregates.
	if len(repo.gotIDs) != 3 {
		t.Fatalf("aggregates got ids=%v, want 3", repo.gotIDs)
	}
}

func TestXieLiService_ForbiddenPropagates(t *testing.T) {
	repo := &mockAreaRepo{scope: &ScopeVolunteer{ID: 1, MasterAreaID: 10}} // no flags
	svc := NewXieLiService(NewScopeResolver(repo), repo)

	if _, err := svc.GetDashboard(context.Background(), 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
