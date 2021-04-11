package models

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type UserMasterTaskRequest struct {
	Category  string `json:"category"`
	Frequency string `json:"frequency"`
	Tasks []string `json:"tasks"`
}

type UserMasterTaskResponse struct {
	Category  *string `json:"category"`
	Frequency string `json:"frequency"`
	Tasks []*string `json:"tasks"`
}

type TaskResponse struct {
	Name      string `json:"name"`
	Category  *string `json:"category"`
	Frequency string `json:"frequency"`
	Orphan    bool `json:"orphan"`
}

func NewTaskResponse(val map[string]*dynamodb.AttributeValue) TaskResponse {
	values := strings.Split(*val["SK"].S, "#")

	task := TaskResponse{
		Name:      values[2],
		Frequency: values[0],
	}

	if values[1] == "" {
		task.Orphan = true
	} else {
		task.Category = &values[1]
	}

	return task
}

func NewUserMasterTaskResponses(val []map[string]*dynamodb.AttributeValue) []*UserMasterTaskResponse {
	var resp []*UserMasterTaskResponse

	for _, category := range val {
		values := strings.Split(*category["SK"].S, "#")

		r := &UserMasterTaskResponse{
			Frequency: values[1],
			Tasks: category["tasks"].SS,
		}

		if values[2] != "" {
			r.Category = &values[2]
		}

		resp = append(resp, r)
	}

	return resp
}