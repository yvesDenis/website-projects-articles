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
	"github.com/aws/aws-sdk-go/service/sfn"
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

type SqsEventWrapper struct {
	Data OrderRequest `json:"data"`
}

type Response struct {
	StatusCode int
}

const DEFAULT_ORDER_STATUS = "PENDING"

var (
	stateMachineArn    string
	awsRegion          string
	endpoint           string
	stepfunctionClient *sfn.SFN
)

func init() {

	stateMachineArn = os.Getenv("STATE_MACHINE_ARN")
	log.Printf("STATE_MACHINE_ARN : %v", stateMachineArn)
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
	// Create Stepfunction client
	stepfunctionClient = sfn.New(awsSession)
}

func HandleRequest(request events.SQSEvent) (Response, error) {

	var sqsEventWrapper SqsEventWrapper
	message := request.Records[0]

	log.Printf("Order request from SQS Queue : %v", message.Body)
	if err := json.Unmarshal([]byte(message.Body), &sqsEventWrapper); err != nil {
		return handleError(err)
	}

	orderRequest := sqsEventWrapper.Data

	orderRequest.OrderStatus = DEFAULT_ORDER_STATUS
	orderRequest.CreatedAt = time.Now()
	orderRequest.Id = message.MessageId

	log.Println("Trying to serialize order request...")

	orderRequestBytes, err := json.Marshal(orderRequest)

	if err != nil {
		return handleError(err)
	}

	bytes := string(orderRequestBytes)

	log.Printf("Start stepfunction execution with SQS message id %v ...", message.MessageId)

	startExecutionInput := sfn.StartExecutionInput{
		Input:           &bytes,
		Name:            &message.MessageId,
		StateMachineArn: &stateMachineArn,
	}

	startExecutionOutput, err := stepfunctionClient.StartExecution(&startExecutionInput)
	if err != nil {
		return handleError(err)
	}

	log.Printf("Stepfunction execution ARN : %v , started on %v ", *startExecutionOutput.ExecutionArn, *startExecutionOutput.StartDate)

	return Response{
		StatusCode: http.StatusAccepted,
	}, nil
}

func handleError(err error) (Response, error) {
	log.Panicf("Unable to process request %v", err)
	return Response{
		StatusCode: http.StatusInternalServerError,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
