package service

import (
	"github.com/golang-jwt/jwt"
)

type IGatewayService interface {
	// Validate a user-token
	ValidateToken(token string) (jwt.Claims, error)

	// Authorize a user to a certain resource
	Authorize(resource string, method string, token string) bool

	// AuthorizeWithoutToken is for guest users
	AuthorizeWithoutToken(resource string, method string) bool

	// Activate a user. This is done by the url given on Email
	ActivateUser(userId string) error

	// If a user does not exist in the database, make it
	VerifyUser(tokenString string) (bool, error)
}
