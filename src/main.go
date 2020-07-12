package main

import (
	"fmt"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/flightlogteam/api-gateway/src/presentation"
	"github.com/flightlogteam/api-gateway/src/repository"
	"github.com/flightlogteam/api-gateway/src/service"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

func main() {
	// Get configuration from environment variables
	config := getConfiguration()

	log.Print(config)

	// Create the casbin-adapter

	adapter, err := xormadapter.NewAdapter("mysql", config.createConnectionString())
	if err != nil {
		log.Fatalf("Unable to establish casbin-adapter: %v", err)
	}


	gatewayService := service.NewGatewayService("/etc/certificates/fly.rsa.pub",
		"/etc/certificates/fly.rsa",
		adapter,
		getUserService())

	routes := []presentation.ProxyRoute{
		{
			DestinationAddress: fmt.Sprintf("http://%s:%s", os.Getenv("USERSERVICE_URL"), "61226"),
			Target:             "Users",
		},
		{
			DestinationAddress: "http://localhost:61227", // TODO: replace localhost
			Target:             "Flights",
		},
		{
			DestinationAddress: "http://localhost:61228", // TODO: replace localhost
			Target:             "Locations",
		},
	}

	api := presentation.NewGatewayApi(gatewayService, routes)
	api.StartAPI()
}

func (c * databaseConfiguration) createConnectionString() string {
	if len(c.hostname) > 0 { // Full config
		return fmt.Sprintf("%v:%v@tcp(%v:%v)/", c.username, c.password, c.hostname, c.port)
	}

	return fmt.Sprintf("%v:%v@/", c.username, c.password)
}

func getUserService() repository.IUserServiceRepository {
	return repository.NewUserRepository(os.Getenv("USERSERVICE_URL"))

}

func getConfiguration() databaseConfiguration {
	return databaseConfiguration{
		password: os.Getenv("DATABASE_PASSWORD"),
		username: os.Getenv("DATABASE_USERNAME"),
		port: os.Getenv("DATABASE_PORT"),
		hostname: os.Getenv("DATABASE_HOSTNAME"),
	}
}

type databaseConfiguration struct {
	password string
	username string
	port string
	hostname string
}



func (c * databaseConfiguration) IsValidConfiguration() bool {
	if len(c.password) > 0 && len(c.username) > 0 {
		return true
	}
	return false
}
