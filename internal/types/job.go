package types

import (
	"strconv"
)

type JobParams struct {
	PRNumber int    `json:"pr_number"`
	Commit   string `json:"commit"`
	IsMaster bool   `json:"is_master"`
}

func getJobParams(event map[string]interface{}) (*JobParams, error) {
	params := event["params"].(map[string]interface{})
	prNumber, err := strconv.Atoi(params["pr_number"].(string))
	if err != nil {
		return nil, err
	}

	isMaster, err := strconv.ParseBool(params["is_master"].(string))
	if err != nil {
		return nil, err
	}

	commit := params["commit"].(string)

	return &JobParams{
		PRNumber: prNumber,
		Commit:   commit,
		IsMaster: isMaster,
	}, nil
}

func GetJobParams(event map[string]interface{}) (*JobParams, error) {
	return getJobParams(event)
}
