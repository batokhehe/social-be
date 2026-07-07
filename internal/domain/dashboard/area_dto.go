package dashboard

// Shared response shapes for the area (Hu Ai / Xie Li) dashboards. The DTO is
// common infrastructure; the KPI business logic lives in each dashboard's own
// service (HuAiService / XieLiService), not in a shared engine.

// AreaSummary is the current-month KPI card block.
type AreaSummary struct {
	TotalActivities     int64   `json:"total_activities"`
	TotalVolunteerHours float64 `json:"total_volunteer_hours"`
	ActiveVolunteers    int64   `json:"active_volunteers"` // membership status = 'active'
	TotalVolunteers     int64   `json:"total_volunteers"`
	TotalDonors         int64   `json:"total_donors"`
	TotalDonations      float64 `json:"total_donations"`
	TotalExpenses       float64 `json:"total_expenses"` // SUM(expenses.amount), status != cancelled, area scope
}

// AreaMonthlyStat is one month of the 6-month chart. active_volunteers here is
// monthly engagement (distinct volunteers with a check-in that month), which is
// the meaningful per-month metric.
type AreaMonthlyStat struct {
	Month            string  `json:"month"`
	Activities       int64   `json:"activities"`
	VolunteerHours   float64 `json:"volunteer_hours"`
	ActiveVolunteers int64   `json:"active_volunteers"`
	TotalDonors      int64   `json:"total_donors"`
	TotalDonations   float64 `json:"total_donations"`
	TotalExpenses    float64 `json:"total_expenses"` // SUM(expenses.amount), status != cancelled, area scope
}

// AreaDashboardResponse is the payload for GET /dashboard/hu-ai and /xie-li.
type AreaDashboardResponse struct {
	Summary      AreaSummary       `json:"summary"`
	MonthlyChart []AreaMonthlyStat `json:"monthly_chart"`
}
