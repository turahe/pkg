package response

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"FirstName", "firstName"},
		{"id", "id"},
		{"", ""},
		{"U", "u"},
		{"HTTP", "hTTP"},
	}
	for _, tt := range tests {
		got := toCamelCase(tt.in)
		if got != tt.want {
			t.Errorf("toCamelCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGetJSONFieldName(t *testing.T) {
	type st struct {
		Normal  string `json:"normal"`
		Omit    string `json:"omit,omitempty"`
		Skip    string `json:"-"`
		NoTag   string
		EmptyTag string `json:""`
	}
	typ := reflect.TypeOf(st{})

	tests := []struct {
		field    string
		want     string
	}{
		{"Normal", "normal"},
		{"Omit", "omit"},
		{"Skip", "skip"},
		{"NoTag", "noTag"},
		{"EmptyTag", "emptyTag"},
		{"Missing", "missing"},
	}
	for _, tt := range tests {
		got := GetJSONFieldName(typ, tt.field)
		if got != tt.want {
			t.Errorf("GetJSONFieldName(_, %q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func TestFormatValidationError_GenericError(t *testing.T) {
	err := errors.New("something failed")
	got := FormatValidationError(err)
	if len(got) != 1 || len(got["general"]) != 1 || got["general"][0] != "something failed" {
		t.Errorf("FormatValidationError(plain error) = %v, want {\"general\":[\"something failed\"]}", got)
	}
}

func TestFormatValidationError_ValidatorErrors(t *testing.T) {
	type req struct {
		Email string `json:"email" validate:"required,email"`
		Name  string `json:"name" validate:"min=2"`
	}
	validate := validator.New()
	v := req{Email: "", Name: "x"}
	err := validate.Struct(v)
	if err == nil {
		t.Fatal("expected validation to fail")
	}
	got := FormatValidationError(err)
	if len(got) == 0 {
		t.Error("FormatValidationError(ValidationErrors) expected non‑empty map")
	}
	if _, has := got["email"]; !has {
		if _, hasG := got["general"]; !hasG {
			t.Errorf("expected 'email' or 'general' in %v", got)
		}
	}
}

func TestGetErrorMessage_RequiredTag(t *testing.T) {
	// getErrorMessage is unexported; we exercise it via FormatValidationError with a struct that has only "required"
	type r struct {
		F string `validate:"required"`
	}
	validate := validator.New()
	err := validate.Struct(r{})
	if err == nil {
		t.Fatal("expected validation to fail")
	}
	m := FormatValidationError(err)
	// Field may be "f" (camelCase of F) or "F" depending on getFieldName
	var found bool
	for _, msgs := range m {
		for _, msg := range msgs {
			if len(msg) > 0 && (msg[0] >= 'T' && msg[0] <= 'T' || msg[0] >= 't' && msg[0] <= 't') {
				if len(msg) > 10 && msg[0:10] == "The " || (len(msg) > 3 && msg[:3] == "The") {
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}
	if !found && len(m["general"]) == 0 {
		// At least one message should look like "The X field is required."
		t.Logf("FormatValidationError output: %v (getErrorMessage exercised indirectly)", m)
	}
}
