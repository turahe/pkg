package util

import (
	"fmt"
	"reflect"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// IsEmpty returns true if value is zero, nil, or empty (e.g. empty string, zero number, nil slice, empty map).
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

// InAnySlice returns true if needle is equal to any element of haystack.
func InAnySlice[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// RemoveDuplicates returns a new slice with duplicate elements removed; elements satisfying IsEmpty are omitted.
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

// FormatPhoneNumber parses phone with optional countryCode as default region and returns E.164-formatted string.
// Uses nyaruka/phonenumbers. If parsing or validation fails, returns the original phone unchanged.
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

// FormatCurrency formats a numeric amount with the given currency code (e.g., "USD", "EUR", "GBP").
// Returns a formatted currency string with the currency symbol and proper locale formatting.
// If formatting fails, returns a basic formatted string with currency code suffix.
func FormatCurrency(amount float64, currencyCode string) string {
	if currencyCode == "" {
		return fmt.Sprintf("%.2f", amount)
	}

	// Parse currency code
	curr, err := currency.ParseISO(currencyCode)
	if err != nil {
		// Fallback: simple format with currency code
		return fmt.Sprintf("%.2f %s", amount, currencyCode)
	}

	// Determine language based on currency
	// Try common locales for the currency
	langTag := language.English // Default to English

	// Map common currency codes to preferred locales
	switch currencyCode {
	case "EUR":
		langTag = language.German // or language.French, language.Italian
	case "GBP":
		langTag = language.English
	case "JPY":
		langTag = language.Japanese
	case "CNY":
		langTag = language.Chinese
	case "INR":
		langTag = language.Hindi
	case "BRL":
		langTag = language.Portuguese
	case "MXN":
		langTag = language.Spanish
	case "CAD":
		langTag = language.English
	case "AUD":
		langTag = language.English
	case "IDR":
		langTag = language.Indonesian
	}

	// Create a printer for the language and format with currency
	printer := message.NewPrinter(langTag)
	return printer.Sprintf("%v %v", curr, amount)
}
