package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/flightlogteam/api-gateway/src/models"
	"github.com/flightlogteam/userservice/grpc/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewUserRepository(serviceUrl string) IUserServiceRepository {
	log.Println("Initializing the UserRepository")

	if len(serviceUrl) == 0 {
		log.Println("No user-service is configured. Repository will not work")
		return &UserRepository{}
	}

	log.Printf("attempting to dial %v", serviceUrl)
	connection, err := dialUserService(serviceUrl)

	if err != nil {
		log.Printf("Unable to dial userService during userActivation due to error: %v", err)
	}

	log.Println("Connected to a userservice")

	return &UserRepository{
		serviceUrl:  serviceUrl,
		userService: userservice.NewUserServiceClient(connection),
	}
}

type UserRepository struct {
	serviceUrl  string
	userService userservice.UserServiceClient
}

func (u *UserRepository) RegisterUser(firstName string, lastName string, email string, username string, password string, privacyLevel int) (int, error) {
	pvl := userservice.CreateUserRequest_PrivacyLevel(privacyLevel)

	requestBody := userservice.CreateUserRequest{
		Username:  username,
		Email:     email,
		Firstname: firstName,
		Lastname:  lastName,
		Level:     pvl,
		Password:  password,
	}

	response, err := u.userService.RegisterUser(context.Background(), &requestBody)

	if err != nil {
		return 0, err
	}

	return int(response.Status), nil

}

func dialUserService(serviceUrl string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s:%s", serviceUrl, "61226"), grpc.WithInsecure())
	//return grpc.Dial(fmt.Sprintf("%s:%s", serviceUrl, "61226"), grpc.WithTransportCredentials(createCredentials()))
}

func (u *UserRepository) ActivateUser(userId string) error {
	response, err := u.userService.ActivateUser(context.Background(), &userservice.ActivateUserRequest{UserId: userId})

	if err != nil {
		return err
	}

	if !response.Status {
		return errors.New("user not activated")
	}

	return nil
}

func (u *UserRepository) LoginUser(username string, email string, password string) (*models.User, error) {

	loginRequest := userservice.LoginRequest{Password: password}

	if len(username) > 0 {
		loginRequest.UserCredential = &userservice.LoginRequest_Username{Username: username}
	} else {
		loginRequest.UserCredential = &userservice.LoginRequest_Email{Email: email}
	}

	response, err := u.userService.LoginUser(context.Background(), &loginRequest)

	if err != nil {
		return nil, err
	}

	switch response.Status {
	case userservice.LoginResponse_SUCCESS:
		return &models.User{
			UserId: response.UserId,
			Role:   response.Role,
		}, nil
	case userservice.LoginResponse_INVALID_CREDENTIALS:
		return nil, ErrorInvalidCredentials
	case userservice.LoginResponse_NOT_ACTIVATED:
		return nil, ErrorUserNotActivated
	}

	return nil, ErrorInternalServer
}

func createCredentials() credentials.TransportCredentials {
	creds, err := credentials.NewClientTLSFromFile("/etc/certificates/server.crt", "")
	if err != nil {
		log.Fatalf("Unable to start the Gateway due to missing certificates. Generate please: %v", err)
	}
	return creds
}
