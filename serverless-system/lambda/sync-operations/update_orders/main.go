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
)

var (
	tableName      string
	awsRegion      string
	endpoint       string
	dynamodbCLient *dynamodb.DynamoDB
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
		panic("Unable to initalize the configuration")
	}
	// Create DynamoDB client
	dynamodbCLient = dynamodb.New(awsSession)
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var orderUpdateRequest OrderRequest

	log.Printf("Order update request from the API GATEWAY : %v", request.Body)
	if err := json.Unmarshal([]byte(request.Body), &orderUpdateRequest); err != nil {
		return handleError(err)
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":q": {
				S: aws.String(orderUpdateRequest.Quantity),
			},
			":r": {
				S: aws.String(orderUpdateRequest.RestaurantId),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(orderUpdateRequest.UserId),
			},
			"id": {
				S: aws.String(orderUpdateRequest.Id),
			},
		},
		UpdateExpression: aws.String("set quantity=:q , restaurant_id=:r"),
	}

	log.Println("Trying to update the order item in the dynamoDb table...")
	_, err := dynamodbCLient.UpdateItem(input)

	if err != nil {
		return handleError(err)
	}

	log.Printf("Order item id %v updated!", orderUpdateRequest.Id)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
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
