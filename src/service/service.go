package service

import (
	"crypto/rsa"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/flightlogteam/api-gateway/repository"
	"io/ioutil"
	"log"
	"regexp"
	"time"
)

type GatewayService struct {
	signingKey *rsa.PrivateKey
	validationKey *rsa.PublicKey
	casbinEnforcer *casbin.Enforcer
	userRepository repository.IUserServiceRepository
}

func NewGatewayService(
		publicKeyPath string,
		privateKeyPath string,
		storageAdapter persist.Adapter,
		userRepository repository.IUserServiceRepository,
	) IGatewayService {

	// Load certificates into memory
	signingKey, validationKey := getSigningKeys(privateKeyPath, publicKeyPath)

	return &GatewayService{
		signingKey: signingKey,
		validationKey: validationKey,
		casbinEnforcer: createCasbinEnforcer(storageAdapter),
		userRepository: userRepository,
	}
}


func getSigningKeys(privateKeyPath string, publicKeyPath string) (*rsa.PrivateKey, *rsa.PublicKey) {
	var signBytes, verifyBytes []byte
	var signKey *rsa.PrivateKey
	var verifyKey *rsa.PublicKey
	var err error

	if signBytes, err = ioutil.ReadFile(privateKeyPath); err != nil {
		log.Fatalf("Unable to read PrivateKey: %v", err)
	}

	if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
		log.Fatalf("Unable to parse PrivateKey: %v", err)
	}

	if verifyBytes, err = ioutil.ReadFile(publicKeyPath); err != nil {
		log.Fatalf("Unable to read PublicKey: %v", err)
	}

	if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
		log.Fatalf("Unable to parse PublicKey: %v", err)
	}
	return signKey, verifyKey
}

func (k * GatewayService) ValidateToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k.validationKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims, err
}

func (k * GatewayService) RenewToken(refreshToken string) string {
	return ""
}

func (k * GatewayService) IssueToken(userCredential string, password string) (string, error) {
	// Determine if user-credential is an email or not
	isEmail, _ := regexp.MatchString(`^([a-zA-Z0-9_\-\.]+)@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.)|(([a-zA-Z0-9\-]+\.)+))([a-zA-Z]{2,4}|[0-9]{1,3})(\]?)$`, userCredential)

	var email, username string

	if isEmail {
		email = userCredential
	} else {
		username = userCredential
	}

	// Do a LOGIN-Request

	user, err := k.userRepository.LoginUser(username, email, password)

	// If we could not login. We dont issue a token
	if err != nil {
		return "", err
	}

	token, err := k.createLoginToken(user.Role, user.UserId)

	return token, err
}

func (k * GatewayService) Authorize(resource string, method string, tokenString string) bool {
	// Get role and userId
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k.validationKey, nil
	})

	if err != nil {
		return false
	}
	userID := token.Claims.(jwt.MapClaims)["UserID"]
	role := token.Claims.(jwt.MapClaims)["Role"]

	if role == "" {
		role = "anonymous"
	}

	res, err := k.casbinEnforcer.Enforce(role, userID, resource, method)

	if err != nil {
		log.Printf("Authorization failed with following error: %v", err)
		return false
	}
	// Enforce

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

// CreateVerificationToken creates a token used in an verification-email
func (k * GatewayService) createLoginToken(role string, ID string) (string, error) {
	expiration := time.Now().Add(time.Second * time.Duration(3600)).Unix()

	// Lets keep the token quite light
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{ExpiresAt: expiration},
		UserID:         ID,
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(k.signingKey)
}

type VerificationClaims struct {
	jwt.StandardClaims
	UserID string
}

type Claims struct {
	Role string
	UserID    string
	jwt.StandardClaims
}
