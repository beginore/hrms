package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	reOrgName       = regexp.MustCompile(`^[\p{Latin}0-9&., ()'\/]+$`)
	reVAT           = regexp.MustCompile(`^[\p{Latin}0-9\-.]+$`)
	reDescription   = regexp.MustCompile(`^[\p{Latin}0-9\-.,+:;()—–&'"\/\s]+$`)
	reAddress       = regexp.MustCompile(`^[\p{Latin}0-9\-'\/\s]+$`)
	rePhone         = regexp.MustCompile(`^\+\d{5,19}$`)
	reEmail         = regexp.MustCompile(`^[a-zA-Z0-9\-'_.]+@[a-zA-Z0-9\-.]+\.[a-zA-Z]{2,}$`)
	reFirstLastName = regexp.MustCompile(`^[\p{Latin}\s\-']+$`)
)

type CreateOrganizationValidationError struct {
	Message string `json:"message"`
}

func (e CreateOrganizationValidationError) Error() string { return e.Message }

func ValidateOrganization(_ context.Context, req CreateOrganizationRequest) error {
	validators := []func() error{
		func() error { return validateOrgName(req.OrganizationName) },
		func() error { return validateVAT(req.Vat) },
		func() error { return validateDescription(req.OrganizationDescription) },
		func() error { return validateAddress(req.StreetAddress) },
		func() error { return validateCityID(req.CityId) },
		func() error { return validateFirstName(req.Firstname) },
		func() error { return validateLastName(req.Lastname) },
		func() error { return validatePhone(req.PhoneNumber) },
		func() error { return validateEmail(req.Email) },
		func() error { return validatePassword(req.Password) },
		func() error { return validatePrivacyPolicy(req.PrivacyPolicyAccepted) },
		func() error { return validateTermsAndConditions(req.TermsAndConditionsAccepted) },
	}

	var errs []error
	for _, v := range validators {
		if err := v(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func validateOrgName(name string) error {
	if name == "" {
		return ErrInvalidOrganizationName
	}
	if n := utf8.RuneCountInString(name); n < 1 || n > 50 {
		return ErrInvalidOrganizationNameSize
	}
	if strings.TrimSpace(name) != name {
		return ErrInvalidOrganizationNameWhiteSpace
	}
	if !reOrgName.MatchString(name) {
		return ErrInvalidOrganizationNameSymbols
	}
	return nil
}

func validateVAT(vat string) error {
	if vat == "" {
		return ErrInvalidOrganizationVATID
	}
	if n := utf8.RuneCountInString(vat); n < 8 || n > 30 {
		return ErrInvalidOrganizationVATIDSize
	}
	if strings.TrimSpace(vat) != vat {
		return ErrInvalidOrganizationVATIDWhiteSpace
	}
	if !reVAT.MatchString(vat) {
		return ErrInvalidOrganizationVATIDSymbols
	}
	return nil
}

func validateDescription(desc string) error {
	if desc == "" {
		return ErrInvalidOrganizationDescription
	}
	if n := utf8.RuneCountInString(desc); n < 15 || n > 250 {
		return ErrInvalidOrganizationDescriptionSize
	}
	if strings.TrimSpace(desc) != desc {
		return ErrInvalidOrganizationDescriptionWhiteSpace
	}
	if !reDescription.MatchString(desc) {
		return ErrInvalidOrganizationDescriptionSymbols
	}
	return nil
}

func validateAddress(addr string) error {
	if addr == "" {
		return ErrInvalidOrganizationAddress
	}
	if n := utf8.RuneCountInString(addr); n < 2 || n > 50 {
		return ErrInvalidOrganizationAddressSize
	}
	if strings.TrimSpace(addr) != addr {
		return ErrInvalidOrganizationAddressWhiteSpace
	}
	if !reAddress.MatchString(addr) {
		return ErrInvalidOrganizationAddressSymbols
	}
	return nil
}

func validateCityID(id int32) error {
	if id <= 0 {
		return ErrInvalidCityID
	}
	return nil
}

func validateFirstName(name string) error {
	if name == "" {
		return ErrInvalidFirstName
	}
	if n := utf8.RuneCountInString(name); n < 1 || n > 30 {
		return ErrInvalidFirstNameSize
	}
	if !reFirstLastName.MatchString(name) {
		return ErrInvalidFirstNameSymbols
	}
	return nil
}

func validateLastName(name string) error {
	if name == "" {
		return ErrInvalidLastName
	}
	if n := utf8.RuneCountInString(name); n < 1 || n > 30 {
		return ErrInvalidLastNameSize
	}
	if !reFirstLastName.MatchString(name) {
		return ErrInvalidLastNameSymbols
	}
	return nil
}

func validatePhone(phone string) error {
	if phone == "" {
		return ErrInvalidPhoneNumber
	}
	if !strings.HasPrefix(phone, "+") {
		return ErrInvalidPhoneNumberStart
	}
	if strings.ContainsAny(phone, " \t") {
		return ErrInvalidPhoneNumberWhiteSpace
	}
	if n := utf8.RuneCountInString(phone); n < 6 || n > 20 {
		return ErrInvalidPhoneNumberSize
	}
	if !rePhone.MatchString(phone) {
		return ErrInvalidPhoneNumberSymbols
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmailAddress
	}
	if strings.ContainsAny(email, " \t") {
		return ErrInvalidEmailAddressWhiteSpace
	}
	if n := utf8.RuneCountInString(email); n < 7 || n > 255 {
		return ErrInvalidEmailAddressSize
	}
	if !reEmail.MatchString(email) {
		return ErrInvalidEmailAddressFormat
	}
	return nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return ErrInvalidPassword
	}
	if !strings.ContainsAny(password, "!@#$%^&*") {
		return ErrInvalidPassword
	}
	return nil
}

func validatePrivacyPolicy(accepted bool) error {
	if !accepted {
		return ErrInvalidPrivacyPolicyAccepted
	}
	return nil
}

func validateTermsAndConditions(accepted bool) error {
	if !accepted {
		return ErrInvalidTermsAndConditionsAccepted
	}
	return nil
}
