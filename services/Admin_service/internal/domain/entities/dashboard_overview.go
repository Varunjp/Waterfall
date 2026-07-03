package entities

type DashboardOverview struct {
	TotalUsers        int64   `json:"totalUsers"`
	TotalApps         int64   `json:"totalApps"`
	ActiveSubscribers int64   `json:"activeSubscribers"`
	RevenueMonth      float64 `json:"revenueMonth"`
	RevenueLastMonth  float64 `json:"revenueLastMonth"`
	JobsToday         int64   `json:"jobsToday"`
	FailedJobsToday   int64   `json:"jobsFailed"`
}
