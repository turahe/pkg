package types

type Conditions map[string]interface{}

// PageInfo represents simple offset-based pagination request
type PageInfo struct {
	PageNumber int `form:"pageNumber" json:"pageNumber"`
	PageSize   int `form:"pageSize" json:"pageSize"`
}
