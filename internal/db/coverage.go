package db

import (
	"time"

	"gorm.io/gorm/clause"
)

const (
	COVERAGE_REPORT_STATUS_PENDING = "pending"
	COVERAGE_REPORT_STATUS_SUCCESS = "success"
	COVERAGE_REPORT_STATUS_FAILURE = "failure"
)

type CoverageReport struct {
	ID                int                                      `json:"id,omitempty" gorm:"primaryKey"`
	Status            string                                   `json:"status" gorm:"default:pending"`
	BenchmarkStatus   string                                   `json:"benchmark_status" gorm:"default:pending"`
	IsMaster          bool                                     `json:"is_master"`
	PRNumber          int                                      `json:"pr_number"`
	Commit            string                                   `json:"commit"`
	BaseCommit        string                                   `json:"base_commit"`
	BaseReport        *CoverageReport                          `json:"base_report" gorm:"-"`
	Benchmarks        []BenchmarkResult                        `json:"-" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	BenchmarksGrouped map[string]*BenchmarkResult              `json:"benchmarks_grouped" gorm:"-"`
	Hunks             []CoverageFileHunk                       `json:"-" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	Coverage          map[string]map[string][]CoverageFileHunk `json:"coverage" gorm:"-"`
	CreatedAt         time.Time                                `json:"created_at"`
}

type CoverageFileHunkLine struct {
	ID                 int    `json:"id,omitempty" gorm:"primaryKey"`
	CoverageFileHunkID int    `json:"hunk_id"`
	LineNumber         int    `json:"line_number"`
	Content            string `json:"content"`
	Highlight          bool   `json:"highlight"`
	Context            bool   `json:"context"`
	Covered            bool   `json:"covered"`
	Tested             bool   `json:"tested"`
}

type CoverageFileHunk struct {
	ID               int `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID int `json:"coverage_report_id"`

	CoverageType string                 `json:"coverage_type"`
	Filename     string                 `json:"filename"`
	Lines        []CoverageFileHunkLine `json:"lines" gorm:"foreignKey:CoverageFileHunkID;constraint:OnDelete:CASCADE"`
}

func CreateCoverageReport(report *CoverageReport) error {
	return DB.Create(report).Error
}

func GetCoverageReport(id int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload(clause.Associations).Preload("Hunks.Lines").Where("id = ?", id).First(&report).Error
	return &report, err
}

func GetCoverageReportByCommitPr(commit string, prNum int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Hunks").Preload("Benchmarks").Where("commit = ? AND pr_number = ?", commit, prNum).First(&report).Error
	return &report, err
}

func GetOrCreateCoverageReportByCommitPr(commit string, prNum int, baseCommit string) (*CoverageReport, error) {
	report, err := GetCoverageReportByCommitPr(commit, prNum)
	if err != nil {
		if err.Error() == "record not found" {
			report = &CoverageReport{
				PRNumber:   prNum,
				Commit:     commit,
				IsMaster:   false,
				BaseCommit: baseCommit,
			}

			err = CreateCoverageReport(report)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return report, nil
}

func GetCoverageReportByCommitMaster(commit string) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Hunks").Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}

func GetOrCreateCoverageReportByCommitMaster(commit string) (*CoverageReport, error) {
	report, err := GetCoverageReportByCommitMaster(commit)
	if err != nil {
		if err.Error() == "record not found" {
			report = &CoverageReport{
				Commit:     commit,
				BaseCommit: commit,
				IsMaster:   true,
			}

			err = CreateCoverageReport(report)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return report, nil
}

func UpdateCoverageReport(reportID int, status string, benchStatus string, baseCommit string) error {
	return DB.Model(&CoverageReport{}).Where("id = ?", reportID).Updates(map[string]interface{}{
		"status":           status,
		"benchmark_status": benchStatus,
		"base_commit":      baseCommit,
	}).Error
}

func HasCoverageReportForCommit(commit string) (bool, error) {
	var count int64
	err := DB.Model(&CoverageReport{}).Where("commit = ?", commit).Count(&count).Error
	return count > 0, err
}

func CreateCoverageHunks(reportID int, hunks []*CoverageFileHunk) error {
	err := DB.Where("coverage_report_id = ?", reportID).Delete(&CoverageFileHunk{}).Error
	if err != nil {
		return err
	}

	err = DB.CreateInBatches(hunks, 100).Error
	if err != nil {
		return err
	}

	return nil
}

func GetLatestMasterCoverageReport() (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("is_master = ? AND status = ? AND benchmark_status = ?", true, COVERAGE_REPORT_STATUS_SUCCESS, BENCHMARK_STATUS_SUCCESS).Order("created_at desc").First(&report).Error
	return &report, err
}

func GetLatestPullCoverageReport(prNum int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload(clause.Associations).Preload("Hunks.Lines").Where("pr_number = ? AND (status = ? OR status = ?)", prNum, COVERAGE_REPORT_STATUS_SUCCESS, COVERAGE_REPORT_STATUS_PENDING).Order("created_at desc").First(&report).Error
	return &report, err
}

func GetMasterCoverageReport(commit string) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}

func GetPullReports(number int) ([]*CoverageReport, error) {
	var reports []*CoverageReport
	err := DB.Where("pr_number = ?", number).Order("created_at desc").Find(&reports).Error
	return reports, err
}
