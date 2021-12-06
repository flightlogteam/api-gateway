package repository

import "github.com/flightlogteam/api-gateway/src/models"

type IUserServiceRepository interface {
	// Activate the user, so that the user can log in
	ActivateUser(userId string) error

	// Login the user
	LoginUser(username string, email string, password string) (*models.User, error)

	// RegisterUser registers a user
	RegisterUser(id string, firstName string, lastName string, email string, username string, privacyLevel int) (int, error)

	GetUserById(userId string) (*models.UserWithPrivacy, error)
}
