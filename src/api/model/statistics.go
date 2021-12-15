package model

import "com.t-systems-mms.cwa/repositories"

type ReportStatisticsDTO struct {
	Subject     string `json:"subject"`
	ReportCount uint   `json:"report_count"`
}

func (dto ReportStatisticsDTO) FromModel(model repositories.ReportStatistics) *ReportStatisticsDTO {
	return &ReportStatisticsDTO{
		Subject:     model.Subject,
		ReportCount: model.Count,
	}
}
