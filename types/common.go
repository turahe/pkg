package types

// Conditions is a map from SQL condition fragments to argument values, used by repository WHERE clauses (e.g. "id = ?" -> value).
type Conditions map[string]interface{}

// PageInfo holds offset-based pagination request parameters (pageNumber, pageSize) with form and JSON tags.
type PageInfo struct {
	PageNumber int `form:"pageNumber" json:"pageNumber"`
	PageSize   int `form:"pageSize" json:"pageSize"`
}
