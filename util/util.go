package util

import (
	"reflect"
	"time"

	"github.com/turahe/pkg/types"
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

func GetCurrentTimeRange() *types.TimeRange {
	var timeRange = []types.TimeRange{
		{Start: "07:30", End: "08:30"},
		{Start: "08:30", End: "09:30"},
		{Start: "09:30", End: "10:30"},
		{Start: "10:30", End: "11:30"},
		{Start: "11:30", End: "13:30"},
		{Start: "13:30", End: "14:30"},
		{Start: "14:30", End: "15:30"},
		{Start: "15:30", End: "16:30"},
		{Start: "16:30", End: "17:30"},
		{Start: "17:30", End: "18:30"},
	}

	now := time.Now()
	currentTime := now.Format("15:04")

	for _, tr := range timeRange {
		if currentTime >= tr.Start && currentTime < tr.End {
			return &tr
		}
	}

	return nil
}
