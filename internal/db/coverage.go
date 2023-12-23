package db

import (
	"time"

	"github.com/corecheck/corecheck/internal/types"
	"gorm.io/gorm/clause"
)

const (
	COVERAGE_REPORT_STATUS_PENDING = "pending"
	COVERAGE_REPORT_STATUS_SUCCESS = "success"
	COVERAGE_REPORT_STATUS_FAILURE = "failure"

	COVERAGE_TYPE_UNCOVERED_NEW_CODE               = "uncovered_new_code"
	COVERAGE_TYPE_LOST_BASELINE_COVERAGE           = "lost_baseline_coverage"
	COVERAGE_TYPE_UNCOVERED_INCLUDED_CODE          = "uncovered_included_code"
	COVERAGE_TYPE_UNCOVERED_BASELINE_CODE          = "uncovered_baseline_code"
	COVERAGE_TYPE_GAINED_BASELINE_COVERAGE         = "gained_baseline_coverage"
	COVERAGE_TYPE_GAINED_COVERAGE_INCLUDED_CODE    = "gained_coverage_included_code"
	COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE         = "gained_coverage_new_code"
	COVERAGE_TYPE_COVERED_BASELINE_CODE            = "covered_baseline_code"
	COVERAGE_TYPE_EXCLUDED_UNCOVERED_BASELINE_CODE = "excluded_uncovered_baseline_code"
	COVERAGE_TYPE_EXCLUDED_COVERED_BASELINE_CODE   = "excluded_covered_baseline_code"
	COVERAGE_TYPE_DELETED_UNCOVERED_BASELINE_CODE  = "deleted_uncovered_baseline_code"
	COVERAGE_TYPE_DELETED_COVERED_BASELINE_CODE    = "deleted_covered_baseline_code"
)

type CoverageReport struct {
	ID                int                           `json:"id,omitempty" gorm:"primaryKey"`
	Status            string                        `json:"status" gorm:"default:pending"`
	BenchmarkStatus   string                        `json:"benchmark_status" gorm:"default:pending"`
	IsMaster          bool                          `json:"is_master"`
	PRNumber          int                           `json:"pr_number"`
	Commit            string                        `json:"commit"`
	BaseCommit        string                        `json:"base_commit"`
	BaseReport        *CoverageReport               `json:"base_report" gorm:"-"`
	CoverageRatio     *float64                      `json:"coverage_ratio"`
	CoverageLines     []CoverageLine                `json:"coverage_lines" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	Benchmarks        []BenchmarkResult             `json:"-" gorm:"foreignKey:CoverageReportID;constraint:OnDelete:CASCADE"`
	BenchmarksGrouped map[string][]BenchmarkResult  `json:"benchmarks_grouped" gorm:"-"`
	Coverage          map[string][]CoverageFileHunk `json:"coverage" gorm:"-"`
	CreatedAt         time.Time                     `json:"created_at"`
}

type CoverageLine struct {
	ID                 int    `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID   int    `json:"coverage_report_id"`
	CoverageType       string `json:"coverage_type"`
	File               string `json:"file"`
	OriginalLineNumber int    `json:"original_line_number"`
	NewLineNumber      int    `json:"new_line_number"`
}
type CoverageFileHunkLine struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	Highlight  bool   `json:"highlight"`
	Context    bool   `json:"context"`
}

type CoverageFileHunk struct {
	Filename string                 `json:"filename"`
	Lines    []CoverageFileHunkLine `json:"lines"`
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
	err := DB.Preload("CoverageLines").Preload("Benchmarks").Where("id = ?", id).First(&report).Error
	return &report, err
}

func GetCoverageReportByCommitPr(commit string, prNum int) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("CoverageLines").Preload("Benchmarks").Where("commit = ? AND pr_number = ?", commit, prNum).First(&report).Error
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
	err := DB.Preload("CoverageLines").Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}

func GetOrCreateCoverageReportByCommitMaster(commit string) (*CoverageReport, error) {
	report, err := GetCoverageReportByCommitMaster(commit)
	if err != nil {
		if err.Error() == "record not found" {
			report = &CoverageReport{
				Commit:   commit,
				IsMaster: true,
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
	err := DB.Where("coverage_report_id = ?", reportID).Delete(&CoverageLine{}).Error
	if err != nil {
		return err
	}

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

func StoreDifferentialCoverage(reportID int, differentialCoverage *types.DifferentialCoverage) error {
	var lines []*CoverageLine
	for _, line := range differentialCoverage.UncoveredNewCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_UNCOVERED_NEW_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.LostBaselineCoverage {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_LOST_BASELINE_COVERAGE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.UncoveredIncludedCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_UNCOVERED_INCLUDED_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.GainedBaselineCoverage {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_GAINED_BASELINE_COVERAGE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.GainedCoverageIncludedCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_GAINED_COVERAGE_INCLUDED_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.GainedCoverageNewCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}

	for _, line := range differentialCoverage.ExcludedUncoveredBaselineCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_EXCLUDED_UNCOVERED_BASELINE_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}
	for _, line := range differentialCoverage.ExcludedCoveredBaselineCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_EXCLUDED_COVERED_BASELINE_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      line.NewLineNumber,
		})
	}
	for _, line := range differentialCoverage.DeletedUncoveredBaselineCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_DELETED_UNCOVERED_BASELINE_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      -1,
		})
	}
	for _, line := range differentialCoverage.DeletedCoveredBaselineCode {
		lines = append(lines, &CoverageLine{
			CoverageReportID:   reportID,
			CoverageType:       COVERAGE_TYPE_DELETED_COVERED_BASELINE_CODE,
			File:               line.File,
			OriginalLineNumber: line.OriginalLineNumber,
			NewLineNumber:      -1,
		})
	}

	return CreateLinesCoverage(reportID, lines)
}

func GetLatestMasterCoverageReport() (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("is_master = ? AND status = ?", true, COVERAGE_REPORT_STATUS_SUCCESS).Order("created_at desc").First(&report).Error
	return &report, err
}

func GetMasterCoverageReport(commit string) (*CoverageReport, error) {
	var report CoverageReport
	err := DB.Preload("Benchmarks").Where("commit = ? AND is_master = ?", commit, true).First(&report).Error
	return &report, err
}
