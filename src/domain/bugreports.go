package domain

import "time"

const (
	ReportReceiverCenter   = "center"
	ReportReceiverOperator = "operator"
)

type BugReport struct {
	UUID          string `gorm:"primaryKey"`
	Created       time.Time
	Email         string
	OperatorUUID  string
	Operator      Operator
	CenterUUID    string
	CenterName    string
	CenterAddress string
	Subject       string
	Message       *string
	Leader        *string
}
