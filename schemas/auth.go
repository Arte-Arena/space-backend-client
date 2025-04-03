package schemas

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AuthValidationResponse struct {
	UserID string `json:"user_id"`
	Valid  bool   `json:"valid"`
}
