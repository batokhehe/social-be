package dashboard

import "time"

// MetricCard is a KPI card comparing the current month against the previous
// month, with the precomputed growth/decline percentage.
type MetricCard struct {
	Current    float64 `json:"current"`
	Previous   float64 `json:"previous"`
	Percentage float64 `json:"percentage"`
}

// UpcomingEventsCard is a count-only card (no month-over-month comparison).
type UpcomingEventsCard struct {
	Count int64 `json:"count"`
}

// SummaryResponse is the payload for GET /dashboard/summary.
type SummaryResponse struct {
	TotalDonation    MetricCard         `json:"total_donation"`
	ActiveDonors     MetricCard         `json:"active_donors"`
	ActiveVolunteers MetricCard         `json:"active_volunteers"`
	UpcomingEvents   UpcomingEventsCard `json:"upcoming_events"`
}

// --- GET /dashboard/home widgets ---

// OngoingActivity is an event currently running ("Kegiatan Berlangsung").
type OngoingActivity struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
	Status  string    `json:"status"`
}

// LatestDonation is a recent donation with its resolved donor name ("Donasi Terbaru").
type LatestDonation struct {
	ID          int       `json:"id"`
	DonatorName string    `json:"donator_name"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// TopVolunteer is a volunteer ranked by total contribution hours ("Relawan Terbaik").
type TopVolunteer struct {
	VolunteerID int     `json:"volunteer_id"`
	Name        string  `json:"name"`
	TotalHours  float64 `json:"total_hours"`
}

// ImpactSummary is the "Dampak Kita" card.
//
//   - ActiveVolunteers:    volunteers with status 'active' (not soft-deleted).
//   - CompletedActivities: events that have already finished, i.e. end_at < now
//     and not soft-deleted. Completion is derived from the schedule because the
//     event lifecycle only uses status 'active'/'inactive' (never 'completed').
type ImpactSummary struct {
	ActiveVolunteers    int64 `json:"active_volunteers"`
	CompletedActivities int64 `json:"completed_activities"`
}

// HomeResponse is the payload for GET /dashboard/home.
type HomeResponse struct {
	OngoingActivities []OngoingActivity `json:"ongoing_activities"`
	LatestDonations   []LatestDonation  `json:"latest_donations"`
	TopVolunteers     []TopVolunteer    `json:"top_volunteers"`
	ImpactSummary     ImpactSummary     `json:"impact_summary"`
}

// --- GET /dashboard/donations-by-category (pie chart) ---

// DonationCategorySlice is one pie slice: a category's summed money-donation
// amount for the current month. Percentage is intentionally omitted -- the
// frontend derives each slice's share from total / total_amount.
type DonationCategorySlice struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
}

// DonationByCategoryResponse is the payload for the donations-by-category pie
// chart (current month). TotalAmount is the grand total across all slices, so
// the frontend can compute percentages without a second pass.
type DonationByCategoryResponse struct {
	TotalAmount float64                 `json:"total_amount"`
	Categories  []DonationCategorySlice `json:"categories"`
}
