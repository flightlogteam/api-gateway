package presentation

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/flightlogteam/api-gateway/src/models"
	"github.com/gorilla/mux"
	"github.com/klyngen/jsend"
	"log"
	"net/http"
)


type credentials struct {
	Username string
	Password string
}

func (f *GatewayApi) mountAuthenticationRoutes(router *mux.Router) {
	router.HandleFunc("/login", f.loginHandler).Methods("POST")
	router.HandleFunc("/verify", f.verifyUserAccount).Methods("GET")
	router.HandleFunc("/register", f.registerUser).Methods("POST")
}

// TODO: make this redirect to some GUI
func (f *GatewayApi) verifyUserAccount(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"]
	log.Printf("Get user for ID: %v", token)

	if len(token[0]) > 0 {
		claims, err := f.service.ValidateToken(token[0])

		if err != nil {
			log.Printf("Invalid token passed to service %v", token[0])
			jsend.FormatResponse(w, "Bad token", jsend.BadRequest)
			return
		}

		// "Parse" the claims
		userID := claims.(jwt.MapClaims)["UserID"]
		log.Println(claims, userID)

		if err = f.service.ActivateUser(userID.(string)); err != nil {
			log.Printf("Unable to activate userID %s, due to erro %v", userID, err)
			jsend.FormatResponse(w, "Could not activate the user", jsend.InternalServerError)
			return
		}

		jsend.FormatResponse(w, "User is activated", jsend.Success)

		return
	}

	jsend.FormatResponse(w, "No token is present. Are you trying to hack me!?", jsend.BadRequest)

}

func (f *GatewayApi) registerUser(w http.ResponseWriter, r *http.Request) {
	var payload models.UserRegistration

	if json.NewDecoder(r.Body).Decode(&payload) != nil {
		jsend.FormatResponse(w, "Unable to deserialize", jsend.BadRequest)
		return
	}

	result, err := f.service.RegisterUser(payload)

	if err != nil {
		jsend.FormatResponse(w, err.Error(), jsend.BadRequest)
		return
	}

	switch result {
	case 0:
		jsend.FormatResponse(w, "Success", jsend.Success)
		return
	case 1:
		jsend.FormatResponse(w, "Email already in use", jsend.BadRequest)
		return
	case 2:
		jsend.FormatResponse(w, "Username already in use", jsend.BadRequest)
		return
	}

	jsend.FormatResponse(w, "Unexpected issue during creation", jsend.InternalServerError)

}

func (f *GatewayApi) loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds credentials

	// If we cannot decode the request
	if json.NewDecoder(r.Body).Decode(&creds) != nil {
		jsend.FormatResponse(w, "Bad request data. RTFM", jsend.BadRequest)
		return
	}

	token, err := f.service.IssueToken(creds.Username, creds.Password)

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if err != nil {
		jsend.FormatResponse(w, err.Error(), jsend.UnAuthorized)
		return
	}

	jsend.FormatResponse(w, "Authenticated!", jsend.Success)
}
