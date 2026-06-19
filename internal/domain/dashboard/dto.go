package dashboard

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
