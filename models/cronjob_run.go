package models

import (
	"database/sql"
	"time"
)

// CronJobRun is a single run of a cron job
type CronJobRun struct {
	ID          uint64 `gorm:"primaryKey"`
	CronJobID   uint64
	CronJob     *CronJob
	StartDate   time.Time
	EndDate     sql.NullTime
	Success     sql.NullBool
	Result      sql.NullString
	Error       sql.NullString
	NextRunDate sql.NullTime
}

// TableName gets the name of the table
func (CronJobRun) TableName() string {
	return "cronjob-runs"
}
