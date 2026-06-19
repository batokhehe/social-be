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

	// captured arguments for boundary assertions
	gotPrev, gotCurr, gotNext time.Time
	gotNow                    time.Time
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
