package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepo struct {
	donation    AmountPair
	expense     AmountPair
	donors      CountPair
	volunteers  CountPair
	upcoming    int64
	expenseRows []monthlyExpenseRow

	// home widgets
	ongoing    []OngoingActivity
	latest     []LatestDonation
	top        []TopVolunteer
	impact     ImpactSummary
	byCategory []DonationCategorySlice

	// captured arguments for boundary assertions
	gotPrev, gotCurr, gotNext time.Time
	gotNow                    time.Time
	gotOngoingNow             time.Time
	gotImpactNow              time.Time
	gotCatCurr, gotCatNext    time.Time

	// volunteer dashboard
	volunteerID          int
	volunteerErr         error
	activityRows         []monthlyActivityRow
	donationRows         []monthlyDonationRow
	donorRows            []monthlyDonorRow
	donorCount           int64
	gotUserID            int
	gotVolID             int
	gotVolKey            string
	gotActFrom, gotActTo time.Time
}

func (m *mockRepo) ExpenseTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error) {
	return m.expense, nil
}

func (m *mockRepo) MonthlyExpenseStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyExpenseRow, error) {
	return m.expenseRows, nil
}

func (m *mockRepo) DonationTotals(ctx context.Context, prevStart, currStart, nextStart time.Time) (AmountPair, error) {
	m.gotPrev, m.gotCurr, m.gotNext = prevStart, currStart, nextStart
	return m.donation, nil
}

func (m *mockRepo) ActiveDonorCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error) {
	return m.donors, nil
}

func (m *mockRepo) ActiveVolunteerCounts(ctx context.Context, prevStart, currStart, nextStart time.Time) (CountPair, error) {
	return m.volunteers, nil
}

func (m *mockRepo) UpcomingEventCount(ctx context.Context, now time.Time) (int64, error) {
	m.gotNow = now
	return m.upcoming, nil
}

func (m *mockRepo) OngoingActivities(ctx context.Context, now time.Time) ([]OngoingActivity, error) {
	m.gotOngoingNow = now
	return m.ongoing, nil
}

func (m *mockRepo) LatestDonations(ctx context.Context) ([]LatestDonation, error) {
	return m.latest, nil
}

func (m *mockRepo) TopVolunteers(ctx context.Context) ([]TopVolunteer, error) {
	return m.top, nil
}

func (m *mockRepo) ImpactSummary(ctx context.Context, now, currStart, nextStart time.Time) (ImpactSummary, error) {
	m.gotImpactNow = now
	return m.impact, nil
}

func (m *mockRepo) DonationsByCategory(ctx context.Context, currStart, nextStart time.Time) ([]DonationCategorySlice, error) {
	m.gotCatCurr, m.gotCatNext = currStart, nextStart
	return m.byCategory, nil
}

func (m *mockRepo) ResolveVolunteerID(ctx context.Context, userID int) (int, error) {
	m.gotUserID = userID
	if m.volunteerErr != nil {
		return 0, m.volunteerErr
	}
	return m.volunteerID, nil
}

func (m *mockRepo) MonthlyActivityStats(ctx context.Context, volunteerID int, from, to time.Time) ([]monthlyActivityRow, error) {
	m.gotVolID, m.gotActFrom, m.gotActTo = volunteerID, from, to
	return m.activityRows, nil
}

func (m *mockRepo) MonthlyDonationStats(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonationRow, error) {
	m.gotVolKey = volunteerKey
	return m.donationRows, nil
}

func (m *mockRepo) CountMyDonors(ctx context.Context, volunteerKey string) (int64, error) {
	return m.donorCount, nil
}

func (m *mockRepo) MonthlyDonorCounts(ctx context.Context, volunteerKey string, from, to time.Time) ([]monthlyDonorRow, error) {
	return m.donorRows, nil
}

func TestGetSummary(t *testing.T) {
	repo := &mockRepo{
		donation:   AmountPair{Current: 150, Previous: 100}, // +50%
		expense:    AmountPair{Current: 40, Previous: 50},   // -20%
		donors:     CountPair{Current: 110, Previous: 100},  // +10%
		volunteers: CountPair{Current: 90, Previous: 100},   // -10%
		upcoming:   4,
	}
	fixed := time.Date(2026, time.March, 15, 10, 30, 0, 0, time.UTC)
	svc := &Service{Repo: repo, now: func() time.Time { return fixed }}

	got, err := svc.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TotalDonation.Current != 150 || got.TotalDonation.Previous != 100 || got.TotalDonation.Percentage != 50 {
		t.Fatalf("total_donation = %+v, want {150 100 50}", got.TotalDonation)
	}
	if got.TotalExpense.Current != 40 || got.TotalExpense.Previous != 50 || got.TotalExpense.Percentage != -20 {
		t.Fatalf("total_expense = %+v, want {40 50 -20}", got.TotalExpense)
	}
	if got.ActiveDonors.Percentage != 10 {
		t.Fatalf("active_donors percentage = %v, want 10", got.ActiveDonors.Percentage)
	}
	if got.ActiveVolunteers.Percentage != -10 {
		t.Fatalf("active_volunteers percentage = %v, want -10", got.ActiveVolunteers.Percentage)
	}
	if got.UpcomingEvents.Count != 4 {
		t.Fatalf("upcoming_events count = %v, want 4", got.UpcomingEvents.Count)
	}

	// Month boundaries derived from the fixed clock (March 2026, UTC).
	wantCurr := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	wantNext := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	wantPrev := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	if !repo.gotCurr.Equal(wantCurr) || !repo.gotNext.Equal(wantNext) || !repo.gotPrev.Equal(wantPrev) {
		t.Fatalf("boundaries: prev=%v curr=%v next=%v; want prev=%v curr=%v next=%v",
			repo.gotPrev, repo.gotCurr, repo.gotNext, wantPrev, wantCurr, wantNext)
	}
	if !repo.gotNow.Equal(fixed) {
		t.Fatalf("upcoming events now = %v, want %v", repo.gotNow, fixed)
	}
}

func TestGetHome(t *testing.T) {
	fixed := time.Date(2026, time.June, 19, 9, 0, 0, 0, time.UTC)
	repo := &mockRepo{
		ongoing: []OngoingActivity{
			{ID: 1, Name: "Donor Darah", Status: "active"},
		},
		latest: []LatestDonation{
			{ID: 1, DonatorName: "Budi Santoso", Amount: 2500000},
		},
		top: []TopVolunteer{
			{VolunteerID: 1, Name: "Rangga R", TotalHours: 48},
		},
		impact: ImpactSummary{ActiveVolunteers: 4, CompletedActivities: 15, CurrentMonthExpense: 750000},
	}
	svc := &Service{Repo: repo, now: func() time.Time { return fixed }}

	got, err := svc.GetHome(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.OngoingActivities) != 1 || got.OngoingActivities[0].Name != "Donor Darah" {
		t.Fatalf("ongoing_activities = %+v", got.OngoingActivities)
	}
	if len(got.LatestDonations) != 1 || got.LatestDonations[0].DonatorName != "Budi Santoso" || got.LatestDonations[0].Amount != 2500000 {
		t.Fatalf("latest_donations = %+v", got.LatestDonations)
	}
	if len(got.TopVolunteers) != 1 || got.TopVolunteers[0].TotalHours != 48 {
		t.Fatalf("top_volunteers = %+v", got.TopVolunteers)
	}
	if got.ImpactSummary.ActiveVolunteers != 4 || got.ImpactSummary.CompletedActivities != 15 || got.ImpactSummary.CurrentMonthExpense != 750000 {
		t.Fatalf("impact_summary = %+v", got.ImpactSummary)
	}
	if !repo.gotOngoingNow.Equal(fixed) {
		t.Fatalf("ongoing now = %v, want %v", repo.gotOngoingNow, fixed)
	}
	// completed_activities is schedule-derived (end_at < now), so the same clock
	// must reach ImpactSummary.
	if !repo.gotImpactNow.Equal(fixed) {
		t.Fatalf("impact now = %v, want %v", repo.gotImpactNow, fixed)
	}
}

func TestGetHome_EmptyWidgetsSerializeAsArrays(t *testing.T) {
	repo := &mockRepo{} // all slices nil
	svc := &Service{Repo: repo, now: time.Now}

	got, err := svc.GetHome(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.OngoingActivities == nil || got.LatestDonations == nil || got.TopVolunteers == nil {
		t.Fatalf("slices must be non-nil to serialize as []: %+v", got)
	}
	if len(got.OngoingActivities) != 0 || len(got.LatestDonations) != 0 || len(got.TopVolunteers) != 0 {
		t.Fatalf("expected empty widgets, got %+v", got)
	}
}

func TestGetDonationsByCategory(t *testing.T) {
	repo := &mockRepo{
		byCategory: []DonationCategorySlice{
			{CategoryID: 1, CategoryName: "Pendidikan", Total: 7000000},
			{CategoryID: 2, CategoryName: "Bencana", Total: 3000000},
		},
	}
	fixed := time.Date(2026, time.June, 19, 9, 0, 0, 0, time.UTC)
	svc := &Service{Repo: repo, now: func() time.Time { return fixed }}

	got, err := svc.GetDonationsByCategory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TotalAmount != 10000000 {
		t.Fatalf("total_amount = %v, want 10000000", got.TotalAmount)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("categories len = %d, want 2", len(got.Categories))
	}
	if got.Categories[0].CategoryName != "Pendidikan" || got.Categories[0].Total != 7000000 {
		t.Fatalf("first slice = %+v", got.Categories[0])
	}

	// Current-month bounds derived from the fixed clock (June 2026).
	wantCurr := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	wantNext := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	if !repo.gotCatCurr.Equal(wantCurr) || !repo.gotCatNext.Equal(wantNext) {
		t.Fatalf("month bounds: curr=%v next=%v; want curr=%v next=%v",
			repo.gotCatCurr, repo.gotCatNext, wantCurr, wantNext)
	}
}

func TestGetDonationsByCategory_EmptySerializesAsArray(t *testing.T) {
	svc := &Service{Repo: &mockRepo{}, now: time.Now}

	got, err := svc.GetDonationsByCategory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Categories == nil {
		t.Fatalf("categories must be non-nil to serialize as []")
	}
	if got.TotalAmount != 0 || len(got.Categories) != 0 {
		t.Fatalf("expected empty pie, got %+v", got)
	}
}

func TestGetVolunteerDashboard(t *testing.T) {
	repo := &mockRepo{
		volunteerID: 7,
		donorCount:  12,
		activityRows: []monthlyActivityRow{
			{YM: "2026-06", ActivityHours: 42.5, PhilosophyHours: 8, MissionHours: 4}, // current
			{YM: "2026-04", ActivityHours: 10},                                        // older
		},
		donationRows: []monthlyDonationRow{
			{YM: "2026-06", MyDonation: 2500000, CollectedDonation: 8000000}, // current
		},
		donorRows: []monthlyDonorRow{
			{YM: "2026-06", DonorCount: 3},
			{YM: "2026-04", DonorCount: 1},
		},
		expenseRows: []monthlyExpenseRow{
			{YM: "2026-06", ExpenseAmount: 320000},
			{YM: "2026-04", ExpenseAmount: 90000},
		},
	}
	fixed := time.Date(2026, time.June, 15, 9, 0, 0, 0, time.UTC)
	svc := &Service{Repo: repo, now: func() time.Time { return fixed }}

	got, err := svc.GetVolunteerDashboard(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Identity comes from the resolved volunteer; donor/group key is the id as text.
	if repo.gotUserID != 99 || repo.gotVolID != 7 || repo.gotVolKey != "7" {
		t.Fatalf("identity wiring wrong: userID=%d volID=%d volKey=%q", repo.gotUserID, repo.gotVolID, repo.gotVolKey)
	}

	// Cards = current month (June) bucket.
	if got.TotalActivityHours != 42.5 || got.PhilosophyHours != 8 || got.MissionHours != 4 {
		t.Fatalf("hour cards wrong: %+v", got)
	}
	if got.TotalDonors != 12 {
		t.Fatalf("total_donors = %d, want 12", got.TotalDonors)
	}
	if got.TotalMyDonation != 2500000 || got.TotalCollectedDonation != 8000000 {
		t.Fatalf("donation cards wrong: my=%v collected=%v", got.TotalMyDonation, got.TotalCollectedDonation)
	}
	if got.TotalMyExpense != 320000 {
		t.Fatalf("total_my_expense = %v, want 320000", got.TotalMyExpense)
	}
	if got.MonthlyChart[5].ExpenseAmount != 320000 || got.MonthlyChart[3].ExpenseAmount != 90000 {
		t.Fatalf("chart expense: jun=%v apr=%v", got.MonthlyChart[5].ExpenseAmount, got.MonthlyChart[3].ExpenseAmount)
	}

	// 6-month chart, oldest first: Jan..Jun.
	if len(got.MonthlyChart) != 6 {
		t.Fatalf("chart len = %d, want 6", len(got.MonthlyChart))
	}
	wantMonths := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	for i, m := range wantMonths {
		if got.MonthlyChart[i].Month != m {
			t.Fatalf("chart[%d].month = %q, want %q", i, got.MonthlyChart[i].Month, m)
		}
	}
	// April had activity + 1 donor, but no donations -> partial fill.
	if got.MonthlyChart[3].ActivityHours != 10 || got.MonthlyChart[3].MyDonation != 0 || got.MonthlyChart[3].DonorCount != 1 {
		t.Fatalf("April bucket wrong: %+v", got.MonthlyChart[3])
	}
	// Missing month (Feb) -> zero-filled.
	if got.MonthlyChart[1].ActivityHours != 0 || got.MonthlyChart[1].CollectedDonation != 0 || got.MonthlyChart[1].DonorCount != 0 {
		t.Fatalf("Feb bucket should be zero: %+v", got.MonthlyChart[1])
	}
	// June bucket matches the cards + donor_count.
	if got.MonthlyChart[5].ActivityHours != 42.5 || got.MonthlyChart[5].CollectedDonation != 8000000 || got.MonthlyChart[5].DonorCount != 3 {
		t.Fatalf("June bucket wrong: %+v", got.MonthlyChart[5])
	}

	// 6-month window passed to the repo: [2026-01-01, 2026-07-01).
	wantFrom := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	if !repo.gotActFrom.Equal(wantFrom) || !repo.gotActTo.Equal(wantTo) {
		t.Fatalf("window: from=%v to=%v; want from=%v to=%v", repo.gotActFrom, repo.gotActTo, wantFrom, wantTo)
	}
}

func TestGetVolunteerDashboard_NotAVolunteer(t *testing.T) {
	repo := &mockRepo{volunteerErr: ErrVolunteerNotFound}
	svc := &Service{Repo: repo, now: time.Now}

	_, err := svc.GetVolunteerDashboard(context.Background(), 1)
	if !errors.Is(err, ErrVolunteerNotFound) {
		t.Fatalf("expected ErrVolunteerNotFound, got %v", err)
	}
}

func TestGetSummary_ZeroPreviousMonth(t *testing.T) {
	repo := &mockRepo{
		donation:   AmountPair{Current: 5000, Previous: 0}, // => 100
		donors:     CountPair{Current: 0, Previous: 0},     // => 0
		volunteers: CountPair{Current: 10, Previous: 0},    // => 100
		upcoming:   0,
	}
	svc := &Service{Repo: repo, now: time.Now}

	got, err := svc.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TotalDonation.Percentage != 100 {
		t.Fatalf("total_donation percentage = %v, want 100", got.TotalDonation.Percentage)
	}
	if got.ActiveDonors.Percentage != 0 {
		t.Fatalf("active_donors percentage = %v, want 0", got.ActiveDonors.Percentage)
	}
	if got.ActiveVolunteers.Percentage != 100 {
		t.Fatalf("active_volunteers percentage = %v, want 100", got.ActiveVolunteers.Percentage)
	}
}
