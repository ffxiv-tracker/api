package models

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type TaskResponse struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	Frequency string `json:"frequency"`
}

func NewTaskResponse(val map[string]*dynamodb.AttributeValue) TaskResponse {
	values := strings.Split(*val["SK"].S, "#")

	return TaskResponse{
		Name:      values[2],
		Category:  values[1],
		Frequency: values[0],
	}
}
