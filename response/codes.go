package response

// Service codes (2 digits: 00-06)
const (
	ServiceCodeCommon                      = "00" // Common/General services
	ServiceCodeAuth                        = "01" // Authentication service
	ServiceCodeTransaction                 = "02" // Transaction service
	ServiceCodeWithdrawal                  = "03" // Withdrawal service
	ServiceCodeUser                        = "04" // User service
	ServiceCodeAdmin                       = "05" // Admin service
	ServiceCodeMerchant                    = "06" // Merchant service
	ServiceCodeSetting                     = "07" // Setting service
	ServiceCodeRole                        = "08" // Role service
	ServiceCodePermission                  = "09" // Permission service
	ServiceCodeNotificationTemplate        = "10" // Notification template service
	ServiceCodeNotificationTemplateChannel = "11" // Notification template channel service
	ServiceCodeNotification                = "12" // Notification service
	ServiceCodeIPWhitelist                 = "13" // IP Whitelist service
	ServiceCodeApiKey                      = "14" // API Key service
	ServiceCodeDeposit                     = "15" // Deposit service
	ServiceCodeWallet                      = "16" // Wallet service
	ServiceCodePhone                       = "17" // Phone / phone verification service
	ServiceCodeEmail                       = "18" // Email / email verification service
	ServiceCodeTwoFactor                   = "19" // Two-factor authentication service
)

// Case codes (2 digits: 01-99)
const (
	// Success cases (01-10)
	CaseCodeSuccess            = "01" // General success
	CaseCodeCreated            = "02" // Resource created
	CaseCodeUpdated            = "03" // Resource updated
	CaseCodeDeleted            = "04" // Resource deleted
	CaseCodeRetrieved          = "05" // Resource retrieved
	CaseCodeListRetrieved      = "06" // List retrieved
	CaseCodeLoginSuccess       = "07" // Login successful
	CaseCodeLogoutSuccess      = "08" // Logout successful
	CaseCodePasswordChanged    = "09" // Password changed
	CaseCodeOperationCompleted = "10" // Operation completed

	// Validation errors (11-20)
	CaseCodeValidationError  = "11" // General validation error
	CaseCodeRequiredField    = "12" // Required field missing
	CaseCodeInvalidFormat    = "13" // Invalid format
	CaseCodeInvalidValue     = "14" // Invalid value
	CaseCodeDuplicateEntry   = "15" // Duplicate entry
	CaseCodeInvalidEmail     = "16" // Invalid email format
	CaseCodeInvalidPassword  = "17" // Invalid password
	CaseCodePasswordTooShort = "18" // Password too short
	CaseCodeInvalidDate      = "19" // Invalid date format
	CaseCodeInvalidRange     = "20" // Invalid range

	// Authentication errors (21-30)
	CaseCodeUnauthorized       = "21" // Unauthorized access
	CaseCodeInvalidToken       = "22" // Invalid token
	CaseCodeTokenExpired       = "23" // Token expired
	CaseCodeInvalidCredentials = "24" // Invalid credentials
	CaseCodeAccountLocked      = "25" // Account locked
	CaseCodeAccountDisabled    = "26" // Account disabled
	CaseCodePermissionDenied   = "27" // Permission denied
	CaseCodeSessionExpired     = "28" // Session expired
	CaseCodeTwoFactorRequired  = "29" // Two-factor authentication required
	CaseCodeInvalidOTP         = "30" // Invalid OTP

	// Not found errors (31-40)
	CaseCodeNotFound                            = "31" // Resource not found
	CaseCodeUserNotFound                        = "32" // User not found
	CaseCodeAdminNotFound                       = "33" // Admin not found
	CaseCodeMerchantNotFound                    = "34" // Merchant not found
	CaseCodeTransactionNotFound                 = "35" // Transaction not found
	CaseCodeSettingNotFound                     = "36" // Setting not found
	CaseCodeRoleNotFound                        = "37" // Role not found
	CaseCodeNotificationTemplateNotFound        = "38" // Notification template not found
	CaseCodeNotificationTemplateChannelNotFound = "39" // Notification template channel not found
	CaseCodeNotificationNotFound                = "40" // Notification not found
	CaseCodeMethodNotFound                      = "41" // Method not found
	CaseCodeRouteNotFound                       = "42" // Route not found
	CaseCodeResourceNotFound                    = "43" // General resource not found
	CaseCodeApiKeyNotFound                      = "44" // API key not found

	// Business logic errors (45-53)
	CaseCodeInsufficientBalance = "45" // Insufficient balance
	CaseCodeInvalidAmount       = "46" // Invalid amount
	CaseCodeTransactionFailed   = "47" // Transaction failed
	CaseCodeLimitExceeded       = "48" // Limit exceeded
	CaseCodeInvalidStatus       = "49" // Invalid status
	CaseCodeOperationNotAllowed = "50" // Operation not allowed
	CaseCodeAlreadyProcessed    = "51" // Already processed
	CaseCodePendingTransaction  = "52" // Pending transaction
	CaseCodeExpiredTransaction  = "53" // Expired transaction
	CaseCodeInvalidCurrency     = "54" // Invalid currency

	// Server errors (55-63)
	CaseCodeInternalError        = "55" // Internal server error
	CaseCodeDatabaseError        = "56" // Database error
	CaseCodeExternalServiceError = "57" // External service error
	CaseCodeTimeout              = "58" // Request timeout
	CaseCodeServiceUnavailable   = "59" // Service unavailable
	CaseCodeMaintenance          = "60" // Under maintenance
	CaseCodeRateLimitExceeded    = "61" // Rate limit exceeded
	CaseCodeConfigurationError   = "62" // Configuration error
	CaseCodeEncryptionError      = "63" // Encryption error
	CaseCodeDecryptionError      = "64" // Decryption error

	// Conflict errors (65-69)
	CaseCodeConflict               = "65" // General conflict
	CaseCodeResourceExists         = "66" // Resource already exists
	CaseCodeConcurrentModification = "67" // Concurrent modification
	CaseCodeVersionMismatch        = "68" // Version mismatch
	CaseCodeStateConflict          = "69" // State conflict

	// Email change (70-74)
	CaseCodeEmailChangeRequested   = "70" // Email change OTP sent
	CaseCodeEmailChangeVerified    = "71" // Email change confirmed
	CaseCodeEmailChangeCancelled   = "72" // Email change cancelled
	CaseCodeEmailAlreadyUsed       = "73" // Email already registered to another account
	CaseCodeEmailVerificationFailed = "74" // Email OTP verification failed

	// Phone change (75-79)
	CaseCodePhoneChangeRequested   = "75" // Phone change OTP sent
	CaseCodePhoneChangeVerified    = "76" // Phone change confirmed
	CaseCodePhoneChangeCancelled   = "77" // Phone change cancelled
	CaseCodePhoneAlreadyUsed       = "78" // Phone already registered to another account
	CaseCodePhoneVerificationFailed = "79" // Phone OTP verification failed

	// Two-factor authentication (80-89)
	CaseCode2FAEnabled             = "80" // 2FA enabled
	CaseCode2FADisabled            = "81" // 2FA disabled
	CaseCode2FASetupInitiated      = "82" // 2FA setup started (secret + QR generated)
	CaseCode2FASetupVerified       = "83" // 2FA setup confirmed with valid OTP
	CaseCode2FAInvalidCode         = "84" // 2FA code invalid or expired
	CaseCode2FACodeRequired        = "85" // 2FA code missing from request
	CaseCode2FANotEnabled          = "86" // 2FA not enabled on account
	CaseCode2FAAlreadyEnabled      = "87" // 2FA already enabled
	CaseCode2FARecoveryCodeUsed    = "88" // 2FA bypass via recovery code
	CaseCode2FARecoveryCodeInvalid = "89" // Recovery code invalid or already used
)

// BuildResponseCode builds a response code from HTTP status, service code, and case code
// Format: HTTP_STATUS_CODE (3 digits) + SERVICE_CODE (2 digits) + CASE_CODE (2 digits)
// Example: 2010301 = HTTP 201 + Service 03 (Withdrawal) + Case 01 (Success)
func BuildResponseCode(httpStatus int, serviceCode, caseCode string) int {
	// Convert HTTP status to 3-digit string (e.g., 200 -> "200", 404 -> "404")
	httpStatusStr := ""
	if httpStatus < 10 {
		httpStatusStr = "00" + string(rune(httpStatus+'0'))
	} else if httpStatus < 100 {
		httpStatusStr = "0" + string(rune(httpStatus/10+'0')) + string(rune(httpStatus%10+'0'))
	} else {
		httpStatusStr = string(rune(httpStatus/100+'0')) + string(rune((httpStatus/10)%10+'0')) + string(rune(httpStatus%10+'0'))
	}

	// Combine: HTTP_STATUS (3) + SERVICE (2) + CASE (2) = 7 digits
	codeStr := httpStatusStr + serviceCode + caseCode

	// Convert to int
	var code int
	for _, char := range codeStr {
		code = code*10 + int(char-'0')
	}

	return code
}

// ParseResponseCode parses a response code into its components
func ParseResponseCode(code int) (httpStatus int, serviceCode, caseCode string) {
	codeStr := ""
	temp := code
	for temp > 0 {
		codeStr = string(rune(temp%10+'0')) + codeStr
		temp /= 10
	}

	if len(codeStr) >= 7 {
		// Extract HTTP status (first 3 digits)
		httpStatus = int(codeStr[0]-'0')*100 + int(codeStr[1]-'0')*10 + int(codeStr[2]-'0')
		// Extract service code (next 2 digits)
		serviceCode = codeStr[3:5]
		// Extract case code (last 2 digits)
		caseCode = codeStr[5:7]
	}

	return httpStatus, serviceCode, caseCode
}
