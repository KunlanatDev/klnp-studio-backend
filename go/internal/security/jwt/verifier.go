package jwtverify

import (
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	jwt "github.com/golang-jwt/jwt/v5"
)

type TokenVerifier interface {
	Verify(rawToken string) (*Claims, error)
}

type Verifier struct {
	keys     *keyfunc.JWKS
	issuer   string
	audience string
}

func New(
	jwksURL string,
	issuer string,
	audience string,
) (*Verifier, error) {
	keys, err := keyfunc.Get(
		jwksURL,
		keyfunc.Options{
			RefreshInterval: time.Hour,
		},
	)
	if err != nil {
		return nil, err
	}

	return &Verifier{
		keys:     keys,
		issuer:   issuer,
		audience: audience,
	}, nil
}

func (v *Verifier) Verify(rawToken string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		v.keys.Keyfunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Subject == "" {
		return nil, errors.New("missing subject")
	}

	if claims.Role != "" && claims.Role != "authenticated" {
		return nil, errors.New("unexpected token role")
	}

	return claims, nil
}
