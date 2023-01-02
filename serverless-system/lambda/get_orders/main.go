package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Create struct to hold info about new item
type OrderRequest struct {
	UserId       string    `json:"user_id"`
	RestaurantId string    `json:"restaurant_id"`
	Quantity     string    `json:"quantity"`
	Id           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	OrderStatus  string    `json:"order_status"`
}

const DEFAULT_ORDER_STATUS = "PENDING"

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

	inputItem := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	log.Println("Trying to scan all table items...")
	resultScan, err := dynamodbCLient.Scan(inputItem)

	if err != nil {
		return handleError(err)
	}

	resultItem := new([]OrderRequest)
	log.Println("Trying to unmarshall scan results...")
	err = dynamodbattribute.UnmarshalListOfMaps(resultScan.Items, resultItem)

	if err != nil {
		return handleError(err)
	}

	result_marshalled, err := json.Marshal(resultItem)

	if err != nil {
		return handleError(err)
	}
	log.Println("Retrieval Items succeeded!")
	return events.APIGatewayProxyResponse{
		Body: string(result_marshalled),
	}, nil
}

func handleError(err error) (events.APIGatewayProxyResponse, error) {
	log.Panicf("Unable to process request , Exception occured : %v", err)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
