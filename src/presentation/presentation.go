package presentation

import (
	"fmt"
	"github.com/flightlogteam/api-gateway/src/service"
	"github.com/gorilla/mux"
	"github.com/klyngen/jsend"
	"log"
	"net/http"
	"strings"
)

type GatewayApi struct {
	router *mux.Router
	service service.IGatewayService
	proxy proxy
}

func NewGatewayApi(service service.IGatewayService, routes []ProxyRoute) GatewayApi {
	router := mux.NewRouter()


	protected := router.PathPrefix("/api/protected").Subrouter()
	auth := router.PathPrefix("/auth").Subrouter()

	// Implements the proxy
	proxy := proxy{routes: routes}

	auth.HandleFunc("*", func(w http.ResponseWriter, req *http.Request) {
		proxy.routeMessage(req.RequestURI, w, req)
	})

	// Jsendify the default handlers
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(notAllowedHandler)


	// Create the API
	api := GatewayApi{router: router, service: service, proxy: proxy}
	api.mountAuthenticationRoutes(auth)
	//api.mountUserRoutes(protected)

	// Middleware to require login for certain endpoints
	protected.Use(api.authMiddleware)

	return api
}

func (a * GatewayApi) StartAPI() {
	printRoutes(a.router)

	log.Printf("Started FlightLogger on port: %s", "61225")

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", "61225"), a.router)

	if err != nil {
		log.Fatalf("Unable to start the API due to the following error: \n %v", err)
	}
	//a.router.Host
}

func (a *GatewayApi) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Will result in something like Bearer <TOKEN>
		authorizationHeader := r.Header.Get("Authorization")

		token := strings.Split(authorizationHeader, "Bearer ")[1]
		if len(token) == 0 {
			jsend.FormatResponse(w, "You have no accesstoken. RTFM", jsend.UnAuthorized)
			return
		}

		if _, err := a.service.ValidateToken(token); err == nil {
			if a.service.Authorize(r.RequestURI, r.Method, token) {
				// Forward the request
				next.ServeHTTP(w, r)
				return
			}
			// 403
			jsend.FormatResponse(w, "The requested resource is not available for your user / group", jsend.Forbidden)
			return
		} else {
			if a.service.AuthorizeWithoutToken(r.RequestURI, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			jsend.FormatResponse(w, "Not authorized. You must log in to see this resource", jsend.Forbidden)
			return
		}
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
