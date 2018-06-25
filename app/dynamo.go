package app

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"fmt"
)

type Dynamo struct {
	Host   string `env:"DYNAMO_ENDPOINT,required"`
	Region string `env:"AWS_REGION,required"`
	Key    string `env:"AWS_ACCESS_KEY_ID,required"`
	Secret string `env:"AWS_SECRET_ACCESS_KEY,required"`
}

func (dynamo Dynamo) CreateSession() (db DbSessionDynamo) {

	config, err := external.LoadDefaultAWSConfig()
	config.Region = dynamo.Region
	config.Credentials = &aws.SafeCredentialsProvider{
		RetrieveFn: func() (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     dynamo.Key,
				SecretAccessKey: dynamo.Secret,
			}, nil
		},
	}

	if err != nil {
		// TODO cuz i dunno yet
		fmt.Println("Dynamo session failed")
	}

	return &DynamoSession{dynamodb.New(config)}
}

type DbSessionDynamo interface {
	ScanRequest(input *dynamodb.ScanInput) dynamodb.ScanRequest
	Insert(input *dynamodb.PutItemInput) dynamodb.PutItemRequest
	ListTablesRequest(input *dynamodb.ListTablesInput) dynamodb.ListTablesRequest
	IsHealthy() bool
}

type DynamoSession struct {
	*dynamodb.DynamoDB
}

func (dynamoSession *DynamoSession) IsHealthy() bool {
	listTablesRequest := dynamoSession.ListTablesRequest(&dynamodb.ListTablesInput{})
	response, err := listTablesRequest.Send()
	if err != nil {
		fmt.Println("Error occurred listing tables from DynamoDB: ", err.Error())
		return false
	}

	if len(response.TableNames) > 0 {
		return true
	}

	fmt.Println("There are no Tables in DynamoDB")
	return false
}

func (dynamoSession *DynamoSession) ScanRequest(input *dynamodb.ScanInput) dynamodb.ScanRequest {
	return dynamoSession.DynamoDB.ScanRequest(input)
}

func (dynamoSession *DynamoSession) Insert(input *dynamodb.PutItemInput) dynamodb.PutItemRequest {
	return dynamoSession.DynamoDB.PutItemRequest(input)
}

func (dynamoSession *DynamoSession) Update(input *dynamodb.UpdateItemInput) dynamodb.UpdateItemRequest {
	return dynamoSession.DynamoDB.UpdateItemRequest(input)
}

func (dynamoSession *DynamoSession) ListTablesRequest(input *dynamodb.ListTablesInput) dynamodb.ListTablesRequest {
	return dynamoSession.DynamoDB.ListTablesRequest(input)
}
