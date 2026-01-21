package types

import (
	"encoding/json"
	"testing"
)

func TestTimeRange(t *testing.T) {
	tr := TimeRange{Start: "09:00", End: "10:00"}
	if tr.Start != "09:00" || tr.End != "10:00" {
		t.Errorf("TimeRange = %+v", tr)
	}
}

func TestTimeRange_JSON(t *testing.T) {
	tr := TimeRange{Start: "08:30", End: "09:30"}
	b, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out TimeRange
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Start != tr.Start || out.End != tr.End {
		t.Errorf("Unmarshal got %+v", out)
	}
}
