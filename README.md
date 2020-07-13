# api-gateway
Gateway service authorizing and authenticating

## Background
In the flightlog microservice architecture we have gone for a Gateway design. This means we have a casbin powered gateway in front of the other services to ensure authentication and authorization.
The gateway will also handle login and registering. Communication internally in the cluster will be using certificate authentication, to ensure that outside requests to internal services will be impossible.

## Getting started developing
### Install dependencies
* Docker / podman
* Skaffold
* Minikube (should work with openshift too)

```bash
skaffold dev
```
