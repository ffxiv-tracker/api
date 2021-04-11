package manager

import (
	"fmt"

	"ffxiv.anid.dev/internal/config"
	"ffxiv.anid.dev/internal/dao"
	"ffxiv.anid.dev/internal/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

// SaveUserMasterTasks a
func (tm *TasksManager) SaveUserMasterTasks(userID string, tasks *models.UserMasterTaskRequest) (*models.UserMasterTaskResponse, error) {
	taskKey := dao.UserMasterTaskListKey{
		UserID: userID,
		Type: string(fmt.Sprintf("%s#%s#%s", tm.MasterPK, tasks.Frequency, tasks.Category)),
	}

	key, err := dynamodbattribute.MarshalMap(taskKey)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling task key: %w", err)
	}

	update, err := dynamodbattribute.MarshalMap(dao.UserMasterTaskListUpdate{
		Tasks: tasks.Tasks,
	})
	if err != nil {
		return nil, fmt.Errorf("error while marshaling update: %w", err)
	}

	result, err := tm.DynSvc.UpdateItem(&dynamodb.UpdateItemInput{
		Key: key,
		TableName: aws.String(string(tm.Table)),
		UpdateExpression: aws.String("SET tasks = :t"),
		ExpressionAttributeValues: update,
		ReturnValues: aws.String("UPDATED_NEW"),
	})
	if err != nil {
		return nil, fmt.Errorf("error while updating user tasks: %w", err)
	}

	updated := dao.UserMasterTaskListUpdated{}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &updated)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling update: %w", err)
	}

	response := models.UserMasterTaskResponse{
		Frequency: tasks.Frequency,
		Tasks: updated.Tasks,
	}

	if tasks.Category != "" {
		response.Category = &tasks.Category
	}

	return &response, nil
}

func (tm *TasksManager) ValidateUserMasterTaskRequest(tasks *models.UserMasterTaskRequest) (error, error) {
	mapOfAttrKeys := []map[string]*dynamodb.AttributeValue{}

	for _, task := range tasks.Tasks {
		mapOfAttrKeys = append(mapOfAttrKeys, map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(string(tm.MasterPK)),
			},
			"SK": {
				S: aws.String(fmt.Sprintf("%s#%s#%s", tasks.Frequency, tasks.Category, task)),
			},
		})
	}
	
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			string(tm.Table): {
				Keys: mapOfAttrKeys,
			},
		},
	}
	
	batch, err := tm.DynSvc.BatchGetItem(input)
	if err != nil {
		return nil, fmt.Errorf("batch load of tasks failed, err: %w", err)
	}

	for _, items := range batch.Responses {
		if len(items) != len(tasks.Tasks) {
			return fmt.Errorf("an invalid master task was submitted"), nil
		}
	}

	return nil, nil
}

func (tm *TasksManager) GetUserMasterTasks(userID string) ([]*models.UserMasterTaskResponse, error) {
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(string(tm.Table)),
		KeyConditions: map[string]*dynamodb.Condition{
			"PK": {
				ComparisonOperator: aws.String(dynamodb.ComparisonOperatorEq),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userID),
					},
				},
			},
			"SK": {
				ComparisonOperator: aws.String(dynamodb.ComparisonOperatorBeginsWith),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(string(tm.MasterPK)+"#"),
					},
				},
			},
		},
	}
	
	var resp, err = tm.DynSvc.Query(queryInput)
	if err != nil {
		return nil, err
	}

	return models.NewUserMasterTaskResponses(resp.Items), nil
}