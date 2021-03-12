package config

// DynamoDBTableName represents the table name
type DynamoDBTableName string

// MasterPK represents the Primary Key for master records
type MasterPK string

// Config holds all config values
type Config struct {
	TableName DynamoDBTableName
	MasterPK MasterPK
}

// DefaultConfig contains default values
var DefaultConfig = Config{
	TableName: "Tasks",
	MasterPK: "MASTER",
}