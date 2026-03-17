package service

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type TokenResponse struct {
	IdToken      string `json:"idToken"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int32  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}
