package main

import (
	"encoding/json"
	"log"
	"math/rand"
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
	Id           int       `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	OrderStatus  string    `json:"order_status"`
}

const DEFAULT_ORDER_STATUS = "PENDING"

var (
	tableName      string
	awsRegion      string
	dynamodbCLient *dynamodb.DynamoDB
)

func init() {

	tableName = os.Getenv("ORDER_TABLE")
	log.Printf("TABLE NAME : %v", tableName)
	awsRegion = os.Getenv("AWS_REGION")
	log.Printf("AWS REGION : %v", awsRegion)
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion)},
	)

	if err != nil {
		panic("Unable to initalize the configuration")
	}
	// Create DynamoDB client
	dynamodbCLient = dynamodb.New(awsSession)
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var orderRequest OrderRequest

	log.Printf("Order request from the API GATEWAY : %v", request.Body)
	if err := json.Unmarshal([]byte(request.Body), &orderRequest); err != nil {
		return handleError(err)
	}

	orderRequest.OrderStatus = DEFAULT_ORDER_STATUS
	orderRequest.CreatedAt = time.Now()
	orderRequest.Id = rand.Intn(100)

	log.Println("Trying to marshall the order request...")
	orderMarshall, err := dynamodbattribute.MarshalMap(orderRequest)
	if err != nil {
		return handleError(err)
	}

	input := &dynamodb.PutItemInput{
		Item:      orderMarshall,
		TableName: aws.String(tableName),
	}

	log.Println("Trying to save the order item in the dynamoDb table...")
	_, err = dynamodbCLient.PutItem(input)

	if err != nil {
		return handleError(err)
	}

	log.Printf("Order item id %v saved!", orderRequest.Id)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
	}, nil
}

func handleError(err error) (events.APIGatewayProxyResponse, error) {
	log.Panicf("unable to process request %v", err)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
