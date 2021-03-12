package manager

import (
	"fmt"

	"ffxiv.anid.dev/internal/config"
	"ffxiv.anid.dev/internal/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// TasksManager a
type TasksManager struct {
	DynSvc dynamodbiface.DynamoDBAPI
	Table config.DynamoDBTableName
	MasterPK config.MasterPK
}

// GetMasterTasks a
func (tm *TasksManager) GetMasterTasks() (*[]models.TaskResponse, error) {
	result, err := tm.DynSvc.Query(&dynamodb.QueryInput{
		TableName: aws.String(string(tm.Table)),
		KeyConditionExpression: aws.String("PK = :p"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {S:aws.String(string(tm.MasterPK))},
		},

	})
	if err != nil {
		return nil, fmt.Errorf("error while querying for master tasks: %w", err)
	}

	if result.Count == aws.Int64(0) {
		return nil, fmt.Errorf("no master tasks found")
	}

	tasks := make([]models.TaskResponse, 0)

	for _, record := range result.Items {
		tasks = append(tasks, models.NewTaskResponse(record))
	}

	return &tasks, nil
}