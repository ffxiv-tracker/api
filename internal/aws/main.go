package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// NewSession is a wire provider that provides an AWS session
func NewSession() (*session.Session) {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

// NewDynamoClient is a wire provider that provide a dynamodb client
// given an AWS session
func NewDynamoClient(session *session.Session) dynamodbiface.DynamoDBAPI {
	return dynamodb.New(session)
}
