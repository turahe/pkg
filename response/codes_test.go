package response

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// BuildResponseCode
// ---------------------------------------------------------------------------

func TestBuildResponseCode(t *testing.T) {
	tests := []struct {
		name        string
		httpStatus  int
		serviceCode string
		caseCode    string
		want        int
	}{
		// Core service codes
		{"200 common success", http.StatusOK, ServiceCodeCommon, CaseCodeSuccess, 2000001},
		{"201 withdrawal created", http.StatusCreated, ServiceCodeWithdrawal, CaseCodeCreated, 2010302},
		{"404 user not found", http.StatusNotFound, ServiceCodeUser, CaseCodeNotFound, 4040431},
		{"422 validation error", http.StatusUnprocessableEntity, ServiceCodeCommon, CaseCodeValidationError, 4220011},
		{"500 internal error", http.StatusInternalServerError, ServiceCodeCommon, CaseCodeInternalError, 5000055},

		// HTTP status edge cases
		{"single-digit http status", 5, "00", "01", 50001},
		{"two-digit http status", 10, "00", "01", 100001},
		{"three-digit http status", 100, "00", "01", 1000001},

		// Auth service
		{"401 invalid token", http.StatusUnauthorized, ServiceCodeAuth, CaseCodeInvalidToken, 4010122},
		{"401 token expired", http.StatusUnauthorized, ServiceCodeAuth, CaseCodeTokenExpired, 4010123},
		{"401 2FA required", http.StatusUnauthorized, ServiceCodeAuth, CaseCodeTwoFactorRequired, 4010129},
		{"401 invalid OTP", http.StatusUnauthorized, ServiceCodeAuth, CaseCodeInvalidOTP, 4010130},

		// Email service (new)
		{"200 email change requested", http.StatusOK, ServiceCodeEmail, CaseCodeEmailChangeRequested, 2001870},
		{"200 email change verified", http.StatusOK, ServiceCodeEmail, CaseCodeEmailChangeVerified, 2001871},
		{"200 email change cancelled", http.StatusOK, ServiceCodeEmail, CaseCodeEmailChangeCancelled, 2001872},
		{"409 email already used", http.StatusConflict, ServiceCodeEmail, CaseCodeEmailAlreadyUsed, 4091873},
		{"422 email verification failed", http.StatusUnprocessableEntity, ServiceCodeEmail, CaseCodeEmailVerificationFailed, 4221874},

		// Phone service (new)
		{"200 phone change requested", http.StatusOK, ServiceCodePhone, CaseCodePhoneChangeRequested, 2001775},
		{"200 phone change verified", http.StatusOK, ServiceCodePhone, CaseCodePhoneChangeVerified, 2001776},
		{"200 phone change cancelled", http.StatusOK, ServiceCodePhone, CaseCodePhoneChangeCancelled, 2001777},
		{"409 phone already used", http.StatusConflict, ServiceCodePhone, CaseCodePhoneAlreadyUsed, 4091778},
		{"422 phone verification failed", http.StatusUnprocessableEntity, ServiceCodePhone, CaseCodePhoneVerificationFailed, 4221779},

		// Two-factor authentication service (new)
		{"200 2FA enabled", http.StatusOK, ServiceCodeTwoFactor, CaseCode2FAEnabled, 2001980},
		{"200 2FA disabled", http.StatusOK, ServiceCodeTwoFactor, CaseCode2FADisabled, 2001981},
		{"200 2FA setup initiated", http.StatusOK, ServiceCodeTwoFactor, CaseCode2FASetupInitiated, 2001982},
		{"200 2FA setup verified", http.StatusOK, ServiceCodeTwoFactor, CaseCode2FASetupVerified, 2001983},
		{"401 2FA invalid code", http.StatusUnauthorized, ServiceCodeTwoFactor, CaseCode2FAInvalidCode, 4011984},
		{"422 2FA code required", http.StatusUnprocessableEntity, ServiceCodeTwoFactor, CaseCode2FACodeRequired, 4221985},
		{"422 2FA not enabled", http.StatusUnprocessableEntity, ServiceCodeTwoFactor, CaseCode2FANotEnabled, 4221986},
		{"409 2FA already enabled", http.StatusConflict, ServiceCodeTwoFactor, CaseCode2FAAlreadyEnabled, 4091987},
		{"200 2FA recovery code used", http.StatusOK, ServiceCodeTwoFactor, CaseCode2FARecoveryCodeUsed, 2001988},
		{"401 2FA recovery code invalid", http.StatusUnauthorized, ServiceCodeTwoFactor, CaseCode2FARecoveryCodeInvalid, 4011989},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildResponseCode(tt.httpStatus, tt.serviceCode, tt.caseCode)
			assert.Equal(t, tt.want, got,
				"BuildResponseCode(%d, %q, %q)", tt.httpStatus, tt.serviceCode, tt.caseCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseResponseCode
// ---------------------------------------------------------------------------

func TestParseResponseCode(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		wantHTTP    int
		wantService string
		wantCase    string
	}{
		{"200 common success", 2000001, 200, "00", "01"},
		{"201 withdrawal created", 2010302, 201, "03", "02"},
		{"404 user not found", 4040431, 404, "04", "31"},
		{"422 validation error", 4220011, 422, "00", "11"},
		{"500 internal error", 5000055, 500, "00", "55"},

		// New service codes
		{"email change requested", 2001870, 200, "18", "70"},
		{"email already used", 4091873, 409, "18", "73"},
		{"phone change verified", 2001776, 200, "17", "76"},
		{"phone already used", 4091778, 409, "17", "78"},
		{"2FA enabled", 2001980, 200, "19", "80"},
		{"2FA invalid code", 4011984, 401, "19", "84"},
		{"2FA recovery invalid", 4011989, 401, "19", "89"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpStatus, serviceCode, caseCode := ParseResponseCode(tt.code)
			assert.Equal(t, tt.wantHTTP, httpStatus, "httpStatus for code %d", tt.code)
			assert.Equal(t, tt.wantService, serviceCode, "serviceCode for code %d", tt.code)
			assert.Equal(t, tt.wantCase, caseCode, "caseCode for code %d", tt.code)
		})
	}
}

func TestParseResponseCode_ShortCode(t *testing.T) {
	// Codes shorter than 7 digits must return zero values.
	httpStatus, serviceCode, caseCode := ParseResponseCode(200001)
	assert.Equal(t, 0, httpStatus)
	assert.Equal(t, "", serviceCode)
	assert.Equal(t, "", caseCode)
}

func TestParseResponseCode_Zero(t *testing.T) {
	httpStatus, serviceCode, caseCode := ParseResponseCode(0)
	assert.Equal(t, 0, httpStatus)
	assert.Equal(t, "", serviceCode)
	assert.Equal(t, "", caseCode)
}

// ---------------------------------------------------------------------------
// Round-trip
// ---------------------------------------------------------------------------

func TestBuildParseRoundTrip(t *testing.T) {
	cases := []struct {
		httpStatus  int
		serviceCode string
		caseCode    string
	}{
		// Existing codes
		{http.StatusOK, ServiceCodeCommon, CaseCodeSuccess},
		{http.StatusCreated, ServiceCodeWithdrawal, CaseCodeCreated},
		{http.StatusNotFound, ServiceCodeUser, CaseCodeNotFound},
		{http.StatusInternalServerError, ServiceCodeCommon, CaseCodeInternalError},
		// Email
		{http.StatusOK, ServiceCodeEmail, CaseCodeEmailChangeRequested},
		{http.StatusOK, ServiceCodeEmail, CaseCodeEmailChangeVerified},
		{http.StatusConflict, ServiceCodeEmail, CaseCodeEmailAlreadyUsed},
		{http.StatusUnprocessableEntity, ServiceCodeEmail, CaseCodeEmailVerificationFailed},
		// Phone
		{http.StatusOK, ServiceCodePhone, CaseCodePhoneChangeRequested},
		{http.StatusConflict, ServiceCodePhone, CaseCodePhoneAlreadyUsed},
		// 2FA
		{http.StatusOK, ServiceCodeTwoFactor, CaseCode2FAEnabled},
		{http.StatusOK, ServiceCodeTwoFactor, CaseCode2FASetupInitiated},
		{http.StatusUnauthorized, ServiceCodeTwoFactor, CaseCode2FAInvalidCode},
		{http.StatusUnauthorized, ServiceCodeTwoFactor, CaseCode2FARecoveryCodeInvalid},
	}

	for _, c := range cases {
		code := BuildResponseCode(c.httpStatus, c.serviceCode, c.caseCode)
		gotHTTP, gotService, gotCase := ParseResponseCode(code)
		assert.Equal(t, c.httpStatus, gotHTTP, "httpStatus round-trip for (%d,%s,%s)", c.httpStatus, c.serviceCode, c.caseCode)
		assert.Equal(t, c.serviceCode, gotService, "serviceCode round-trip for (%d,%s,%s)", c.httpStatus, c.serviceCode, c.caseCode)
		assert.Equal(t, c.caseCode, gotCase, "caseCode round-trip for (%d,%s,%s)", c.httpStatus, c.serviceCode, c.caseCode)
	}
}

// ---------------------------------------------------------------------------
// Service code constant completeness
// ---------------------------------------------------------------------------

func TestServiceCodeConstants(t *testing.T) {
	// Verify every service code has a unique value so no two constants collide.
	seen := make(map[string]string)
	codes := map[string]string{
		"Common":                      ServiceCodeCommon,
		"Auth":                        ServiceCodeAuth,
		"Transaction":                 ServiceCodeTransaction,
		"Withdrawal":                  ServiceCodeWithdrawal,
		"User":                        ServiceCodeUser,
		"Admin":                       ServiceCodeAdmin,
		"Merchant":                    ServiceCodeMerchant,
		"Setting":                     ServiceCodeSetting,
		"Role":                        ServiceCodeRole,
		"Permission":                  ServiceCodePermission,
		"NotificationTemplate":        ServiceCodeNotificationTemplate,
		"NotificationTemplateChannel": ServiceCodeNotificationTemplateChannel,
		"Notification":                ServiceCodeNotification,
		"IPWhitelist":                 ServiceCodeIPWhitelist,
		"ApiKey":                      ServiceCodeApiKey,
		"Deposit":                     ServiceCodeDeposit,
		"Wallet":                      ServiceCodeWallet,
		"Phone":                       ServiceCodePhone,
		"Email":                       ServiceCodeEmail,
		"TwoFactor":                   ServiceCodeTwoFactor,
	}
	for name, code := range codes {
		if prev, exists := seen[code]; exists {
			t.Errorf("ServiceCode collision: %q and %q both have value %q", name, prev, code)
		}
		seen[code] = name
	}
}
