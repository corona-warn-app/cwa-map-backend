package domain

type Operator struct {
	UUID               string `gorm:"primaryKey"`
	Subject            *string
	OperatorNumber     *string
	Name               string
	Logo               *string
	MarkerIcon         *string
	Email              *string
	BugReportsReceiver *string
}
