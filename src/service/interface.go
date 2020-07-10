package service

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/flightlogteam/api-gateway/models"
)

type IGatewayService interface {
	// Validate a user-token
	ValidateToken(token string) (jwt.Claims, error)

	// Renew a token from a user
	RenewToken(refreshToken string) string

	// Issue a new token. Same as Logging in
	IssueToken(userName string, password string) (string, error)

	// Authorize a user to a certain resource
	Authorize(resource string,  method string, token string) bool

	AuthorizeWithoutToken(resource string, method string) bool

	// Activate a user. This is done by the url given on Email
	ActivateUser(userId string) error

	RegisterUser(userData models.UserRegistration) (int, error)
}