package model

import (
	"com.t-systems-mms.cwa/core/api"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/services"
	"time"
)

type PageCenterDTO struct {
	api.PagedResult
	Result []CenterDTO `json:"result"`
}

type FindCentersResult struct {
	Centers []CenterSummaryDTO `json:"centers"`
}

type ImportCenterRequest struct {
	Centers   []EditCenterDTO `json:"centers"`
	DeleteAll bool            `json:"deleteAll"`
}

type ImportCenterResult struct {
	Center   EditCenterDTO `json:"center"`
	Errors   []string      `json:"errors"`
	Warnings []string      `json:"warnings"`
}

type CenterSummaryDTO struct {
	UUID         string          `json:"uuid"`
	Name         string          `json:"name"`
	Website      *string         `json:"website"`
	Coordinates  *CoordinatesDTO `json:"coordinates"`
	Logo         *string         `json:"logo"`
	Marker       *string         `json:"marker"`
	Address      string          `json:"address"`
	OpeningHours []string        `json:"openingHours"`
	AddressNote  *string         `json:"addressNote"`
	Appointment  *string         `json:"appointment"`
	TestKinds    []string        `json:"testKinds"`
	DCC          *bool           `json:"dcc"`
	Message      *string         `json:"message"`
}

type CenterDTO struct {
	CenterSummaryDTO
	UserReference *string `json:"userReference"`
	EnterDate     *string `json:"enterDate"`
	LeaveDate     *string `json:"leaveDate"`
}

func (CenterSummaryDTO) MapFromDomain(center *domain.Center) *CenterSummaryDTO {
	if center == nil {
		return nil
	}

	return &CenterSummaryDTO{
		UUID:         center.UUID,
		Name:         center.Name,
		Website:      center.Website,
		Coordinates:  CoordinatesDTO{}.MapFromModel(&center.Coordinates),
		Logo:         getCenterLogo(center),
		Marker:       getCenterMarker(center),
		Address:      center.Address,
		OpeningHours: center.OpeningHours,
		AddressNote:  center.AddressNote,
		TestKinds:    center.TestKinds,
		Appointment:  (*string)(center.Appointment),
		DCC:          center.DCC,
		Message:      center.Message,
	}

}

func MapToCenterSummaries(centers []domain.Center) []CenterSummaryDTO {
	result := make([]CenterSummaryDTO, len(centers))
	for i, center := range centers {
		result[i] = *CenterSummaryDTO{}.MapFromDomain(&center)
	}
	return result
}

func (CenterDTO) MapFromDomain(center *domain.Center) *CenterDTO {
	if center == nil {
		return nil
	}

	return &CenterDTO{
		CenterSummaryDTO: *CenterSummaryDTO{}.MapFromDomain(center),
		UserReference:    center.UserReference,
		EnterDate:        mapDateToString(center.EnterDate),
		LeaveDate:        mapDateToString(center.LeaveDate),
	}
}

func getCenterLogo(center *domain.Center) *string {
	if center == nil || center.Operator == nil || center.Operator.Logo == nil {
		return nil
	}
	logo := "/api/operators/" + center.OperatorUUID + "/logo"
	return &logo
}

func getCenterMarker(center *domain.Center) *string {
	if center == nil || center.Operator == nil || center.Operator.MarkerIcon == nil {
		return nil
	}
	logo := "/api/operators/" + center.OperatorUUID + "/marker"
	return &logo
}

func MapToCenterDTOs(centers []domain.Center) []CenterDTO {
	result := make([]CenterDTO, len(centers))
	for i, center := range centers {
		result[i] = *CenterDTO{}.MapFromDomain(&center)
	}
	return result
}

type EditCenterDTO struct {
	UserReference *string  `json:"userReference"`
	Name          string   `json:"name" validate:"required"`
	Website       *string  `json:"website"`
	Address       string   `json:"address" validate:"required"`
	OpeningHours  []string `json:"openingHours"`
	AddressNote   *string  `json:"addressNote"`
	Appointment   *string  `json:"appointment"`
	TestKinds     []string `json:"testKinds"`
	DCC           *bool    `json:"dcc"`
	EnterDate     *string  `json:"enterDate"`
	LeaveDate     *string  `json:"leaveDate"`
	Note          *string  `json:"note"`
}

func (c EditCenterDTO) MapToDomain() domain.Center {
	var enterDate *time.Time
	if c.EnterDate != nil {
		if date, err := time.Parse("02.01.2006", *c.EnterDate); err == nil {
			enterDate = &date
		}
	}

	var leaveDate *time.Time
	if c.LeaveDate != nil {
		if date, err := time.Parse("02.01.2006", *c.LeaveDate); err == nil {
			leaveDate = &date
		}
	}

	return domain.Center{
		UserReference: c.UserReference,
		Name:          c.Name,
		Website:       c.Website,
		Address:       c.Address,
		AddressNote:   c.AddressNote,
		OpeningHours:  c.OpeningHours,
		Appointment:   (*domain.AppointmentType)(c.Appointment),
		TestKinds:     c.TestKinds,
		DCC:           c.DCC,
		EnterDate:     enterDate,
		LeaveDate:     leaveDate,
	}
}

func (EditCenterDTO) MapFromDomain(center domain.Center) EditCenterDTO {
	return EditCenterDTO{
		UserReference: center.UserReference,
		Name:          center.Name,
		Website:       center.Website,
		Address:       center.Address,
		OpeningHours:  center.OpeningHours,
		AddressNote:   center.AddressNote,
		Appointment:   (*string)(center.Appointment),
		TestKinds:     center.TestKinds,
		DCC:           center.DCC,
		EnterDate:     mapDateToString(center.EnterDate),
		LeaveDate:     mapDateToString(center.LeaveDate),
	}
}

func MapToImportCenterResultDTO(result services.ImportCenterResult) ImportCenterResult {
	return ImportCenterResult{
		Center:   EditCenterDTO{}.MapFromDomain(result.Center),
		Errors:   result.Errors,
		Warnings: result.Warnings,
	}
}

func MapToImportCenterResultDTOs(centers []services.ImportCenterResult) []ImportCenterResult {
	result := make([]ImportCenterResult, len(centers))
	for i, center := range centers {
		result[i] = MapToImportCenterResultDTO(center)
	}
	return result
}

func mapDateToString(str *time.Time) *string {
	if str == nil {
		return nil
	}
	tmp := str.Format("02.01.2006")
	return &tmp
}
