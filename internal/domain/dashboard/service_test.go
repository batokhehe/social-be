package dashboard

import (
	"context"
	"testing"
	"time"
)

type mockRepo struct {
	donation   AmountPair
	donors     CountPair
	volunteers CountPair
	upcoming   int64

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

func (m *mockRepo) ImpactSummary(ctx context.Context, now time.Time) (ImpactSummary, error) {
	m.gotImpactNow = now
	return m.impact, nil
}

func (m *mockRepo) DonationsByCategory(ctx context.Context, currStart, nextStart time.Time) ([]DonationCategorySlice, error) {
	m.gotCatCurr, m.gotCatNext = currStart, nextStart
	return m.byCategory, nil
}

func TestGetSummary(t *testing.T) {
	repo := &mockRepo{
		donation:   AmountPair{Current: 150, Previous: 100}, // +50%
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
		impact: ImpactSummary{ActiveVolunteers: 4, CompletedActivities: 15},
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
	if got.ImpactSummary.ActiveVolunteers != 4 || got.ImpactSummary.CompletedActivities != 15 {
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
