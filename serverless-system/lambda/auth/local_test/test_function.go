package main

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.Println("Succesfull mock request!!")
	return events.APIGatewayProxyResponse{}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
