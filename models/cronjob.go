package models

import (
	"database/sql"
	"time"
)

// CronJob is a job that runs on the server side periodically
type CronJob struct {
	ID          uint64 `gorm:"primaryKey"`
	Type        string
	CreatedDate time.Time
	NextRunDate sql.NullTime
	DeletedDate sql.NullTime
}

// TableName gets the name of the table
func (CronJob) TableName() string {
	return "cronjobs"
}
