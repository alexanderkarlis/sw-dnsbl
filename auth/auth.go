package auth

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var signingSalt = []byte("thisisasecret")

// CustomAuthClaims custom jwt signing
type CustomAuthClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateJWT grants a JWT based on username, password, and alotted exp time(min)
func CreateJWT(username, password string, expTime int) (string, error) {
	claims := CustomAuthClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(expTime) * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		username,
		password,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(signingSalt)

	if err != nil {
		return "", err
	}

	return signed, nil
}

// ValidateToken validates an assigned token
func ValidateToken(tokenString string) (*CustomAuthClaims, error) {
	authClaims := &CustomAuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, authClaims, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return signingSalt, nil
	})
	if err != nil {
		log.Println(err)
		return authClaims, err
	}
	if token.Valid {
		log.Println("token is okay.")
		return authClaims, err
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			log.Println("not a valid token")
			return authClaims, err
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			log.Println("Time expired")
			return authClaims, err
		} else {
			log.Println("Something went wrong:", err)
			return authClaims, err
		}
	} else {
		log.Println("Something went wrong:", err)
		return authClaims, err
	}
}
