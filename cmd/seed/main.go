package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type masterList struct {
	Weekly map[string][]string
	Daily  map[string][]string
}

func main() {
	all, err := readJSONFile()
	if err != nil {
		log.Fatalf("failed to get seed data: %s", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	for category, tasks := range all.Weekly {
		for _, task := range tasks {
			err := insert("Weekly", category, task, svc)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	for category, tasks := range all.Daily {
		for _, task := range tasks {
			err := insert("Daily", category, task, svc)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	fmt.Println("Finished importing seed data")
}

func readJSONFile() (*masterList, error) {
	jsonFile, err := os.Open("seed.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	fmt.Println("Successfully opened seed.json")

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var all masterList

	err = json.Unmarshal(byteValue, &all)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal seed data: %w", err)
	}

	return &all, nil
}

func insert(freq string, category string, task string, svc dynamodbiface.DynamoDBAPI) error {
	_, err := svc.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
		TableName:           aws.String("Tasks"),
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String("MASTER")},
			"SK": {S: aws.String(fmt.Sprintf("%s#%s#%s", freq, category, task))},
		},
	})
	if err != nil {
		var e awserr.Error

		if errors.As(err, &e) {
			if e.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return nil
			}
		}

		return fmt.Errorf("failed to write item %s#%s#%s: %w", freq, category, task, err)
	}

	return nil
}
