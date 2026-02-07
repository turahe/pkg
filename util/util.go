package util

import (
	"reflect"

	"github.com/nyaruka/phonenumbers"
)

func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	default:
		return false
	}
}

func InAnySlice[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func RemoveDuplicates[T comparable](haystack []T) []T {
	encountered := map[T]bool{}
	result := []T{}

	for _, item := range haystack {
		if !encountered[item] && !IsEmpty(item) {
			encountered[item] = true
			result = append(result, item)
		}
	}

	return result
}

// FormatPhoneNumber formats a phone number using the nyaruka/phonenumbers library.
// If country code is provided, it will be used for parsing; otherwise, US is used as default region.
// Returns the formatted phone number in E.164 format, or the original phone if formatting fails.
func FormatPhoneNumber(phone *string, countryCode *string) *string {
	if phone == nil || *phone == "" {
		return phone
	}

	// Default to "US" if no country code is provided
	defaultRegion := "US"
	if countryCode != nil && *countryCode != "" {
		defaultRegion = *countryCode
	}

	// Parse the phone number
	num, err := phonenumbers.Parse(*phone, defaultRegion)
	if err != nil {
		// If parsing fails, return original phone number
		return phone
	}

	// Check if the number is valid
	if !phonenumbers.IsValidNumber(num) {
		// If invalid, return original phone number
		return phone
	}

	// Format in E.164 format (e.g., +1234567890)
	formatted := phonenumbers.Format(num, phonenumbers.E164)
	return &formatted
}
