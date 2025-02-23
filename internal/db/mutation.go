package db

import (
	"time"
)

type MutationResult struct {
	ID        int       `json:"id,omitempty" gorm:"primaryKey"`
	Commit    string    `json:"commit"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateMutationResult(result *MutationResult) error {
	return DB.Create(result).Error
}

func GetLatestMutationResult() (*MutationResult, error) {
	var result MutationResult
	err := DB.Order("created_at desc").First(&result).Error
	return &result, err
}
