package service

type CreateOrganizationRequest struct {
	Firstname                  string `json:"firstname"`
	Lastname                   string `json:"lastname"`
	PhoneNumber                string `json:"phoneNumber"`
	Email                      string `json:"email"`
	Password                   string `json:"password"`
	OrganizationName           string `json:"organizationName"`
	OrganizationDescription    string `json:"organizationDescription"`
	Vat                        string `json:"vat"`
	StreetAddress              string `json:"streetAddress"`
	CityId                     int32  `json:"cityId"`
	PrivacyPolicyAccepted      bool   `json:"privacyPolicyAccepted"`
	TermsAndConditionsAccepted bool   `json:"termsAndConditionsAccepted"`
}
type CreateOrganizationResponse struct {
	OrganizationId string `json:"organizationId"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
