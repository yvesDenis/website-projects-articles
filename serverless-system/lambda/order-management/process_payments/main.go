package main

import (
	"log"
	"math/rand"

	"github.com/aws/aws-lambda-go/lambda"
)

type InEvent struct{}

type PaymentOutEvent struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

var (
	PAYMENT_STATE      = [2]string{"ok", "error"}
	ERROR_MESSAGE_LIST = [3]string{"could not contact payment processor", "payment method declined", "unknown error"}
	RandInt            = rand.Intn //useful for testing
)

//we will be returning random value of payment method
func HandleRequest(event InEvent) PaymentOutEvent {

	var response PaymentOutEvent
	paymentRandom := RandInt(len(PAYMENT_STATE))

	log.Println("Start Processing payment...")

	if PAYMENT_STATE[paymentRandom] == "error" {
		errorRandom := RandInt(len(ERROR_MESSAGE_LIST))
		response.ErrorMessage = ERROR_MESSAGE_LIST[errorRandom]
	}

	response.Status = PAYMENT_STATE[paymentRandom]

	log.Println("End Processing payment...")

	return response
}

func main() {
	lambda.Start(HandleRequest)
}
