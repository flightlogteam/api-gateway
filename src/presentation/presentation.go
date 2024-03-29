package presentation

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/flightlogteam/api-gateway/src/service"
	"github.com/gorilla/mux"
	"github.com/klyngen/jsend"
)

type GatewayApi struct {
	router  *mux.Router
	service service.IGatewayService
	proxy   proxy
}

func NewGatewayApi(service service.IGatewayService, routes []ProxyRoute) GatewayApi {
	router := mux.NewRouter()

	protected := router.PathPrefix("/api").Subrouter()
	auth := router.PathPrefix("/auth").Subrouter()

	// Implements the proxy
	proxy := proxy{routes: routes}

	// Jsendify the default handlers
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(notAllowedHandler)

	// Create the API
	api := GatewayApi{router: router, service: service, proxy: proxy}

	auth.HandleFunc("/verify", api.verifyUserEndpoint).Methods("GET", "OPTIONS")

	protected.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		proxy.routeMessage(req.RequestURI, w, req)
	})

	// Middleware to require login for certain endpoints
	protected.Use(api.authMiddleware)

	router.Use(corsMiddleware)

	return api
}

func (a *GatewayApi) verifyUserEndpoint(writer http.ResponseWriter, request *http.Request) {
	log.Println("Starting verification")
	authorizationHeader := request.Header.Get("Authorization")
	tokenParts := strings.Split(authorizationHeader, "Bearer ")

	if len(tokenParts) < 2 {
		jsend.FormatResponse(writer, "Unauthorized", jsend.BadRequest)
	}

	token := tokenParts[1]

	log.Println(token)
	status, err := a.service.VerifyUser(token)

	if err != nil {
		log.Println(err)
		jsend.FormatResponse(writer, "Unable to verify the user", jsend.BadRequest)
		return
	}

	jsend.FormatResponse(writer, status, jsend.Success)
}

func (a *GatewayApi) StartAPI() {
	printRoutes(a.router)

	log.Printf("Started FlightLogger on port: %s", "61225")

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", "61225"), a.router)

	if err != nil {
		log.Fatalf("Unable to start the API due to the following error: \n %v", err)
	}
}

func (a *GatewayApi) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authorizationHeader := r.Header.Get("Authorization")

		if len(authorizationHeader) > 0 {
			token := strings.Split(authorizationHeader, "Bearer ")[1]
			if len(token) == 0 {
				jsend.FormatResponse(w, "You have no accesstoken. RTFM", jsend.UnAuthorized)
				return
			}

			if _, err := a.service.ValidateToken(token); err != nil {
				// 403
				jsend.FormatResponse(w, err.Error(), jsend.Forbidden)
				return
			}

			if a.service.Authorize(r.RequestURI, r.Method, token) {
				// Forward the request
				next.ServeHTTP(w, r)
				return
			}
		}

		if a.service.AuthorizeWithoutToken(r.RequestURI, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		jsend.FormatResponse(w, "Not authorized. You must log in to see this resource", jsend.Forbidden)
		return

	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	jsend.FormatResponse(w, "No such endpoint. RTFM", jsend.NotFound)
}

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
	jsend.FormatResponse(w, "Correct endpoint wrong method. RTFM", jsend.MethodNotAllowed)
}

func printRoutes(router *mux.Router) {
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			log.Println("ROUTE:", pathTemplate)
		}
		methods, err := route.GetMethods()
		if err == nil {
			log.Println("Methods:", strings.Join(methods, ","))
		}

		log.Println()
		return nil
	})

	if err != nil {
		log.Println(err)
	}
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("content-type", "application/json;charset=UTF-8")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
