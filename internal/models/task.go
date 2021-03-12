package models

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

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
