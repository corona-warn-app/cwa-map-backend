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

package model

import "com.t-systems-mms.cwa/domain"

type OperatorDTO struct {
	UUID           string  `json:"uuid"`
	OperatorNumber *string `json:"operatorNumber"`
	Name           string  `json:"name" validate:"required"`
	Email          *string `json:"email" validate:"email"`
	Logo           *string `json:"logo"`
	MarkerIcon     *string `json:"markerIcon"`
	ReportReceiver *string `json:"reportReceiver" validate:"oneof=operator center"`
}

func MapToOperatorDTO(operator *domain.Operator) *OperatorDTO {
	if operator == nil {
		return nil
	}

	var logo *string
	if operator.Logo != nil {
		tmpIcon := "/api/operators/" + operator.UUID + "/logo"
		logo = &tmpIcon
	}

	var markerIcon *string
	if operator.MarkerIcon != nil {
		tmpIcon := "/api/operators/" + operator.UUID + "/marker"
		markerIcon = &tmpIcon
	}

	return &OperatorDTO{
		UUID:           operator.UUID,
		OperatorNumber: operator.OperatorNumber,
		Name:           operator.Name,
		Logo:           logo,
		MarkerIcon:     markerIcon,
		Email:          operator.Email,
		ReportReceiver: operator.BugReportsReceiver,
	}
}
