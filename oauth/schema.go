package oauth

type Request struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	AcceptedTerms bool   `json:"accepted_terms"`
}

type Response struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
