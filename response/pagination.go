package response

// Pagination defaults used by handler and repository layers.
const (
	DefaultPageNumber = 1
	DefaultPageSize   = 10
	MaxPageSize       = 100
)

// CursorPaginationResponse represents cursor-based pagination response
type CursorPaginationResponse struct {
	Data       interface{} `json:"data"`       // The actual data array
	NextCursor *string     `json:"nextCursor"` // Cursor for the next page (null if no more pages)
	HasNext    bool        `json:"hasNext"`    // Whether there are more items available
}

// SimplePaginationResponse represents simple offset-based pagination response
type SimplePaginationResponse struct {
	Data       interface{} `json:"data"`       // The actual data array
	PageNumber int         `json:"pageNumber"` // Current page number
	PageSize   int         `json:"pageSize"`   // Number of items per page
	HasNext    bool        `json:"hasNext"`    // Whether there is a next page
	HasPrev    bool        `json:"hasPrev"`    // Whether there is a previous page
}
