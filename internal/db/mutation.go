package db

import (
	"time"
)

type MutationState string

const (
	StatusStarted   MutationState = "started"
	StatusCompleted MutationState = "completed"
)

type MutationResult struct {
	ID        int           `json:"id,omitempty" gorm:"primaryKey"`
	Commit    string        `json:"commit"`
	State     MutationState `json:"state"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func CreateMutationResult(result *MutationResult) error {
	return DB.Create(result).Error
}

func GetLatestMutationResult() (*MutationResult, error) {
	var result MutationResult
	err := DB.Order("created_at desc").First(&result).Error
	return &result, err
}

func GetLatestCompletedMutationResult() (*MutationResult, error) {
	var result MutationResult
	err := DB.Where("state = ?", StatusCompleted).Order("created_at desc").First(&result).Error
	return &result, err
}
