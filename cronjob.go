package cronjob

import "time"

type Result struct {
	Result        map[string]interface{}
	NextRun       *time.Time
	NeverRunAgain bool
}

type CronJob interface {
	// Type gets the type string for the cron job
	Type() string
	// ScheduleFirstRun determines when to run the cron job for the first time
	ScheduleFirstRun() time.Time
	// DefaultRunInterval determines the default amount of time until the next run
	DefaultRunInterval() time.Duration
	// Run executes the cron job, returning some results afterward
	Run() (*Result, error)
}
