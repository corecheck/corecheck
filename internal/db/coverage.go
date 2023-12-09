package db

import (
	"time"

	"gorm.io/gorm/clause"
)

const (
	COVERAGE_REPORT_STATUS_PENDING = "pending"
	COVERAGE_REPORT_STATUS_SUCCESS = "success"
	COVERAGE_REPORT_STATUS_FAILURE = "failure"

	BENCHMARK_STATUS_PENDING = "pending"
	BENCHMARK_STATUS_SUCCESS = "success"
	BENCHMARK_STATUS_FAILURE = "failure"
)

type BenchmarkResult struct {
	ID               int `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID int `json:"coverage_report_id"`

	Name string  `json:"name"`
	Ir   float64 `json:"Ir"`
	I1mr float64 `json:"I1mr"`
	ILmr float64 `json:"ILmr"`
	Dr   float64 `json:"Dr"`
	D1mr float64 `json:"D1mr"`
	DLmr float64 `json:"DLmr"`
	Dw   float64 `json:"Dw"`
	D1mw float64 `json:"D1mw"`
	DLmw float64 `json:"DLmw"`
}

type CoverageFile struct {
	Name        string          `json:"name" gorm:"-"`
	TestedRatio float64         `json:"tested_ratio" gorm:"-"`
	Hunks       []*CoverageHunk `json:"hunks" gorm:"-"`
}

type CoverageHunk struct {
	Lines []CoverageLine `json:"lines" gorm:"-"`
}

type CoverageReport struct {
	ID                int                          `json:"id,omitempty" gorm:"primaryKey"`
	Status            string                       `json:"status" gorm:"default:pending"`
	BenchmarkStatus   string                       `json:"benchmark_status" gorm:"default:pending"`
	IsMaster          bool                         `json:"is_master"`
	PRNumber          int                          `json:"pr_number"`
	Commit            string                       `json:"commit"`
	BaseCommit        string                       `json:"base_commit"`
	BaseReport        *CoverageReport              `json:"base_report" gorm:"-"`
	CoverageRatio     *float64                     `json:"coverage_ratio"`
	CoverageLines     []CoverageLine               `json:"coverage_lines" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	CoverageFiles     []*CoverageFile              `json:"coverage_files" gorm:"-"`
	Benchmarks        []BenchmarkResult            `json:"-" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	BenchmarksGrouped map[string][]BenchmarkResult `json:"benchmarks_grouped" gorm:"-"`
	Jobs              []Job                        `json:"jobs" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	CreatedAt         time.Time                    `json:"created_at"`
}

type CoverageLine struct {
	ID               int    `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID int    `json:"coverage_report_id"`
	Changed          bool   `json:"changed"`
	File             string `json:"file"`
	LineNumber       int    `json:"line_number"`
	Line             string `json:"line"`
	Covered          bool   `json:"covered"`
	Testable         bool   `json:"testable"`
}

func GetPullCoverageReports(prNum int) ([]*CoverageReport, error) {
	var reports []*CoverageReport
	err := DB.Where("pr_number = ? AND (status = ? OR status = ?)", prNum, COVERAGE_REPORT_STATUS_PENDING, COVERAGE_REPORT_STATUS_SUCCESS).Preload(clause.Associations).Order("created_at desc").Find(&reports).Error
	return reports, err
}

func CreateCoverageReport(report *CoverageReport) error {
	return DB.Create(report).Error
}

func GetCoverageReport(id int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("CoverageLines").Preload("Jobs").Preload("Benchmarks").Where("id = ?", id).First(&report).Error
	return &report, err
}

func GetCoverageReportByCommitPr(commit string, prNum int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("CoverageLines").Preload("Jobs").Preload("Benchmarks").Where("commit = ? AND pr_number = ?", commit, prNum).First(&report).Error
	return &report, err
}

func GetCoverageReportByCommitMaster(commit string) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("CoverageLines").Preload("Jobs").Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}

func UpdateCoverageReport(reportID int, status string, benchStatus string, coverage *float64, baseCommit string) error {
	return DB.Model(&CoverageReport{}).Where("id = ?", reportID).Updates(map[string]interface{}{
		"status":           status,
		"benchmark_status": benchStatus,
		"coverage_ratio":   coverage,
		"base_commit":      baseCommit,
	}).Error
}

func HasCoverageReportForCommit(commit string) (bool, error) {
	var count int64
	err := DB.Model(&CoverageReport{}).Where("commit = ?", commit).Count(&count).Error
	return count > 0, err
}

func CreateLinesCoverage(reportID int, lines []*CoverageLine) error {
	for i := 0; i < len(lines); i += 5000 {
		end := i + 5000
		if end > len(lines) {
			end = len(lines)
		}
		err := DB.CreateInBatches(lines[i:end], len(lines)).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateBenchmarkResults(reportID int, results []*BenchmarkResult) error {
	for _, result := range results {
		result.CoverageReportID = reportID

		err := DB.Create(result).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLatestMasterCoverageReport() (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("is_master = ?", true).Order("created_at desc").First(&report).Error
	return &report, err
}

func GetMasterCoverageReport(commit string) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}
