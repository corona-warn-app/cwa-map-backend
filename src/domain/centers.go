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

package domain

import (
	"github.com/lib/pq"
	"strings"
	"time"
)

type AppointmentType string

const (
	AppointmentRequired    AppointmentType = "Required"
	AppointmentNotRequired AppointmentType = "NotRequired"
	AppointmentPossible    AppointmentType = "Possible"
)

type TestKind string
type TestKinds []TestKind

const (
	TestKindAntigen     TestKind = "Antigen"
	TestKindPCR         TestKind = "PCR"
	TestKindVaccination TestKind = "Vaccination"
	TestKindAntibody    TestKind = "Antibody"
)

func (t TestKinds) Strings() []string {
	str := make([]string, len(t))
	for i, kind := range t {
		str[i] = string(kind)
	}
	return str
}

// ParseTestKind parses the test kind from the given string and reports if the string contains a valid test kind.
func ParseTestKind(value string) (TestKind, bool) {
	switch strings.ToLower(value) {
	case strings.ToLower(string(TestKindAntigen)):
		return TestKindAntigen, true
	case strings.ToLower(string(TestKindPCR)):
		return TestKindPCR, true
	case strings.ToLower(string(TestKindVaccination)):
		return TestKindVaccination, true
	case strings.ToLower(string(TestKindAntibody)):
		return TestKindAntibody, true
	}
	return "", false
}

// ParseAppointmentType parses the appointment type from the given string and
// reports if the string contains a valid appointment type.
func ParseAppointmentType(value string) (AppointmentType, bool) {
	switch strings.ToLower(value) {
	case "possible":
		return AppointmentPossible, true
	case "notrequired":
		return AppointmentNotRequired, true
	case "required":
		return AppointmentRequired, true
	}
	return "", false
}

type Center struct {
	UUID          string `gorm:"primaryKey"`
	UserReference *string
	Name          string `validate:"required,max=128"`
	LabId         *string
	OperatorName  *string
	Website       *string `validate:"omitempty,url,max=264"`
	Coordinates
	Operator     *Operator `gorm:"foreignKey:OperatorUUID"`
	OperatorUUID string
	Address      string `validate:"required,max=264"`
	AddressNote  *string
	OpeningHours pq.StringArray   `gorm:"type:varchar(64)[]" validate:"dive,max=64"`
	Appointment  *AppointmentType `validate:"omitempty,oneof=Required NotRequired Possible"`
	TestKinds    pq.StringArray   `gorm:"type:varchar(32)[]"`
	EnterDate    *time.Time
	LeaveDate    *time.Time
	DCC          *bool
	Message      *string
	Ranking      float64
	Zip          *string
	Region       *string
	Email        *string `validate:"omitempty,email"`
	Visible      *bool
	LastUpdate   *time.Time
}

type CenterWithDistance struct {
	Center
	Distance float64
}
