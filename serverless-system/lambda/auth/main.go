package main

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const AUTH_SECRET = "auth_secret"

func HandleRequest(request events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {

	authSecret := request.Headers["AUTH"]

	log.Printf("Auth secret from apigateway event : %v", authSecret)

	if AUTH_SECRET != authSecret {
		return GenerateDeny("user", request.MethodArn)
	} else {
		return GenerateAllow("user", request.MethodArn)
	}
}

func GeneratePolicy(principalId string, effect string, resource string) events.APIGatewayCustomAuthorizerResponse {

	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action:   []string{"execute-api:Invoke"},
				Effect:   effect,
				Resource: []string{resource},
			},
		},
	}

	return authResponse
}

func GenerateAllow(principalId string, resource string) (events.APIGatewayCustomAuthorizerResponse, error) {
	log.Println("Allow access!")
	return GeneratePolicy(principalId, "Allow", resource), nil
}

func GenerateDeny(principalId string, resource string) (events.APIGatewayCustomAuthorizerResponse, error) {
	log.Println("Deny access!")
	return GeneratePolicy(principalId, "Deny", resource), nil
}

func main() {
	lambda.Start(HandleRequest)
}
