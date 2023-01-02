package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	tableName      string
	awsRegion      string
	endpoint       string
	dynamodbCLient *dynamodb.DynamoDB
)

func init() {
	tableName = os.Getenv("ORDER_TABLE")
	log.Printf("TABLE NAME : %v", tableName)
	awsRegion = os.Getenv("AWS_REGION")
	log.Printf("AWS REGION : %v", awsRegion)
	awsLocal := os.Getenv("AWS_LOCAL")
	log.Printf("AWS LOCAL : %v", awsLocal)

	if awsLocal == "true" {
		endpoint = "http://docker.for.mac.localhost:8000"
	}
	awsSession, err := session.NewSession(&aws.Config{
		Region:   aws.String(awsRegion),
		Endpoint: aws.String(endpoint),
	})

	if err != nil {
		panic("Unable to initialize the configuration")
	}
	// Create DynamoDB client
	dynamodbCLient = dynamodb.New(awsSession)
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	orderId := request.PathParameters["id"]
	userId := request.PathParameters["user_id"]
	log.Printf("Order id: %v and user id: %v to delete", orderId, userId)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(orderId),
			},
			"user_id": {
				S: aws.String(userId),
			},
		},
		TableName: aws.String(tableName),
	}

	log.Println("Trying to delete order item...")
	_, err := dynamodbCLient.DeleteItem(input)
	if err != nil {
		return handleError(err)
	}

	log.Printf("Order id %v deleted!", orderId)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func handleError(err error) (events.APIGatewayProxyResponse, error) {
	log.Panicf("Unable to process request %v", err)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
