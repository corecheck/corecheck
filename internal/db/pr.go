package db

import (
	"time"

	"github.com/google/go-github/v57/github"
)

type PR struct {
	Number    int        `json:"number" gorm:"primaryKey"`
	State     *string    `json:"state"`
	Title     *string    `json:"title"`
	Body      *string    `json:"body"`
	CreatedAt *time.Time `json:"created_at" gorm:"autoCreateTime:false"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime:false"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
	User      string     `json:"user"`
	Head      string     `json:"head"`
	HeadRepo  string     `json:"head_repo"`
	HeadRef   string     `json:"head_ref"`

	Reports []*CoverageReport `json:"reports" gorm:"-"`
}

func GetLatestPRUpdate() (time.Time, error) {
	r, err := DB.Model(&PR{}).Select("updated_at").Order("updated_at desc").Limit(1).Rows()
	if err != nil {
		return time.Time{}, err
	}

	defer r.Close()

	var updatedAt time.Time
	for r.Next() {
		r.Scan(&updatedAt)
	}

	return updatedAt, nil
}

func UpdateOrCreatePR(pr *github.PullRequest) error {
	var existingPR PR
	err := DB.Where("number = ?", pr.GetNumber()).First(&existingPR).Error
	if err != nil {
		if err.Error() == "record not found" {
			return DB.Create(&PR{
				Number:    pr.GetNumber(),
				State:     pr.State,
				Title:     pr.Title,
				Body:      pr.Body,
				CreatedAt: pr.CreatedAt.GetTime(),
				UpdatedAt: pr.UpdatedAt.GetTime(),
				ClosedAt:  pr.ClosedAt.GetTime(),
				MergedAt:  pr.MergedAt.GetTime(),
				User:      pr.User.GetLogin(),
				Head:      pr.Head.GetSHA(),
				HeadRepo:  pr.Head.Repo.GetFullName(),
				HeadRef:   pr.Head.GetRef(),
			}).Error
		}
		return err
	}

	return DB.Model(&existingPR).Updates(&PR{
		Number:    pr.GetNumber(),
		State:     pr.State,
		Title:     pr.Title,
		Body:      pr.Body,
		CreatedAt: pr.CreatedAt.GetTime(),
		UpdatedAt: pr.UpdatedAt.GetTime(),
		ClosedAt:  pr.ClosedAt.GetTime(),
		MergedAt:  pr.MergedAt.GetTime(),
		User:      pr.User.GetLogin(),
		Head:      pr.Head.GetSHA(),
		HeadRepo:  pr.Head.Repo.GetFullName(),
		HeadRef:   pr.Head.GetRef(),
	}).Error
}

func GetPR(number int) (*PR, error) {
	var pr PR
	err := DB.Preload("Reports").Where("number = ?", number).First(&pr).Error
	return &pr, err
}

type SearchPRsOptions struct {
	Title string
	Page  int
}

const pageSize = 100

func ListPulls(opts SearchPRsOptions) ([]PR, error) {
	var prs []PR
	err := DB.Where("title LIKE ? AND state = ?", "%"+opts.Title+"%", "open").Order("updated_at desc").Offset((opts.Page - 1) * pageSize).Limit(pageSize).Find(&prs).Error
	return prs, err
}

func UpdatePR(pr *PR) error {
	return DB.Save(pr).Error
}
