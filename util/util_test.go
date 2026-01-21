package util

import (
	"testing"

	"github.com/turahe/pkg/types"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"empty string", "", true},
		{"non-empty string", "x", false},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"zero int64", int64(0), true},
		{"zero float64", 0.0, true},
		{"non-zero float64", 1.5, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"nil slice", ([]int)(nil), true},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"nil map", (map[string]int)(nil), true},
		{"empty map", map[string]int{}, true},
		{"nil ptr", (*int)(nil), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInAnySlice(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if !InAnySlice([]int{1, 2, 3}, 2) {
			t.Error("expected true for 2 in [1,2,3]")
		}
		if InAnySlice([]int{1, 2, 3}, 4) {
			t.Error("expected false for 4 in [1,2,3]")
		}
	})
	t.Run("string", func(t *testing.T) {
		if !InAnySlice([]string{"a", "b"}, "a") {
			t.Error("expected true for 'a' in slice")
		}
		if InAnySlice([]string{"a", "b"}, "c") {
			t.Error("expected false for 'c' in slice")
		}
	})
	t.Run("empty slice", func(t *testing.T) {
		if InAnySlice([]int{}, 1) {
			t.Error("expected false for empty slice")
		}
	})
}

func TestRemoveDuplicates(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		got := RemoveDuplicates([]int{1, 2, 2, 3, 1, 3})
		want := []int{1, 2, 3}
		if len(got) != len(want) {
			t.Errorf("len = %d, want %d", len(got), len(want))
		}
		for i, v := range want {
			if i >= len(got) || got[i] != v {
				t.Errorf("RemoveDuplicates() = %v, want %v", got, want)
				break
			}
		}
	})
	t.Run("string", func(t *testing.T) {
		got := RemoveDuplicates([]string{"a", "b", "a", "c"})
		if len(got) != 3 {
			t.Errorf("len = %d, want 3", len(got))
		}
	})
	t.Run("empty and zeros", func(t *testing.T) {
		// IsEmpty(0) is true, so 0 is filtered out
		got := RemoveDuplicates([]int{0, 1, 0, 2})
		// 0 is empty so excluded; we get 1, 2
		if len(got) != 2 {
			t.Errorf("len = %d (got %v), zeros should be excluded", len(got), got)
		}
	})
}

func TestGetCurrentTimeRange(t *testing.T) {
	tr := GetCurrentTimeRange()
	// Either nil (outside 07:30–18:30) or valid *TimeRange
	if tr != nil {
		if tr.Start == "" || tr.End == "" {
			t.Error("TimeRange should have Start and End when non-nil")
		}
		// ensure it's a copy of one of the defined ranges
		valid := map[string]bool{
			"07:30": true, "08:30": true, "09:30": true, "10:30": true,
			"11:30": true, "13:30": true, "14:30": true, "15:30": true,
			"16:30": true, "17:30": true, "18:30": true,
		}
		if !valid[tr.Start] || !valid[tr.End] {
			t.Errorf("unexpected TimeRange %+v", tr)
		}
	}
}

func TestGetCurrentTimeRange_ReturnsTimeRangeOrNil(t *testing.T) {
	// Just ensure it doesn't panic
	_ = GetCurrentTimeRange()
}

// Test types import and struct
func TestTypesTimeRange(t *testing.T) {
	_ = types.TimeRange{Start: "09:00", End: "10:00"}
}
