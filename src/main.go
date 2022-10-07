package main

import (
	"fmt"
	"log"
	"os"

	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/flightlogteam/api-gateway/src/presentation"
	"github.com/flightlogteam/api-gateway/src/repository"
	"github.com/flightlogteam/api-gateway/src/service"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get configuration from environment variables
	config := getConfiguration()
	serviceConfig := getServiceConfiguration()

	adapter, err := xormadapter.NewAdapter("mysql", config.createConnectionString())
	if err != nil {
		log.Fatalf("Unable to establish casbin-adapter: %v", err)
	}

	gatewayService := service.NewGatewayService(
		adapter,
		getUserService(serviceConfig.userServiceURL),
		serviceConfig.authenticationProvider,
	)

	routes := []presentation.ProxyRoute{
		{
			DestinationAddress: fmt.Sprintf("%s:%s", serviceConfig.userServiceURL, serviceConfig.userServiceRESTPort),
			Target:             "Users",
		},
		{
			DestinationAddress: fmt.Sprintf("%s:%s", serviceConfig.flightServiceURL, serviceConfig.flightServicePort),
			Target:             "Flights",
		},
	}

	log.Println(serviceConfig)

	api := presentation.NewGatewayApi(gatewayService, routes)
	api.StartAPI()
}

func (c *databaseConfiguration) createConnectionString() string {
	if len(c.hostname) > 0 { // Full config
		return fmt.Sprintf("%v:%v@tcp(%v:%v)/", c.username, c.password, c.hostname, c.port)
	}

	return fmt.Sprintf("%v:%v@/", c.username, c.password)
}

func getUserService(serviceURL string) repository.IUserServiceRepository {
	return repository.NewUserRepository(serviceURL)
}

func getConfiguration() databaseConfiguration {
	return databaseConfiguration{
		password: os.Getenv("DATABASE_PASSWORD"),
		username: os.Getenv("DATABASE_USERNAME"),
		port:     os.Getenv("DATABASE_PORT"),
		hostname: os.Getenv("DATABASE_HOSTNAME"),
	}
}

func getServiceConfiguration() serviceConfiguration {
	return serviceConfiguration{
		flightServiceURL:       os.Getenv("SERVICE_FLIGHTSERVICE_URL"),
		flightServicePort:      os.Getenv("SERVICE_FLIGHTSERVICE_PORT"),
		userServiceURL:         os.Getenv("SERVICE_USERSERVICE_URL"),
		userServicePort:        os.Getenv("SERVICE_USERSERVICE_PORT"),
		userServiceRESTPort:    os.Getenv("SERVICE_USERSERVICE_REST_PORT"),
		authenticationProvider: os.Getenv("AUTHENTICATION_PROVIDER"),
	}
}

type serviceConfiguration struct {
	flightServiceURL       string
	flightServicePort      string
	userServiceURL         string
	userServicePort        string
	userServiceRESTPort    string
	authenticationProvider string
}

type databaseConfiguration struct {
	password string
	username string
	port     string
	hostname string
	baseURL  string
}

func (c *databaseConfiguration) IsValidConfiguration() bool {
	if len(c.password) > 0 && len(c.username) > 0 {
		return true
	}
	return false
}
