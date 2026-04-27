package types

import "strconv"

type JobParams struct {
	PRNumber                 string `json:"pr_number"`
	Commit                   string `json:"commit"`
	BaseCommit               string `json:"base_commit"`
	IsMaster                 string `json:"is_master"`
	StepFunctionExecutionARN string `json:"step_function_execution_arn,omitempty"`
	CoverageBatchJobID       string `json:"coverage_batch_job_id,omitempty"`
}

func (j *JobParams) GetPRNumber() int {
	prNumber, _ := strconv.Atoi(j.PRNumber)
	return prNumber
}

func (j *JobParams) GetIsMaster() bool {
	isMaster, _ := strconv.ParseBool(j.IsMaster)
	return isMaster
}

func (j *JobParams) GetCommit() string {
	return j.Commit
}

func (j *JobParams) GetBaseCommit() string {
	return j.BaseCommit
}

func (j *JobParams) GetStepFunctionExecutionARN() string {
	return j.StepFunctionExecutionARN
}

func (j *JobParams) GetCoverageBatchJobID() string {
	return j.CoverageBatchJobID
}

func getJobParams(event map[string]interface{}) (*JobParams, error) {
	params := event["params"].(map[string]interface{})
	commit := params["commit"].(string)
	prNumber := params["pr_number"].(string)
	isMaster := params["is_master"].(string)
	baseCommit := params["base_commit"].(string)

	return &JobParams{
		PRNumber:                 prNumber,
		Commit:                   commit,
		IsMaster:                 isMaster,
		BaseCommit:               baseCommit,
		StepFunctionExecutionARN: getStepFunctionExecutionARN(event),
		CoverageBatchJobID:       getCoverageBatchJobID(event),
	}, nil
}

func GetJobParams(event map[string]interface{}) (*JobParams, error) {
	return getJobParams(event)
}

func getStepFunctionExecutionARN(event map[string]interface{}) string {
	executionARN, _ := event["step_function_execution_arn"].(string)
	return executionARN
}

func getCoverageBatchJobID(event map[string]interface{}) string {
	coverageJob, ok := event["coverage_job"].(map[string]interface{})
	if !ok {
		return ""
	}

	jobID, _ := coverageJob["JobId"].(string)
	return jobID
}
