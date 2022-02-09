/*
 *   Corona-Warn-App / cwa-map-backend
 *
 *   (C) 2020, T-Systems International GmbH
 *
 *   Deutsche Telekom AG and all other contributors /
 *   copyright owners license this file to you under the Apache
 *   License, Version 2.0 (the "License"); you may not use this
 *   file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing,
 *   software distributed under the License is distributed on an
 *   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *   KIND, either express or implied.  See the License for the
 *   specific language governing permissions and limitations
 *   under the License.
 */

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
