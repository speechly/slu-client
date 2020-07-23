package speechly

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtParser = new(jwt.Parser)

// ErrTokenExpired is returned when parsing an expired AccessToken.
var ErrTokenExpired = errors.New("JWT token has expired")

// AccessToken is a JWT token used for accessing Speechly public APIs.
// It can be obtained using Speechly Identity service.
type AccessToken struct {
	jwt.StandardClaims
	rawToken string
}

// Parse parses the string representation of token into AccessToken.
// If the token is invalid or expired, an error is returned.
func (a *AccessToken) Parse(s string) error {
	_, _, err := jwtParser.ParseUnverified(s, a)
	if err != nil {
		return err
	}

	if !a.VerifyExpiresAt(time.Now().Unix(), true) {
		return ErrTokenExpired
	}

	a.rawToken = s

	return nil
}

func (a AccessToken) String() string {
	return a.rawToken
}
