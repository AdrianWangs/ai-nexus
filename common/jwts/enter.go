package jwts

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JwtPayload struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
}

type CustomClaims struct {
	JwtPayload
	jwt.RegisteredClaims
}

// GenToken generate jwt token
func GenToken(user JwtPayload, accessSecret string, expires int64) (string, error) {

	claim := CustomClaims{
		user,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expires))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	return token.SignedString([]byte(accessSecret))
}

// ParseToken parse jwt token
func ParseToken(tokenString string, accessSecret string, expires int64) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
