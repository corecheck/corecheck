package db

import (
	"time"
)

type Job struct {
	ID               int  `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID *int `json:"coverage_report_id"`

	Master            bool   `json:"master"`
	PullRequestNumber int    `json:"pull_request_number"`
	Commit            string `json:"pull_request_commit"`

	Type      string    `json:"type"`
	Payload   *string   `json:"payload"`
	CreatedAt time.Time `json:"created_at"`

	AWSJobID        *string `json:"aws_job_id"`
	AWSJobName      *string `json:"aws_job_name"`
	AWSStatus       *string `json:"aws_status"`
	AWSStatusReason *string `json:"aws_status_reason"`
	AWSStartedAt    *int64  `json:"aws_started_at"`
	AWSCreatedAt    *int64  `json:"aws_created_at"`
	AWSIsTerminated *bool   `json:"aws_is_terminated"`
	AWSIsCancelled  *bool   `json:"aws_is_cancelled"`
}

func CreateJob(job *Job) error {
	return DB.Create(job).Error
}

func GetJob(id int) (*Job, error) {
	var job Job
	err := DB.Where("id = ?", id).First(&job).Error
	return &job, err
}

type ListJobsFilter struct {
	Status            *string
	PullRequestNumber *int
	Type              *string
}

func ListJobs(filter *ListJobsFilter) ([]*Job, error) {
	var jobs []*Job
	query := DB
	if filter.Status != nil {
		query = query.Where("aws_status = ?", *filter.Status)
	}
	if filter.PullRequestNumber != nil {
		query = query.Where("pull_request_number = ?", *filter.PullRequestNumber)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	err := query.Order("created_at desc").Find(&jobs).Error
	return jobs, err
}

func ListPendingJobs() ([]*Job, error) {
	var jobs []*Job
	err := DB.Where("aws_status IS NULL OR (aws_status != 'SUCCEEDED' AND aws_status != 'FAILED')").Order("created_at desc").Find(&jobs).Error
	return jobs, err
}

func UpdateJob(job *Job) error {
	return DB.Save(job).Error
}

func GetRemainingJobs(jobType string, prNum int, commit string) ([]*Job, error) {
	var jobs []*Job
	err := DB.Where("(aws_status IS NULL) OR (aws_status != 'SUCCEEDED' AND aws_status != 'FAILED') AND type = ? AND pull_request_number = ? AND pull_request_commit = ?", jobType, prNum, commit).Order("created_at desc").Find(&jobs).Error
	return jobs, err
}
