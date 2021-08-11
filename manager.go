package cronjob

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spiretechnology/go-cronjob/models"
	"gorm.io/gorm"
)

type Manager struct {
	db   *gorm.DB
	jobs []CronJob
}

// NewManager creates a new cron job manager instance
func NewManager(db *gorm.DB) (*Manager, error) {

	// Auto-migrate the models
	if err := db.AutoMigrate(
		&models.CronJob{},
		&models.CronJobRun{},
	); err != nil {
		return nil, err
	}

	// Return the manager
	return &Manager{db: db}, nil

}

// Register registers one or more cron jobs
func (m *Manager) Register(cronJob ...CronJob) {

	// Add the jobs to the slice
	m.jobs = append(m.jobs, cronJob...)

}

func (m *Manager) Run(stopChan <-chan bool) {

	// Wait for all the jobs to fully wrap out
	var wg sync.WaitGroup
	wg.Add(len(m.jobs))

	// Loop through the jobs and spawn goroutines for each
	for _, j := range m.jobs {
		go func(job CronJob) {

			// Defer the cleanup function
			defer wg.Done()

			// Manage the job
			if err := m.runCronJob(stopChan, job); err != nil {
				fmt.Println("CronJob manager error for job: ", job.Type(), err.Error())
			}

		}(j)
	}

	// Wait for all jobs to end
	wg.Wait()

}

func (m *Manager) runCronJob(
	stopChan <-chan bool,
	job CronJob,
) error {

	// Find the job model for the cron job type
	dbJob, err := m.getJobModel(job.Type())
	if err != nil {
		return err
	}

	// If there is not yet an instance in the database, create it
	if dbJob == nil {
		dbJob = &models.CronJob{}
		dbJob.Type = job.Type()
		dbJob.CreatedDate = time.Now()
		dbJob.NextRunDate = sql.NullTime{
			Time:  job.ScheduleFirstRun(),
			Valid: true,
		}
		if err := m.db.Save(dbJob).Error; err != nil {
			return err
		}
	}

	// Loop indefinitely
loop:
	for {

		// If there is no next run date
		if !dbJob.NextRunDate.Valid {
			break
		}

		// Wait until the next run time
		select {

		// If the stop channel is closed, we stop waiting for this job and
		// return from the function immediately
		case <-stopChan:
			break loop

		// Wait for the next execution of the job
		case <-time.After(time.Until(dbJob.NextRunDate.Time)):
			if err := m.executeCronJobOnce(job, dbJob); err != nil {
				return err
			}

		}

	}

	// Return without error
	return nil

}

func (m *Manager) executeCronJobOnce(
	job CronJob,
	dbJob *models.CronJob,
) error {

	// Create a job run for the job
	run := models.CronJobRun{
		CronJobID: dbJob.ID,
		StartDate: time.Now(),
	}
	if err := m.db.Save(&run).Error; err != nil {
		return err
	}

	// Run the job. Note that the error returned here is treated differently
	// from other errors. This error will simply be recorded as the result
	// of the cron job, and won't stop the manager flow.
	result, err := job.Run()

	// Update the end date of the job
	run.EndDate = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	run.Success = sql.NullBool{
		Bool:  err == nil,
		Valid: true,
	}

	// If there is an error
	if err != nil {
		run.Error = sql.NullString{
			String: err.Error(),
			Valid:  true,
		}
	}

	// If there is a result map
	if result != nil && len(result.Result) > 0 {
		resultJson, err := json.Marshal(result.Result)
		if err != nil {
			fmt.Println("Error marshalling cron job result: ", err.Error())
		} else {
			run.Result = sql.NullString{
				String: string(resultJson),
				Valid:  true,
			}
		}
	}

	// If there is a next run time
	if result == nil || !result.NeverRunAgain {

		// If a next run time was provided
		if result != nil && result.NextRun != nil {
			run.NextRunDate = sql.NullTime{
				Time:  *result.NextRun,
				Valid: true,
			}
		} else {
			run.NextRunDate = sql.NullTime{
				Time:  time.Now().Add(job.DefaultRunInterval()),
				Valid: true,
			}
		}

	}

	// Save the run in the database
	if err := m.db.Save(&run).Error; err != nil {
		return err
	}

	// Update the job in the database
	dbJob.NextRunDate = run.NextRunDate
	if err := m.db.Save(dbJob).Error; err != nil {
		return err
	}

	// Return without error
	return nil

}

func (m *Manager) getJobModel(jobType string) (*models.CronJob, error) {
	var job models.CronJob
	err := m.db.
		Where("deleted_date IS NULL").
		Where("type = ?", jobType).
		First(&job).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}
