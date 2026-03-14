package service

type ConsentItemResponse struct {
	DocumentType   string  `json:"documentType"`
	CurrentVersion *string `json:"currentVersion"`
	LatestVersion  string  `json:"latestVersion"`
	URL            string  `json:"url"`
}

type ConsentItem struct {
	DocumentType string `json:"DocumentType" validate:"required"`
	Version      string `json:"version" validate:"required"`
}

type RenewConsentRequest struct {
	PrivacyPolicyAccepted      bool `json:"privacyPolicyAccepted" validate:"required"`
	TermsAndConditionsAccepted bool `json:"termsAndConditionsAccepted" validate:"required"`
}

type ConsentValidationError struct {
	DocumentType   string  `json:"documentType" validate:"required"`
	CurrentVersion *string `json:"currentVersion"`
	LatestVersion  string  `json:"latestVersion"`
}

type ConsentValidationResponse struct {
	Valid  bool                     `json:"valid"`
	Errors []ConsentValidationError `json:"errors,omitempty"`
}
