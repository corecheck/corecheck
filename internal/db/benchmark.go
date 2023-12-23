package db

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

const (
	BENCHMARK_STATUS_PENDING = "pending"
	BENCHMARK_STATUS_SUCCESS = "success"
	BENCHMARK_STATUS_FAILURE = "failure"
)

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
