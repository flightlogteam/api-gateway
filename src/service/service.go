package service

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"regexp"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
	"github.com/flightlogteam/api-gateway/src/common"
	"github.com/flightlogteam/api-gateway/src/models"
	"github.com/flightlogteam/api-gateway/src/repository"
	"github.com/golang-jwt/jwt"
	"github.com/klyngen/golang-oidc-discovery"
	"github.com/pkg/errors"
)

type GatewayService struct {
	casbinEnforcer  *casbin.Enforcer
	userRepository  repository.IUserServiceRepository
	discoveryClient *oidcdiscovery.OidcDiscoveryClient
	publicKeys      []oidcdiscovery.PublicKey
}

func NewGatewayService(
	storageAdapter persist.Adapter,
	userRepository repository.IUserServiceRepository,
	authenticationProvider string,
) IGatewayService {
	discoveryClient, err := oidcdiscovery.NewOidcDiscoveryClient(authenticationProvider)

	if err != nil {
		log.Fatalf("Could not discover any auth-provider on %v, with error %v", authenticationProvider, err)
	}

	log.Println("Trying to log the jwks-url", discoveryClient.DiscoveryDocument().JwksURI, discoveryClient.DiscoveryDocument().Issuer)
	publicKeys, err := discoveryClient.GetCertificates()

	if err != nil {
		log.Fatalf("Unable to parse keys of token provider. Gateway cannot function", authenticationProvider)
	}

	return &GatewayService{
		casbinEnforcer:  createCasbinEnforcer(storageAdapter),
		userRepository:  userRepository,
		discoveryClient: discoveryClient,
		publicKeys:      publicKeys,
	}
}

func (k *GatewayService) getPublicKey(token *jwt.Token) (*rsa.PublicKey, error) {
	cert := ""

	for _, key := range k.publicKeys {
		if token.Header["kid"] == key.Kid {
			cert = key.GetCertificate()
		}
	}
	log.Println(cert)

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
}

func (k *GatewayService) ValidateToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k.getPublicKey(token)
	})

	if err != nil {
		return nil, errors.Wrap(err, "Unable to validate token")
	}

	var expiration time.Time
	switch iat := token.Claims.(jwt.MapClaims)["exp"].(type) {
	case float64:
		expiration = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		expiration = time.Unix(v, 0)
	default:
		return nil, errors.New("Invalid time. Token is garbage")
	}

	if expiration.Before(time.Now()) {
		return nil, errors.New("Expired token")
	}

	return token.Claims, err
}

// RegisterUser Deprecated
func (k *GatewayService) RegisterUser(userData models.UserRegistration) (int, error) {
	// Validate the data
	// Is this an valid email?
	log.Println(userData)
	isEmail, _ := regexp.MatchString(`^([a-zA-Z0-9_\-\.]+)@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.)|(([a-zA-Z0-9\-]+\.)+))([a-zA-Z]{2,4}|[0-9]{1,3})(\]?)$`, userData.Email)

	if userData.PrivacyLevel > 2 || userData.PrivacyLevel < 0 {
		return 0, common.ServiceNoSuchPrivacyLevel
	}

	if !(len(userData.Username) > 0) || !(len(userData.FirstName) > 0) || !(len(userData.LastName) > 0) || !isEmail {
		return 0, common.ServiceMissingRequiredData
	}

	userResponse, err := k.userRepository.RegisterUser("", userData.FirstName, userData.LastName, userData.Email, userData.Username, userData.PrivacyLevel)

	return userResponse, err
}

func (k *GatewayService) RenewToken(refreshToken string) string {
	return ""
}

func (k *GatewayService) Authorize(resource string, method string, tokenString string) bool {

	role := ""

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k.getPublicKey(token)
	})

	if err != nil {
		return false
	}

	userID := token.Claims.(jwt.MapClaims)["sub"]
	roles, err := k.casbinEnforcer.GetRolesForUser(userID.(string))
	if len(roles) > 0 {
		role = roles[0]
	} else {
		k.casbinEnforcer.AddRoleForUser(userID.(string), "default")
	}
	res, err := k.casbinEnforcer.Enforce(role, userID, resource, method)

	if err != nil {
		log.Printf("Authorization failed with following error: %v", err)
		return false
	}

	return res
}

func createCasbinEnforcer(persist persist.Adapter) *casbin.Enforcer {
	cs, _ := casbin.NewEnforcer("./model.conf", persist)

	err := cs.LoadPolicy()

	if err != nil {
		log.Fatalf("Error thrown when creating casbin enforcer \nCannot start the application due to the following error: %v \n", err)
	}

	cs.AddFunction("isOwner", isOwnerWrapper)
	cs.AddFunction("keyMatch3", util.KeyMatch4Func)

	cs.EnableAutoSave(true)

	return cs
}

// This wrapper is almost too thin
func (k *GatewayService) ActivateUser(userId string) error {

	err := k.userRepository.ActivateUser(userId)

	if err != nil {
		return err
	}

	_, err = k.casbinEnforcer.AddRoleForUser(userId, "default")

	return err

}

func (k *GatewayService) VerifyUser(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k.getPublicKey(token)
	})

	if err != nil {
		return false, errors.Wrap(err, "Could not read token")
	}

	userID := token.Claims.(jwt.MapClaims)["sub"]
	userActivated := token.Claims.(jwt.MapClaims)["email_verified"].(bool)

	user, err := k.userRepository.GetUserById(userID.(string))

	if user != nil {
		return true, nil
	}

	if userActivated {
		_, err = k.userRepository.RegisterUser(userID.(string), tokenParameter(token, "given_name").(string), tokenParameter(token, "family_name").(string), tokenParameter(token, "email").(string), "", 1)

		if err != nil {
			return false, errors.Wrap(err, "Unable to create the user")
		}

		k.userRepository.ActivateUser(userID.(string))
		return true, nil
	}

	return false, nil

}

func tokenParameter(token *jwt.Token, parameter string) interface{} {
	return token.Claims.(jwt.MapClaims)[parameter]
}

func (k *GatewayService) AuthorizeWithoutToken(resource string, method string) bool {
	res, err := k.casbinEnforcer.Enforce("", "anonymous", resource, method)
	if err != nil {
		return false
	}
	return res
}

type VerificationClaims struct {
	jwt.StandardClaims
	UserID string
}

type Claims struct {
	Role   string
	UserID string
	jwt.StandardClaims
}
