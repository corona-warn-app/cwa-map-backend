package services

import (
	"com.t-systems-mms.cwa/domain"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const csvHeader = ";ANLAGE 5 - LISTE DER BETREIBER VON TESTSTELLEN;;;;;;;;;;;;;;;;;\n;;;;;;;;;;;;;;;;;;\n\"Partner\nID\";NR. ;\"NAME/FIRMA DES BETREIBERS\nVetragspartner\";LADUNGSFÄHIGE ANSCHRIFT ;;;;;ENTRITTSDATUM;AUSTRITTSDATUM;Name Ansprechpartner/in;Tel.;E-Mail;Öffnungszeiten;Terminbuchung: [erforderlich, möglich, nicht notwendig];[Antigen, PCR, Impfung];Link zu der Webseite Schnelltestzentrums ggf. direkt zur Terminbuchung;Ausstellung eines Dicital Covid Zertifikates (DCC);Freitext\n;;;Straße;Hausnr.;PLZ;Ort;Bundesland;;;;;;;;;;;\n"

var parser = CsvParser{}

func TestCsvParser_SimpleParse(t *testing.T) {
	csv := csvHeader + ";00001;Testzentrum Gaggenau;Eckenerstrasse 1;;76571;Gaggenau;;01.05.2021;01.06.2021;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;möglich;Antigentest,PCR,Impfung;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Empty(t, firstResult.Messages)
	assert.False(t, firstResult.Imported)
	assert.Equal(t, "00001", *firstResult.Center.UserReference)
	assert.Equal(t, "01.05.2021", firstResult.Center.EnterDate.Format("02.01.2006"))
	assert.Equal(t, "01.06.2021", firstResult.Center.LeaveDate.Format("02.01.2006"))
	assert.Equal(t, "Testzentrum Gaggenau", firstResult.Center.Name)
	assert.Equal(t, "Eckenerstrasse 1, 76571 Gaggenau", firstResult.Center.Address)
	assert.Equal(t, domain.AppointmentPossible, firstResult.Center.Appointment)
	assert.Equal(t, "https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau", *firstResult.Center.Website)
	assert.True(t, *firstResult.Center.DCC)
	assert.Equal(t, 2, len(firstResult.Center.OpeningHours))
	assert.Equal(t, 3, len(firstResult.Center.TestKinds))
}

func TestCsvParser_Address(t *testing.T) {
	csv := csvHeader + ";00001;Testzentrum Gaggenau;Eckenerstrasse;1;76571;Gaggenau;;01.05.2021;;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;möglich;Schnelltest,PCR,Impfung;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Equal(t, "Eckenerstrasse 1, 76571 Gaggenau", firstResult.Center.Address)
	assert.Equal(t, 1, len(firstResult.Center.TestKinds))
}

func TestCsvParser_SkipParseErrors(t *testing.T) {
	csv := csvHeader + ";;;\n;00001;Testzentrum Gaggenau;Eckenerstrasse;1;76571;Gaggenau;;01.05.2021;;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;möglich;Antigentest,PCR,Impfung;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Equal(t, "Eckenerstrasse 1, 76571 Gaggenau", firstResult.Center.Address)
}

func TestCsvParser_FixPostalCode(t *testing.T) {
	csv := csvHeader + ";00001;Testzentrum Gaggenau;Eckenerstrasse 1;;1234;Gaggenau;;01.05.2021;;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;erforderlich;INVALID;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Equal(t, "Eckenerstrasse 1, 01234 Gaggenau", firstResult.Center.Address)
	assert.Equal(t, domain.AppointmentRequired, firstResult.Center.Appointment)
}

func TestCsvParser_InvalidAppointment(t *testing.T) {
	csv := csvHeader + ";00001;Testzentrum Gaggenau;Eckenerstrasse 1;;76571;Gaggenau;;01.13.2021;01.00.2021;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;INVALID;Antigentest;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Equal(t, 3, len(firstResult.Messages))
	assert.Equal(t, "invalid appointment type", firstResult.Messages[0])
	assert.Equal(t, "invalid enter date", firstResult.Messages[1])
	assert.Equal(t, "invalid leave date", firstResult.Messages[2])
}

func TestCsvParser_InvalidTestKind(t *testing.T) {
	csv := csvHeader + ";00001;Testzentrum Gaggenau;Eckenerstrasse 1;;76571;Gaggenau;;01.05.2021;;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;Mo.-Fr.: 07:00 - 19:00|Sa.: 07:00 - 12:00;nicht erforderlich;INVALID;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}

	firstResult := result[0]
	assert.Equal(t, 1, len(firstResult.Messages))
	assert.Equal(t, "no valid testkinds found", firstResult.Messages[0])
	assert.Equal(t, 0, len(firstResult.Center.TestKinds))
	assert.Equal(t, domain.AppointmentNotRequired, firstResult.Center.Appointment)
}

func TestCsvParser_HeaderOnly(t *testing.T) {
	csv := csvHeader
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 0, len(result)) {
		return
	}
}

func TestCsvParser_TrimmedCSV(t *testing.T) {
	csv := "\"Partner\nID\";NR. ;\"NAME/FIRMA DES BETREIBERS\nVetragspartner\";LADUNGSFÄHIGE ANSCHRIFT ;;;;;ENTRITTSDATUM;AUSTRITTSDATUM;Name Ansprechpartner/in;Tel.;E-Mail;Öffnungszeiten;Terminbuchung: [erforderlich, möglich, nicht notwendig];[Antigen, PCR, Impfung];Link zu der Webseite Schnelltestzentrums ggf. direkt zur Terminbuchung;Ausstellung eines Dicital Covid Zertifikates (DCC);Freitext\n;;;Straße;Hausnr.;PLZ;Ort;Bundesland;;;;;;;;;;;\n;00001;Testzentrum Gaggenau;Eckenerstrasse 1;;76571;Gaggenau;;01.05.2021;;Tatjana Zambo e.K.;+49 (7225) 79873;info@vitalapo.de;;möglich;Antigentest;https://apo-schnelltest.de/vitalapotheke-im-gesundheitszentrum-gaggenau;ja;"
	result, err := parser.Parse(strings.NewReader(csv))
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	if !assert.Equal(t, 1, len(result)) {
		return
	}
}
