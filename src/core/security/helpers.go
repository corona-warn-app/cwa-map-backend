package security

import (
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
)

func GetTokenFromContext(ctx context.Context) (jwt.Token, error) {
	token, _, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, err
	} else if token == nil {
		return nil, ErrUnauthorized
	}
	return token, nil
}

func HasRole(ctx context.Context, value string) bool {
	token, err := GetTokenFromContext(ctx)
	if err != nil {
		logrus.WithError(err).Error("Error getting token")
		return false
	}

	realmAccess, ok := token.PrivateClaims()["realm_access"].(map[string]interface{})
	if !ok {
		return false
	}

	roles, ok := realmAccess["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == value {
			return true
		}
	}
	return false
}
