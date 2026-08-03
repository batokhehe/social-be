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
	TotalExpense     MetricCard         `json:"total_expense"`
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
	ActiveVolunteers    int64   `json:"active_volunteers"`
	CompletedActivities int64   `json:"completed_activities"`
	CurrentMonthExpense float64 `json:"current_month_expense"`
}

// TrendPoint is one month of a Home Dashboard trend chart. Month is "YYYY-MM";
// Total is a count for the registration trends and a summed amount for the
// donation trend.
type TrendPoint struct {
	Month string  `json:"month"`
	Total float64 `json:"total"`
}

// HomeResponse is the payload for GET /dashboard/home.
type HomeResponse struct {
	OngoingActivities []OngoingActivity `json:"ongoing_activities"`
	LatestDonations   []LatestDonation  `json:"latest_donations"`
	TopVolunteers     []TopVolunteer    `json:"top_volunteers"`
	ImpactSummary     ImpactSummary     `json:"impact_summary"`

	// Monthly trends for the last 6 months (current month included), oldest
	// first, always exactly 6 points with zero-filled gaps. Global: not filtered
	// by master area or by the logged-in user.
	VolunteerTrend []TrendPoint `json:"volunteer_trend"`
	DonorTrend     []TrendPoint `json:"donor_trend"`
	DonationTrend  []TrendPoint `json:"donation_trend"`
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

// --- GET /dashboard/volunteer (personal volunteer dashboard) ---

// VolunteerMonthlyStat is one month of the 6-month chart for the logged-in
// volunteer. Hours come from event_attendances.total_hours; donations are money
// donations (type = 0).
type VolunteerMonthlyStat struct {
	Month             string  `json:"month"` // short month label, e.g. "Jan"
	ActivityHours     float64 `json:"activity_hours"`
	PhilosophyHours   float64 `json:"philosophy_hours"`
	MissionHours      float64 `json:"mission_hours"`
	MyDonation        float64 `json:"my_donation"`
	CollectedDonation float64 `json:"collected_donation"`
	DonorCount        int64   `json:"donor_count"` // donors acquired that month under the volunteer's groups
	ExpenseAmount     float64 `json:"expense_amount"`
}

// VolunteerDashboardResponse is the payload for GET /dashboard/volunteer. The
// scalar cards are the CURRENT month; MonthlyChart covers the last 6 months
// (including the current month), oldest first.
type VolunteerDashboardResponse struct {
	TotalActivityHours     float64                `json:"total_activity_hours"`     // card 1
	PhilosophyHours        float64                `json:"philosophy_hours"`         // card 2 (category_activity_id = 4)
	MissionHours           float64                `json:"mission_hours"`            // card 3 (category_activity_id = 2)
	TotalDonors            int64                  `json:"total_donors"`             // card 4 (all-time)
	TotalMyDonation        float64                `json:"total_my_donation"`        // card 5
	TotalCollectedDonation float64                `json:"total_collected_donation"` // card 6
	TotalMyExpense         float64                `json:"total_my_expense"`         // current-month expenses
	MonthlyChart           []VolunteerMonthlyStat `json:"monthly_chart"`
}
