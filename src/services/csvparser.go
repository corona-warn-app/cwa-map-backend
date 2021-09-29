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
	"strings"
	"time"
)

const (
	userReferenceIndex = 1
	nameIndex          = 2
	streetIndex        = 3
	houseNumberIndex   = 4
	postalCodeIndex    = 5
	cityIndex          = 6
	enterDateIndex     = 8
	leaveDateIndex     = 9
	emailIndex         = 12
	openingHoursIndex  = 13
	appointmentIndex   = 14
	testKindsIndex     = 15
	websiteIndex       = 16
	dccIndex           = 17
	noteIndex          = 18

	expectedFieldCount = 19
)

type CsvParser struct {
}

func (c *CsvParser) Parse(reader io.Reader) ([]ImportCenterResult, error) {
	result := make([]ImportCenterResult, 0)

	validate := validator.New()

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1

	headerFound := false
	fieldOffset := 0
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

		if !headerFound {
			for i, v := range entry {
				// Upload template is a bit strange, so we are looking for the "Partner ID" (with newline) for alignment
				if v == "Partner\nID" {
					headerFound = true
					fieldOffset = i
					// Skip one more line, as we have a two row header
					// error (EOF) will be catched in the next iteration
					_, _ = csvReader.Read()
					break
				} else if v == "Partner ID" {
					headerFound = true
					fieldOffset = i
				}
			}
			continue
		}

		entry = entry[fieldOffset:]
		if len(entry) < expectedFieldCount {
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
		center := c.parseCsvRow(entry)

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

func (c *CsvParser) parseCsvRow(entry []string) ImportCenterResult {
	result := ImportCenterResult{}

	var userReference *string
	if ref := strings.TrimSpace(entry[userReferenceIndex]); ref != "" {
		userReference = &ref
	}

	address, warnings := c.parseAddress(entry)
	if warnings != nil {
		result.Warnings = append(result.Warnings, warnings...)
	}

	openingHours := c.parseOpeningHours(strings.TrimSpace(entry[openingHoursIndex]))

	appointment, err := c.parseAppointmentType(entry[appointmentIndex])
	if err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	}

	testKinds, err := c.parseTestKinds(entry[testKindsIndex])
	if err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	} else if len(testKinds) == 0 {
		result.Warnings = append(result.Warnings, "no valid testkinds found")
	}

	var website *string
	if entry := strings.TrimSpace(entry[websiteIndex]); entry != "" && strings.ToLower(entry) != "null" {
		website = &entry
	}

	var email *string
	if entry := strings.TrimSpace(entry[emailIndex]); entry != "" && strings.ToLower(entry) != "null" {
		email = &entry
	}

	dcc := strings.ToLower(strings.TrimSpace(entry[dccIndex])) == "ja"

	var note *string
	if entry := strings.TrimSpace(entry[noteIndex]); entry != "" {
		note = &entry
	}

	var enterDate *time.Time
	if dateEntry := strings.TrimSpace(entry[enterDateIndex]); dateEntry != "" {
		if date, err := time.Parse("02.01.2006", dateEntry); err == nil {
			enterDate = &date
		} else {
			result.Errors = append(result.Errors, "invalid date: "+dateEntry)
		}
	}

	var leaveDate *time.Time
	if dateEntry := strings.TrimSpace(entry[leaveDateIndex]); dateEntry != "" {
		if date, err := time.Parse("02.01.2006", dateEntry); err == nil {
			leaveDate = &date
		} else {
			result.Errors = append(result.Errors, "invalid date: "+dateEntry)
		}
	}

	result.Center = domain.Center{
		UserReference: userReference,
		Name:          strings.TrimSpace(entry[nameIndex]),
		Email:         email,
		Website:       website,
		Coordinates: domain.Coordinates{
			Longitude: 0,
			Latitude:  0,
		},
		Address:      address,
		AddressNote:  note,
		OpeningHours: openingHours,
		Appointment:  appointment,
		TestKinds:    *pq.Array(testKinds.Strings()).(*pq.StringArray),
		EnterDate:    enterDate,
		LeaveDate:    leaveDate,
		DCC:          &dcc,
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

func (*CsvParser) parseAddress(entry []string) (string, []string) {
	address := strings.TrimSpace(entry[streetIndex])
	houseNumber := strings.TrimSpace(entry[houseNumberIndex])
	postalCode := strings.TrimSpace(entry[postalCodeIndex])
	if len(postalCode) == 4 {
		postalCode = "0" + postalCode
	}

	city := strings.TrimSpace(entry[cityIndex])
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
