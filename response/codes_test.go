package response

import (
	"net/http"
	"testing"
)

func TestBuildResponseCode(t *testing.T) {
	tests := []struct {
		httpStatus  int
		serviceCode string
		caseCode    string
		want        int
	}{
		{http.StatusOK, ServiceCodeCommon, CaseCodeSuccess, 2000001},
		{http.StatusCreated, "03", CaseCodeCreated, 2010302},
		{http.StatusNotFound, "04", CaseCodeNotFound, 4040431},
		{http.StatusUnprocessableEntity, "00", CaseCodeValidationError, 4220011},
		{http.StatusInternalServerError, ServiceCodeCommon, CaseCodeInternalError, 5000055},
		{5, "00", "01", 50001},
		{10, "00", "01", 100001},
		{100, "00", "01", 1000001},
	}
	for _, tt := range tests {
		got := BuildResponseCode(tt.httpStatus, tt.serviceCode, tt.caseCode)
		if got != tt.want {
			t.Errorf("BuildResponseCode(%d, %q, %q) = %d, want %d", tt.httpStatus, tt.serviceCode, tt.caseCode, got, tt.want)
		}
	}
}

func TestParseResponseCode(t *testing.T) {
	tests := []struct {
		code            int
		wantHTTP        int
		wantService     string
		wantCase        string
	}{
		{2000001, 200, "00", "01"},
		{2010302, 201, "03", "02"},
		{4040431, 404, "04", "31"},
		{4220011, 422, "00", "11"},
		{5000055, 500, "00", "55"},
	}
	for _, tt := range tests {
		httpStatus, serviceCode, caseCode := ParseResponseCode(tt.code)
		if httpStatus != tt.wantHTTP || serviceCode != tt.wantService || caseCode != tt.wantCase {
			t.Errorf("ParseResponseCode(%d) = (%d, %q, %q), want (%d, %q, %q)",
				tt.code, httpStatus, serviceCode, caseCode, tt.wantHTTP, tt.wantService, tt.wantCase)
		}
	}
}

func TestParseResponseCode_ShortCode(t *testing.T) {
	httpStatus, serviceCode, caseCode := ParseResponseCode(200001)
	if httpStatus != 0 || serviceCode != "" || caseCode != "" {
		t.Errorf("ParseResponseCode(200001) = (%d, %q, %q), want (0, \"\", \"\") for short code", httpStatus, serviceCode, caseCode)
	}
}

func TestBuildResponseCode_RoundTrip(t *testing.T) {
	httpStatus, serviceCode, caseCode := 201, "03", "02"
	code := BuildResponseCode(httpStatus, serviceCode, caseCode)
	gotHTTP, gotService, gotCase := ParseResponseCode(code)
	if gotHTTP != httpStatus || gotService != serviceCode || gotCase != caseCode {
		t.Errorf("RoundTrip: BuildResponseCode(%d,%q,%q)=%d, ParseResponseCode => (%d,%q,%q)",
			httpStatus, serviceCode, caseCode, code, gotHTTP, gotService, gotCase)
	}
}
