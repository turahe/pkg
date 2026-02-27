package types

// TimeRange holds start and end time strings (e.g. IANA or RFC3339) for time-bounded queries.
type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
