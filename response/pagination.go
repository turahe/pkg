package response

// DefaultPageNumber, DefaultPageSize, and MaxPageSize are the pagination defaults and cap used by handler and repository.
const (
	DefaultPageNumber = 1
	DefaultPageSize   = 10
	MaxPageSize       = 100
)

// CursorPaginationResponse holds data, nextCursor, and hasNext for cursor-based pagination (used with CursorPaginated).
type CursorPaginationResponse struct {
	Data       interface{} `json:"data"`       // The actual data array
	NextCursor *string     `json:"nextCursor"` // Cursor for the next page (null if no more pages)
	HasNext    bool        `json:"hasNext"`    // Whether there are more items available
}

// SimplePaginationResponse holds data and pagination state for offset-based pagination (used with SimplePaginated).
type SimplePaginationResponse struct {
	Data       interface{} `json:"data"`       // The actual data array
	PageNumber int         `json:"pageNumber"` // Current page number
	PageSize   int         `json:"pageSize"`   // Number of items per page
	HasNext    bool        `json:"hasNext"`    // Whether there is a next page
	HasPrev    bool        `json:"hasPrev"`    // Whether there is a previous page
}
