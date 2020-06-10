package presentation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProxyRoute(t * testing.T) {
	// Create the setup

	routeA := ProxyRoute{
		DestinationAddress: "some.Clusterified.address:8080",
		Target: "Users",
	}

	assert.True(t, routeA.isDestination("http://localhost/Users/something/something"))
	assert.True(t, routeA.isDestination("http://localhost/Users/"))
	assert.True(t, routeA.isDestination("http://localhost/Users"))
	assert.True(t, routeA.isDestination("http://localhost:8080/Users"))
	assert.True(t, routeA.isDestination("localhost:8080/Users"))

	assert.False(t, routeA.isDestination("http://localhost/NotUsers"))
	assert.False(t, routeA.isDestination("http://localhost/users"))
	assert.False(t, routeA.isDestination("http://localhost/users"))
}
