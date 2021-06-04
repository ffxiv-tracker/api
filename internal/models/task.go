package models

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type UserMasterTaskRequest struct {
	Category  string   `json:"category"`
	Frequency string   `json:"frequency"`
	Tasks     []string `json:"tasks"`
}

type UserMasterTaskResponse struct {
	Category  *string   `json:"category"`
	Frequency string    `json:"frequency"`
	Tasks     []*string `json:"tasks"`
}

type MasterTaskResponse struct {
	Category  *string   `json:"category"`
	Frequency string    `json:"frequency"`
	Tasks     []*string `json:"tasks"`
}

func NewUserMasterTaskResponses(val []map[string]*dynamodb.AttributeValue) []*UserMasterTaskResponse {
	resp := make([]*UserMasterTaskResponse, 0)

	for _, category := range val {
		values := strings.Split(*category["SK"].S, "#")

		r := &UserMasterTaskResponse{
			Frequency: values[1],
			Tasks:     category["tasks"].SS,
		}

		if values[2] != "" {
			r.Category = &values[2]
		}

		resp = append(resp, r)
	}

	return resp
}

func NewMasterTaskResponses(val []map[string]*dynamodb.AttributeValue) []*MasterTaskResponse {
	var resp []*MasterTaskResponse

	index := make(map[string]*MasterTaskResponse)

	for _, record := range val {
		values := strings.Split(*record["SK"].S, "#")

		freq := values[0]
		cat := values[1]
		name := values[2]

		if rsp, ok := index[freq+cat]; ok {
			rsp.Tasks = append(rsp.Tasks, &name)
		} else {
			rsp = &MasterTaskResponse{
				Frequency: freq,
				Tasks:     []*string{&name},
			}

			if cat != "" {
				rsp.Category = &cat
			}

			resp = append(resp, rsp)

			index[freq+cat] = rsp
		}

	}

	return resp
}
