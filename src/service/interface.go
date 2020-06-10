package service

import "github.com/dgrijalva/jwt-go"

type IGatewayService interface {
	// Validate a user-token
	ValidateToken(token string) (jwt.Claims, error)

	// Renew a token from a user
	RenewToken(refreshToken string) string

	// Issue a new token. Same as Logging in
	IssueToken(userName string, password string) (string, error)

	// Authorize a user to a certain resource
	Authorize(resource string,  method string, token string) bool
}