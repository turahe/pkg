package util

import (
	"testing"
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

func TestFormatPhoneNumber(t *testing.T) {
	t.Run("nil phone returns nil", func(t *testing.T) {
		got := FormatPhoneNumber(nil, nil)
		if got != nil {
			t.Errorf("FormatPhoneNumber(nil, nil) = %v, want nil", got)
		}
	})

	t.Run("empty phone returns same pointer", func(t *testing.T) {
		empty := ""
		got := FormatPhoneNumber(&empty, nil)
		if got != &empty {
			t.Errorf("FormatPhoneNumber(&empty, nil) should return same pointer for empty string")
		}
		if got == nil || *got != "" {
			t.Errorf("FormatPhoneNumber(&empty, nil) = %v, want \"\"", got)
		}
	})

	t.Run("valid US number with no country code formats to E164", func(t *testing.T) {
		phone := "6502530000"
		got := FormatPhoneNumber(&phone, nil)
		if got == nil {
			t.Fatal("FormatPhoneNumber returned nil")
		}
		// E.164 for US (650) 253-0000
		if *got != "+16502530000" {
			t.Errorf("FormatPhoneNumber(%q, nil) = %q, want +16502530000", phone, *got)
		}
	})

	t.Run("valid US number with country US formats to E164", func(t *testing.T) {
		phone := "(650) 253-0000"
		country := "US"
		got := FormatPhoneNumber(&phone, &country)
		if got == nil {
			t.Fatal("FormatPhoneNumber returned nil")
		}
		if *got != "+16502530000" {
			t.Errorf("FormatPhoneNumber(%q, US) = %q, want +16502530000", phone, *got)
		}
	})

	t.Run("valid number with explicit region", func(t *testing.T) {
		phone := "08123456789"
		country := "ID"
		got := FormatPhoneNumber(&phone, &country)
		if got == nil {
			t.Fatal("FormatPhoneNumber returned nil")
		}
		// Indonesian mobile in E.164
		if *got != "+628123456789" {
			t.Errorf("FormatPhoneNumber(%q, ID) = %q, want +628123456789", phone, *got)
		}
	})

	t.Run("invalid number returns original phone", func(t *testing.T) {
		phone := "123"
		got := FormatPhoneNumber(&phone, nil)
		if got != &phone {
			t.Errorf("FormatPhoneNumber should return original pointer when invalid")
		}
		if *got != "123" {
			t.Errorf("FormatPhoneNumber(123) = %q, want 123 (original)", *got)
		}
	})

	t.Run("unparseable returns original phone", func(t *testing.T) {
		phone := "not-a-number"
		got := FormatPhoneNumber(&phone, nil)
		if got != &phone {
			t.Errorf("FormatPhoneNumber should return original pointer when parse fails")
		}
		if *got != "not-a-number" {
			t.Errorf("FormatPhoneNumber(not-a-number) = %q, want not-a-number (original)", *got)
		}
	})

	t.Run("empty country code uses US default", func(t *testing.T) {
		phone := "6502530000"
		emptyCountry := ""
		got := FormatPhoneNumber(&phone, &emptyCountry)
		if got == nil {
			t.Fatal("FormatPhoneNumber returned nil")
		}
		if *got != "+16502530000" {
			t.Errorf("FormatPhoneNumber(%q, &\"\") = %q, want +16502530000 (US default)", phone, *got)
		}
	})
}
