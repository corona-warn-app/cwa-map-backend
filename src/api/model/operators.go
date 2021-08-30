package model

import "com.t-systems-mms.cwa/domain"

type OperatorDTO struct {
	UUID           string  `json:"uuid"`
	OperatorNumber *string `json:"operatorNumber"`
	Name           string  `json:"name"`
	Logo           *string `json:"logo"`
	MarkerIcon     *string `json:"markerIcon"`
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
	}
}
