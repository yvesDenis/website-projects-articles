package main

import (
	"log"
	"math/rand"

	"github.com/aws/aws-lambda-go/lambda"
)

type InEvent struct{}

type OutEvent struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

var (
	SEND_ORDER_STATE   = [2]string{"ok", "error"}
	ERROR_MESSAGE_LIST = [3]string{"could not contact restaurant", "could not understand order", "unknown error"}
	RandInt            = rand.Intn //useful for testing
)

//we will be returning random value of sendOrder lambda
func HandleRequest(event InEvent) OutEvent {

	var response OutEvent
	sendOrderRandom := RandInt(len(SEND_ORDER_STATE))

	log.Println("Start sending order to restaurant...")

	if SEND_ORDER_STATE[sendOrderRandom] == "error" {
		errorRandom := RandInt(len(ERROR_MESSAGE_LIST))
		response.ErrorMessage = ERROR_MESSAGE_LIST[errorRandom]
	}

	response.Status = SEND_ORDER_STATE[sendOrderRandom]

	log.Println("End sending order to restaurant...")

	return response
}

func main() {
	lambda.Start(HandleRequest)
}
