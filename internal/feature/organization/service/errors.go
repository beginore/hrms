package service

import "errors"

var (
	ErrInvalidOrganizationName           = CreateOrganizationValidationError{"organization name is required"}
	ErrInvalidOrganizationNameSize       = CreateOrganizationValidationError{"organization name must be between 1 and 50 characters"}
	ErrInvalidOrganizationNameWhiteSpace = CreateOrganizationValidationError{"organization name must not have leading or trailing spaces"}
	ErrInvalidOrganizationNameSymbols    = CreateOrganizationValidationError{"organization name contains invalid characters"}

	ErrInvalidOrganizationVATID           = CreateOrganizationValidationError{"VAT number is required"}
	ErrInvalidOrganizationVATIDSize       = CreateOrganizationValidationError{"VAT number must be between 8 and 30 characters"}
	ErrInvalidOrganizationVATIDWhiteSpace = CreateOrganizationValidationError{"VAT number must not have leading or trailing spaces"}
	ErrInvalidOrganizationVATIDSymbols    = CreateOrganizationValidationError{"VAT number contains invalid characters"}

	ErrInvalidOrganizationDescription           = CreateOrganizationValidationError{"organization description is required"}
	ErrInvalidOrganizationDescriptionSize       = CreateOrganizationValidationError{"organization description must be between 15 and 250 characters"}
	ErrInvalidOrganizationDescriptionWhiteSpace = CreateOrganizationValidationError{"organization description must not have leading or trailing spaces"}
	ErrInvalidOrganizationDescriptionSymbols    = CreateOrganizationValidationError{"organization description contains invalid characters"}

	ErrInvalidOrganizationAddress           = CreateOrganizationValidationError{"street address is required"}
	ErrInvalidOrganizationAddressSize       = CreateOrganizationValidationError{"street address must be between 2 and 50 characters"}
	ErrInvalidOrganizationAddressWhiteSpace = CreateOrganizationValidationError{"street address must not have leading or trailing spaces"}
	ErrInvalidOrganizationAddressSymbols    = CreateOrganizationValidationError{"street address contains invalid characters"}

	ErrInvalidCityID = CreateOrganizationValidationError{"city ID must be a positive number"}

	ErrInvalidFirstName             = CreateOrganizationValidationError{"first name is required"}
	ErrInvalidFirstNameSize         = CreateOrganizationValidationError{"first name must be between 1 and 30 characters"}
	ErrInvalidFirstNameSymbols      = CreateOrganizationValidationError{"first name contains invalid characters"}
	ErrInvalidLastName              = CreateOrganizationValidationError{"last name is required"}
	ErrInvalidLastNameSize          = CreateOrganizationValidationError{"last name must be between 1 and 30 characters"}
	ErrInvalidLastNameSymbols       = CreateOrganizationValidationError{"last name contains invalid characters"}
	ErrInvalidPhoneNumber           = CreateOrganizationValidationError{"phone number is required"}
	ErrInvalidPhoneNumberStart      = CreateOrganizationValidationError{"phone number must start with +"}
	ErrInvalidPhoneNumberSize       = CreateOrganizationValidationError{"phone number must be between 6 and 20 characters"}
	ErrInvalidPhoneNumberWhiteSpace = CreateOrganizationValidationError{"phone number must not contain spaces"}
	ErrInvalidPhoneNumberSymbols    = CreateOrganizationValidationError{"phone number must contain only digits after +"}

	ErrInvalidEmailAddress           = CreateOrganizationValidationError{"email address is required"}
	ErrInvalidEmailAddressSize       = CreateOrganizationValidationError{"email address must be between 7 and 255 characters"}
	ErrInvalidEmailAddressWhiteSpace = CreateOrganizationValidationError{"email address must not contain spaces"}
	ErrInvalidEmailAddressFormat     = CreateOrganizationValidationError{"email address format is invalid"}

	ErrInvalidPassword = CreateOrganizationValidationError{"password must be at least 8 characters and contain a special character"}

	ErrInvalidPrivacyPolicyAccepted      = CreateOrganizationValidationError{"privacy policy must be accepted"}
	ErrInvalidTermsAndConditionsAccepted = CreateOrganizationValidationError{"terms and conditions must be accepted"}

	ErrInvalidOTP          = errors.New("invalid or incorrect OTP code")
	ErrOTPExpired          = errors.New("OTP code has expired")
	ErrTooManyOTPAttempts  = errors.New("too many failed OTP attempts – try again later")
	ErrUserAlreadyVerified = errors.New("user is already verified")
	ErrUserNotFound        = errors.New("user not found for this email")
)
