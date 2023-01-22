package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sns"
)

type InEvent struct {
	UserId            string            `json:"user_id"`
	Id                string            `json:"id"`
	SendOrderOutEvent SendOrderOutEvent `json:"sendOrderOutEvent"`
	PaymentOutEvent   PaymentOutEvent   `json:"paymentOutEvent"`
	RestaurantId      string            `json:"restaurant_id"`
	CreatedAt         string            `json:"created_at"`
	Quantity          string            `json:"quantity"`
	OrderStatus       string            `json:"order_status"`
	MessageId         string            `json:"message_id"`
}

// Create struct to hold info about new item
type OrderItem struct {
	UserId       string    `json:"user_id"`
	RestaurantId string    `json:"restaurant_id"`
	Quantity     string    `json:"quantity"`
	Id           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	OrderStatus  string    `json:"order_status"`
	MessageId    string    `json:"message_id"`
}

type SendOrderOutEvent struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type PaymentOutEvent struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type ManageStateResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

var (
	tableName      string
	awsRegion      string
	snsTopicArn    string
	dynamodbCLient *dynamodb.DynamoDB
	snsClient      *sns.SNS
	FAILURE        = "FAILURE"
)

func init() {
	tableName = os.Getenv("ORDER_TABLE")
	log.Printf("TABLE NAME : %v", tableName)
	awsRegion = os.Getenv("AWS_REGION")
	log.Printf("AWS REGION : %v", awsRegion)
	snsTopicArn = os.Getenv("SNS_TOPIC_ARN")
	log.Printf("SNS TOPIC ARN : %v", snsTopicArn)

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	if err != nil {
		panic("Unable to initalize the configuration")
	}
	// Create DynamoDB client
	dynamodbCLient = dynamodb.New(awsSession)

	// Create SNS client
	snsClient = sns.New(awsSession)
}

func HandleRequest(event InEvent) (ManageStateResponse, error) {

	fmt.Printf("Event body from State machine : %+v\n", event)

	// If it gets 'SendOrderOutEvent' payload in an Event, it will call this block
	if (SendOrderOutEvent{}) != event.SendOrderOutEvent {
		return HandleSendOrderRestaurant(event)
	}

	// If it gets 'PaymentOutEvent' payload in an Event, it will call this block
	if (PaymentOutEvent{}) != event.PaymentOutEvent && event.PaymentOutEvent.ErrorMessage == "error" {
		return HandlePaymentError(event)
	}

	// If it is first calling of ManageState function, it will put the Item with 'pending' status to Dynamodb
	return SaveOrder(event)
}

func HandleSendOrderRestaurant(event InEvent) (ManageStateResponse, error) {

	var orderStatus string
	var inputParams *dynamodb.UpdateItemInput
	// If it gets an error from sendOrderToRestaurant function, it updates the Dynamodb with failure message
	if event.SendOrderOutEvent.Status == "error" {

		orderStatus = FAILURE
		errorMessage := event.SendOrderOutEvent.ErrorMessage

		// sets the order status to Failure, and adds 'errorMessage'
		inputParams = &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":s": {
					S: aws.String(orderStatus),
				},
				":m": {
					S: aws.String(errorMessage),
				},
			},
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"user_id": {
					S: aws.String(event.UserId),
				},
				"id": {
					S: aws.String(event.Id),
				},
			},
			UpdateExpression: aws.String("set order_status=:s , error_message=:m"),
			ReturnValues:     aws.String("UPDATED_NEW"),
		}
	} else {
		orderStatus = "SUCCESS"

		// If it doesn't get any failure, it will return with 'Success' message
		inputParams = &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":s": {
					S: aws.String(orderStatus),
				},
			},
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"user_id": {
					S: aws.String(event.UserId),
				},
				"id": {
					S: aws.String(event.Id),
				},
			},
			UpdateExpression: aws.String("set order_status=:s "),
			ReturnValues:     aws.String("UPDATED_NEW"),
		}
	}

	return UpdateOrderAndSendNotification(inputParams, orderStatus)
}

func HandlePaymentError(event InEvent) (ManageStateResponse, error) {
	// if it gets an error from PaymentResult, it updates the message with Failure
	orderStatus := FAILURE
	errorMessage := event.PaymentOutEvent.ErrorMessage

	// sets the order status to Failure
	inputParams := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(orderStatus),
			},
			":m": {
				S: aws.String(errorMessage),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(event.UserId),
			},
			"id": {
				S: aws.String(event.Id),
			},
		},
		UpdateExpression: aws.String("set order_status=:s , error_message=:m"),
		ReturnValues:     aws.String("UPDATED_NEW"),
	}

	return UpdateOrderAndSendNotification(inputParams, orderStatus)
}

func UpdateOrderAndSendNotification(input *dynamodb.UpdateItemInput, orderStatus string) (ManageStateResponse, error) {

	log.Println("Trying to update the order item in the dynamoDb table...")
	_, err := dynamodbCLient.UpdateItem(input)

	if err != nil {
		return HandleError(err)
	}

	SendNotification(orderStatus)

	return ManageStateResponse{
		StatusCode: http.StatusOK,
		Body:       "Success for updating order!",
	}, nil
}

func SendNotification(orderStatus string) {

	message := fmt.Sprintf("Order status: %v", orderStatus)

	_, err := snsClient.Publish(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(snsTopicArn),
	})

	if err != nil {
		log.Printf("Error occurred while sending publishing order status , %v", err.Error())
	}
}

func SaveOrder(event InEvent) (ManageStateResponse, error) {
	orderItem := FromInEventToOrderItem(event)

	log.Println("Trying to marshall the order request...")
	orderMarshall, err := dynamodbattribute.MarshalMap(orderItem)
	if err != nil {
		return HandleError(err)
	}

	input := &dynamodb.PutItemInput{
		Item:      orderMarshall,
		TableName: aws.String(tableName),
	}

	log.Println("Trying to save the order item in the dynamoDb table...")
	_, err = dynamodbCLient.PutItem(input)

	if err != nil {
		return HandleError(err)
	}

	log.Printf("Order item id %v saved!", orderItem.Id)

	return ManageStateResponse{
		StatusCode: http.StatusCreated,
		Body:       "Success for saving order!",
	}, nil
}

func HandleError(err error) (ManageStateResponse, error) {
	log.Panicf("Unable to process request %v", err)
	return ManageStateResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       err.Error(),
	}, nil
}

func FromInEventToOrderItem(event InEvent) OrderItem {

	createdAt, err := time.Parse(time.RFC3339, event.CreatedAt)

	if err != nil {
		log.Println(err)
		createdAt = time.Time{}
	}
	return OrderItem{
		UserId:       event.UserId,
		RestaurantId: event.RestaurantId,
		Quantity:     event.Quantity,
		Id:           event.Id,
		CreatedAt:    createdAt,
		OrderStatus:  event.OrderStatus,
		MessageId:    event.MessageId,
	}
}

func main() {
	lambda.Start(HandleRequest)
}
