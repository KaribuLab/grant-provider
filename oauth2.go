package grantprovider

// GetAccessTokenData es un ejemplo de payload para colocar en InvokeResponse.Data
// (respuestas tipo token OAuth2).
type GetAccessTokenData struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
	ExpiresIn    int    `json:"expires_in" validate:"required"`
}
