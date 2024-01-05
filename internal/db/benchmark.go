package db

type BenchmarkResult struct {
	ID               int `json:"id,omitempty" gorm:"primaryKey"`
	CoverageReportID int `json:"coverage_report_id"`

	Name                                   string  `json:"name"`
	Title                                  string  `json:"title"`
	Unit                                   string  `json:"unit"`
	Batch                                  float64 `json:"batch"`
	ComplexityN                            float64 `json:"complexityN"`
	Epochs                                 float64 `json:"epochs"`
	ClockResolution                        float64 `json:"clockResolution"`
	ClockResolutionMultiple                float64 `json:"clockResolutionMultiple"`
	MaxEpochTime                           float64 `json:"maxEpochTime"`
	MinEpochTime                           float64 `json:"minEpochTime"`
	MinEpochIterations                     float64 `json:"minEpochIterations"`
	EpochIterations                        float64 `json:"epochIterations"`
	Warmup                                 float64 `json:"warmup"`
	Relative                               float64 `json:"relative"`
	MedianElapsed                          float64 `json:"median(elapsed)"`
	MedianAbsolutePercentErrorElapsed      float64 `json:"medianAbsolutePercentError(elapsed)"`
	MedianInstructions                     float64 `json:"median(instructions)"`
	MedianAbsolutePercentErrorInstructions float64 `json:"medianAbsolutePercentError(instructions)"`
	MedianCpucycles                        float64 `json:"median(cpucycles)"`
	MedianContextswitches                  float64 `json:"median(contextswitches)"`
	MedianPagefaults                       float64 `json:"median(pagefaults)"`
	MedianBranchinstructions               float64 `json:"median(branchinstructions)"`
	MedianBranchmisses                     float64 `json:"median(branchmisses)"`
	TotalTime                              float64 `json:"totalTime"`
}

const (
	BENCHMARK_STATUS_PENDING = "pending"
	BENCHMARK_STATUS_SUCCESS = "success"
	BENCHMARK_STATUS_FAILURE = "failure"
)

func CreateBenchmarkResults(reportID int, results []*BenchmarkResult) error {
	err := DB.Where("coverage_report_id = ?", reportID).Delete(&BenchmarkResult{}).Error
	if err != nil {
		return err
	}

	for i := range results {
		results[i].CoverageReportID = reportID
	}

	err = DB.CreateInBatches(&results, 500).Error
	if err != nil {
		return err
	}

	return nil
}
