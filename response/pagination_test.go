package response

import (
	"encoding/json"
	"testing"
)

func TestCursorPaginationResponse(t *testing.T) {
	next := "abc"
	p := CursorPaginationResponse{
		Data:       []int{1, 2, 3},
		NextCursor: &next,
		HasNext:    true,
	}
	if p.Data == nil || !p.HasNext || p.NextCursor == nil || *p.NextCursor != "abc" {
		t.Errorf("CursorPaginationResponse: %+v", p)
	}
	// Round-trip JSON
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var q CursorPaginationResponse
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if q.NextCursor == nil || *q.NextCursor != "abc" || !q.HasNext {
		t.Errorf("Unmarshal: %+v", q)
	}
}

func TestCursorPaginationResponse_NilCursor(t *testing.T) {
	p := CursorPaginationResponse{
		Data:       []string{},
		NextCursor: nil,
		HasNext:    false,
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var q CursorPaginationResponse
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if q.NextCursor != nil || q.HasNext {
		t.Errorf("Unmarshal: NextCursor should be nil, HasNext false: %+v", q)
	}
}

func TestSimplePaginationResponse(t *testing.T) {
	p := SimplePaginationResponse{
		Data:       []string{"a", "b"},
		PageNumber: 2,
		PageSize:   10,
		HasNext:    true,
		HasPrev:    true,
	}
	if p.PageNumber != 2 || p.PageSize != 10 || !p.HasNext || !p.HasPrev {
		t.Errorf("SimplePaginationResponse: %+v", p)
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var q SimplePaginationResponse
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if q.PageNumber != 2 || q.PageSize != 10 || !q.HasNext || !q.HasPrev {
		t.Errorf("Unmarshal: %+v", q)
	}
}
