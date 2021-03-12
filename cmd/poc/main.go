package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
    PK   string
    SK  string
    Plot   string
    Rating float64
}

func main() {

	getItem()
	
	getAll()
}

func getItem() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	// Create DynamoDB client
	svc := dynamodb.New(sess)

	tableName := "Tasks"
	movieName := "MASTER"

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(movieName),
			},
			"SK": {
				S: aws.String("*#Duty Roulettes#PVP Frontline"),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if result.Item == nil {
		fmt.Println("Could not find")
		return
	}
		
	item := Item{}
	
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	
	fmt.Println("Found item:")
	fmt.Println("Year:  ", item.PK)
	fmt.Println("Title: ", item.SK)
	fmt.Println("Plot:  ", item.Plot)
	fmt.Println("Rating:", item.Rating)
}

func getAll() {


	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	// Create DynamoDB client
	svc := dynamodb.New(sess)

	tableName := "Tasks"

	result, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :p"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {S:aws.String("MASTER")},
		},

	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if result.Count == aws.Int64(0) {
		fmt.Println("Could not find")
		return
	}
		
	items := make([]Item, 0)
	
	for _, r := range result.Items {
		item := Item{}
		err = dynamodbattribute.UnmarshalMap(r, &item)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}

		items = append(items, item)
	}
	
	for _, item := range items {
		fmt.Println("Found item:")
		fmt.Println("Year:  ", item.PK)
		fmt.Println("Title: ", item.SK)
		fmt.Println("Plot:  ", item.Plot)
		fmt.Println("Rating:", item.Rating)
	}
}