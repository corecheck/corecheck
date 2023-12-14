package types

import "strconv"

type JobParams struct {
	PRNumber string `json:"pr_number"`
	Commit   string `json:"commit"`
	IsMaster string `json:"is_master"`
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

func getJobParams(event map[string]interface{}) (*JobParams, error) {
	params := event["params"].(map[string]interface{})
	commit := params["commit"].(string)
	prNumber := params["pr_number"].(string)
	isMaster := params["is_master"].(string)

	return &JobParams{
		PRNumber: prNumber,
		Commit:   commit,
		IsMaster: isMaster,
	}, nil
}

func GetJobParams(event map[string]interface{}) (*JobParams, error) {
	return getJobParams(event)
}
