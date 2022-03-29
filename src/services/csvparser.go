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
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	partnerIdIndex     = "Partner ID"
	userReferenceIndex = "NR."
	nameIndex          = "Name der Teststelle"
	operatorNameIndex  = "Name des Betreibers"
	labIdIndex         = "Lab ID"
	streetIndex        = "Straße"
	houseNumberIndex   = "Hausnr."
	postalCodeIndex    = "PLZ"
	cityIndex          = "Ort"
	enterDateIndex     = "Eintrittsdatum"
	leaveDateIndex     = "Austrittsdatum"
	emailIndex         = "E-Mail"
	openingHoursIndex  = "Öffnungszeiten"
	appointmentIndex   = "Terminbuchung"
	testKindsIndex     = "Testmöglichkeiten"
	websiteIndex       = "Webseite"
	dccIndex           = "Ausstellung eines Dicital Covid Zertifikates (DCC)"
	noteIndex          = "Adresshinweis"
	visibleIndex       = "Sichtbar"
	latitudeIndex      = "Breitengrad"
	longitudeIndex     = "Längengrad"
)

const (
	fieldNotFound = -1
	fieldRequired = -2
)

type CsvParser struct {
}

func (c *CsvParser) Parse(reader io.Reader) ([]ImportCenterResult, error) {
	result := make([]ImportCenterResult, 0)

	validate := validator.New()

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1

	columnMappings := map[string]int{
		partnerIdIndex:     fieldNotFound,
		userReferenceIndex: fieldNotFound,
		nameIndex:          fieldRequired, // required
		labIdIndex:         fieldNotFound,
		operatorNameIndex:  fieldNotFound,
		streetIndex:        fieldRequired, // required
		houseNumberIndex:   fieldNotFound,
		postalCodeIndex:    fieldNotFound,
		cityIndex:          fieldNotFound,
		enterDateIndex:     fieldNotFound,
		leaveDateIndex:     fieldNotFound,
		emailIndex:         fieldRequired, // required
		openingHoursIndex:  fieldNotFound,
		appointmentIndex:   fieldNotFound,
		testKindsIndex:     fieldNotFound,
		websiteIndex:       fieldNotFound,
		dccIndex:           fieldNotFound,
		noteIndex:          fieldNotFound,
		visibleIndex:       fieldNotFound,
		latitudeIndex:      fieldNotFound,
		longitudeIndex:     fieldNotFound,
	}

	headerRows := 0
	for {
		entry, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if _, isParseError := err.(*csv.ParseError); isParseError {
			continue
		} else if err != nil {
			return nil, err
		}

		if headerRows < 2 {
			headerRows = headerRows + 1
			for i, v := range entry {
				if _, ok := columnMappings[strings.TrimSpace(v)]; ok {
					columnMappings[strings.TrimSpace(v)] = i
				}
			}

			if headerRows == 2 {
				for k, v := range columnMappings {
					if v == fieldRequired {
						return nil, &csv.ParseError{
							StartLine: 0,
							Line:      0,
							Column:    0,
							Err:       errors.New("column " + k + " not found"),
						}
					}
				}
			}
			continue
		}

		// test whether there is a complete empty line
		// if so, just skip it
		emptyLine := true
		for _, field := range entry {
			if !util.IsNilOrEmpty(&field) {
				emptyLine = false
				break
			}
		}
		if emptyLine {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"entry": entry,
		}).Debug("Importing center")
		center := c.parseCsvRow(entry, columnMappings)

		if err := validate.Struct(center.Center); err != nil {
			if validationErr, ok := err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validationErr {
					center.Errors = append(center.Errors, fmt.Sprintf("'%s' failed for '%s'",
						fieldErr.Field(), fieldErr.Tag()))
				}
			} else {
				return nil, err
			}
		}

		result = append(result, center)
	}

	return result, nil
}

func (c *CsvParser) parseCsvRow(entry []string, columnMappings map[string]int) ImportCenterResult {
	result := ImportCenterResult{}
	var err error

	var userReference *string
	if index, hasColumn := columnMappings[userReferenceIndex]; hasColumn && index > fieldNotFound {
		if ref := strings.TrimSpace(entry[index]); ref != "" {
			userReference = &ref
		}
	}

	address, warnings := c.parseAddress(entry, columnMappings)
	if warnings != nil {
		result.Warnings = append(result.Warnings, warnings...)
	}

	var openingHours []string
	if index, hasColumn := columnMappings[openingHoursIndex]; hasColumn && index > fieldNotFound {
		openingHours = c.parseOpeningHours(strings.TrimSpace(entry[index]))
	}

	var appointment *domain.AppointmentType
	if index, hasColumn := columnMappings[appointmentIndex]; hasColumn && index > fieldNotFound {
		appointment, err = c.parseAppointmentType(entry[index])
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
		}
	}

	var testKinds domain.TestKinds
	if index, hasColumn := columnMappings[testKindsIndex]; hasColumn && index > fieldNotFound {
		testKinds, err = c.parseTestKinds(entry[index])
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
		} else if len(testKinds) == 0 {
			result.Warnings = append(result.Warnings, "no valid testkinds found")
		}
	}

	var website *string
	if index, hasColumn := columnMappings[websiteIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" && strings.ToLower(entry) != "null" {
			website = &entry
		}
	}

	var email *string
	if index, hasColumn := columnMappings[emailIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" && strings.ToLower(entry) != "null" {
			email = &entry
		}
	}

	dcc := false
	if index, hasColumn := columnMappings[dccIndex]; hasColumn && index > fieldNotFound {
		dcc = strings.ToLower(strings.TrimSpace(entry[index])) == "ja"
	}

	var note *string
	if index, hasColumn := columnMappings[noteIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" {
			note = &entry
		}
	}

	var enterDate *time.Time
	if index, hasColumn := columnMappings[enterDateIndex]; hasColumn && index > fieldNotFound {
		if dateEntry := strings.TrimSpace(entry[index]); dateEntry != "" {
			if date, err := time.Parse("_2.1.2006", dateEntry); err == nil {
				enterDate = &date
			} else {
				result.Errors = append(result.Errors, "invalid date: "+dateEntry)
			}
		}
	}

	var leaveDate *time.Time
	if index, hasColumn := columnMappings[leaveDateIndex]; hasColumn && index > fieldNotFound {
		if dateEntry := strings.TrimSpace(entry[index]); dateEntry != "" {
			if date, err := time.Parse("_2.1.2006", dateEntry); err == nil {
				leaveDate = &date
			} else {
				result.Errors = append(result.Errors, "invalid date: "+dateEntry)
			}
		}
	}

	var visible = true
	if index, hasColumn := columnMappings[visibleIndex]; hasColumn && index > fieldNotFound {
		visible = strings.ToLower(strings.TrimSpace(entry[index])) == "ja"
	}

	var operatorName *string
	if index, hasColumn := columnMappings[operatorNameIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" {
			operatorName = &entry
		}
	}

	var labId *string
	if index, hasColumn := columnMappings[labIdIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" {
			labId = &entry
		}
	}

	var latitude, longitude float64
	if index, hasColumn := columnMappings[latitudeIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" {
			latitude, _ = strconv.ParseFloat(entry, 64)
		}
	}

	if index, hasColumn := columnMappings[longitudeIndex]; hasColumn && index > fieldNotFound {
		if entry := strings.TrimSpace(entry[index]); entry != "" {
			longitude, _ = strconv.ParseFloat(entry, 64)
		}
	}

	fixedCoordinates := false
	if latitude != 0.0 && longitude != 0.0 {
		fixedCoordinates = true
	}

	result.Center = domain.Center{
		UserReference: userReference,
		Name:          strings.TrimSpace(entry[columnMappings[nameIndex]]),
		Email:         email,
		Website:       website,
		Coordinates: domain.Coordinates{
			Longitude: longitude,
			Latitude:  latitude,
			Fixed:     fixedCoordinates,
		},
		Address:      address,
		AddressNote:  note,
		OpeningHours: openingHours,
		Appointment:  appointment,
		TestKinds:    *pq.Array(testKinds.Strings()).(*pq.StringArray),
		EnterDate:    enterDate,
		LeaveDate:    leaveDate,
		DCC:          &dcc,
		Visible:      &visible,
		LabId:        labId,
		OperatorName: operatorName,
	}

	return result
}

func (*CsvParser) parseOpeningHours(entry string) []string {
	if entry == "" {
		return nil
	}

	openingHours := strings.Split(entry, "|")
	if len(openingHours) == 1 {
		openingHours = strings.Split(entry, "\n")
	}
	for i, _ := range openingHours {
		openingHours[i] = strings.TrimSpace(openingHours[i])
	}
	return openingHours
}

func (*CsvParser) parseAddress(entry []string, columnMappings map[string]int) (string, []string) {
	address := strings.TrimSpace(entry[columnMappings[streetIndex]])

	houseNumber := ""
	if index, hasColumn := columnMappings[houseNumberIndex]; hasColumn && index > fieldNotFound {
		houseNumber = strings.TrimSpace(entry[index])
	}

	postalCode := ""
	if index, hasColumn := columnMappings[postalCodeIndex]; hasColumn && index > fieldNotFound {
		postalCode = strings.TrimSpace(entry[index])
	}

	if len(postalCode) == 4 {
		postalCode = "0" + postalCode
	}

	city := ""
	if index, hasColumn := columnMappings[cityIndex]; hasColumn && index > fieldNotFound {
		city = strings.TrimSpace(entry[index])
	}
	if houseNumber != "" {
		address = address + " " + houseNumber
	}

	if postalCode != "" || city != "" {
		address = address + ", " + strings.TrimSpace(postalCode+" "+city)
	}

	return address, nil
}

func (*CsvParser) parseAppointmentType(value string) (*domain.AppointmentType, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "möglich":
		tmp := domain.AppointmentPossible
		return &tmp, nil
	case "nicht erforderlich":
		tmp := domain.AppointmentNotRequired
		return &tmp, nil
	case "nicht notwendig":
		tmp := domain.AppointmentNotRequired
		return &tmp, nil
	case "erforderlich":
		tmp := domain.AppointmentRequired
		return &tmp, nil
	}
	return nil, errors.New("invalid appointment type")
}

func (*CsvParser) parseTestKinds(value string) (domain.TestKinds, error) {
	var err error
	kinds := make([]domain.TestKind, 0)
	elements := strings.Split(value, ",")
	for _, element := range elements {
		element = strings.ToLower(strings.TrimSpace(element))
		if strings.Index(element, "antigen") > -1 || strings.Index(element, "schnelltest") > -1 {
			kinds = append(kinds, domain.TestKindAntigen)
		} else if strings.Index(element, "pcr") > -1 {
			kinds = append(kinds, domain.TestKindPCR)
		} else if strings.Index(element, "impfung") > -1 {
			kinds = append(kinds, domain.TestKindVaccination)
		} else {
			err = errors.New("invalid testkind: " + element)
		}
	}

	return kinds, err
}
