package repository

import "github.com/flightlogteam/api-gateway/models"

type IUserServiceRepository interface {
	ActivateUser(userId string) error
	LoginUser(username string, email string, password string) (*models.User, error)
}
