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

func GetAverageBenchmarkResults(results []*BenchmarkResult) *BenchmarkResult {
	if len(results) == 0 {
		return nil
	}

	avg := &BenchmarkResult{
		Name:                                   results[0].Name,
		Title:                                  results[0].Title,
		Unit:                                   results[0].Unit,
		Batch:                                  0,
		ComplexityN:                            0,
		Epochs:                                 0,
		ClockResolution:                        0,
		ClockResolutionMultiple:                0,
		MaxEpochTime:                           0,
		MinEpochTime:                           0,
		MinEpochIterations:                     0,
		EpochIterations:                        0,
		Warmup:                                 0,
		Relative:                               0,
		MedianElapsed:                          0,
		MedianAbsolutePercentErrorElapsed:      0,
		MedianInstructions:                     0,
		MedianAbsolutePercentErrorInstructions: 0,
		MedianCpucycles:                        0,
		MedianContextswitches:                  0,
		MedianPagefaults:                       0,
		MedianBranchinstructions:               0,
		MedianBranchmisses:                     0,
		TotalTime:                              0,
	}

	for _, result := range results {
		avg.Batch += result.Batch
		avg.ComplexityN += result.ComplexityN
		avg.Epochs += result.Epochs
		avg.ClockResolution += result.ClockResolution
		avg.ClockResolutionMultiple += result.ClockResolutionMultiple
		avg.MaxEpochTime += result.MaxEpochTime
		avg.MinEpochTime += result.MinEpochTime
		avg.MinEpochIterations += result.MinEpochIterations
		avg.EpochIterations += result.EpochIterations
		avg.Warmup += result.Warmup
		avg.Relative += result.Relative
		avg.MedianElapsed += result.MedianElapsed
		avg.MedianAbsolutePercentErrorElapsed += result.MedianAbsolutePercentErrorElapsed
		avg.MedianInstructions += result.MedianInstructions
		avg.MedianAbsolutePercentErrorInstructions += result.MedianAbsolutePercentErrorInstructions
		avg.MedianCpucycles += result.MedianCpucycles
		avg.MedianContextswitches += result.MedianContextswitches
		avg.MedianPagefaults += result.MedianPagefaults
		avg.MedianBranchinstructions += result.MedianBranchinstructions
		avg.MedianBranchmisses += result.MedianBranchmisses
		avg.TotalTime += result.TotalTime
	}

	avg.Batch /= float64(len(results))
	avg.ComplexityN /= float64(len(results))
	avg.Epochs /= float64(len(results))
	avg.ClockResolution /= float64(len(results))
	avg.ClockResolutionMultiple /= float64(len(results))
	avg.MaxEpochTime /= float64(len(results))
	avg.MinEpochTime /= float64(len(results))
	avg.MinEpochIterations /= float64(len(results))
	avg.EpochIterations /= float64(len(results))
	avg.Warmup /= float64(len(results))
	avg.Relative /= float64(len(results))
	avg.MedianElapsed /= float64(len(results))
	avg.MedianAbsolutePercentErrorElapsed /= float64(len(results))
	avg.MedianInstructions /= float64(len(results))
	avg.MedianAbsolutePercentErrorInstructions /= float64(len(results))
	avg.MedianCpucycles /= float64(len(results))
	avg.MedianContextswitches /= float64(len(results))
	avg.MedianPagefaults /= float64(len(results))
	avg.MedianBranchinstructions /= float64(len(results))
	avg.MedianBranchmisses /= float64(len(results))
	avg.TotalTime /= float64(len(results))

	return avg
}
