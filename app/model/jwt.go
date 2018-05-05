package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// A JWT is a JSON web token, and contains all the values necessary to create
// and validate tokens.
type JWT struct {
	UserRole   int    `json:"user_role"`
	UserID     string `json:"user_id"`
	UserEmail  string `json:"user_email"`
	Expiration int64  `json:"expiration"`
	Token      string `json:"token"`
}

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// InitializeKey initializes the public and private keys used to create and
// validate JSON web tokens.
func InitializeKey(privKeyPath, pubKeyPath string) {
	createJWTKeyFiles(privKeyPath, pubKeyPath)

	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// NewJWT returns a JWT for the given User.
func NewJWT(user *User) (JWT, error) {
	method := jwt.GetSigningMethod("RS256")
	exp := time.Now().Add(time.Minute * 3600).Unix()
	claims := jwt.MapClaims{
		"UserRole":  user.Role,
		"UserID":    user.Id.Hex(),
		"UserEmail": user.Email,
		"exp":       exp,
	}
	token := jwt.NewWithClaims(method, claims)

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return JWT{}, err
	}

	return JWT{
		UserRole:   user.Role,
		UserID:     string(user.Id),
		UserEmail:  user.Email,
		Expiration: exp,
		Token:      tokenString,
	}, nil
}

// NewJWTFromToken returns a JWT for the given token.
func NewJWTFromToken(token *jwt.Token) JWT {
	userRole := token.Claims.(jwt.MapClaims)["UserRole"].(float64)
	userID := token.Claims.(jwt.MapClaims)["UserID"].(string)
	userEmail := token.Claims.(jwt.MapClaims)["UserEmail"].(string)
	expiration := token.Claims.(jwt.MapClaims)["exp"].(float64)
	return JWT{
		UserRole:   int(userRole),
		UserID:     userID,
		UserEmail:  userEmail,
		Expiration: int64(expiration),
		Token:      token.Raw,
	}
}

// ValidateJWT validates a JSON web token, returning the token if it is
// indeed valid.
func ValidateJWT(t string) (*jwt.Token, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	switch err.(type) {
	case nil:
		if !token.Valid {
			return token, fmt.Errorf("Invalid token: %s\n", token.Raw)
		}
		return token, nil
	case *jwt.ValidationError:
		validationErr := err.(*jwt.ValidationError)

		switch validationErr.Errors {
		case jwt.ValidationErrorExpired:
			return token, fmt.Errorf("Expired token: %s\n", token.Raw)
		default:
			return token, fmt.Errorf("Token validation error: %s\n", token.Raw)
		}
	default:
		return token, fmt.Errorf("Unable to parse token: %s\n", token.Raw)
	}
}

// GenerateJWTKeys generates a new public/private key pair, to be used to
// create and validate JSON web tokens.
func GenerateJWTKeys(bits int) ([]byte, []byte, error) {
	// http://stackoverflow.com/questions/21151714/go-generate-an-ssh-public-key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)

	publicKey := privateKey.PublicKey
	publicKeyDer, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}
	publicKeyPem := pem.EncodeToMemory(&publicKeyBlock)

	return privateKeyPem, publicKeyPem, nil
}

func createJWTKeyFiles(privKeyPath, pubKeyPath string) {
	_, privErr := os.Stat(privKeyPath)
	_, pubErr := os.Stat(pubKeyPath)
	if os.IsNotExist(privErr) || os.IsNotExist(pubErr) {
		privKey, pubKey, err := GenerateJWTKeys(4096)
		if err != nil {
			log.Fatalf("Unable to create JWT keys: %s\n", err)
		}
		ioutil.WriteFile(privKeyPath, privKey, 0600)
		ioutil.WriteFile(pubKeyPath, pubKey, 0600)
	}
}
