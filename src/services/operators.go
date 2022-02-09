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

package services

import (
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/repositories"
	"context"
)

type Operators interface {
	// GetCurrentOperator gets the currently authenticated operator
	// If there is an authenticated context, but no operator exists for this subject, it will be created
	GetCurrentOperator(ctx context.Context) (domain.Operator, error)
}

type operatorsService struct {
	operators repositories.Operators
}

func NewOperatorsService(operators repositories.Operators) Operators {
	return &operatorsService{
		operators: operators,
	}
}

func (o *operatorsService) GetCurrentOperator(ctx context.Context) (domain.Operator, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil {
		return domain.Operator{}, err
	}

	// TODO store operator in context
	return o.operators.GetOrCreateByToken(ctx, token)
}
