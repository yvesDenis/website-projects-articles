USER_POOL_ID := ""
CLIENT_ID := ""
CLIENT_SECRET := ""
SECRET_HASH := ""
IDTOKEN := ""
APIGATEWAY_ENDPOINT := https://XXXXXX.execute-api.ca-central-1.amazonaws.com/Dev/orders/

default: retrieve_user_pool_id retrieve_client_id retrieve_client_secret create_secret_hash user_signup retrieve_token test_get_orders

retrieve_user_pool_id:
	$(eval USER_POOL_ID := $(shell aws cloudformation describe-stacks --stack-name serverlessSystemStack --region $(AWS_REGION) | jq -r '.Stacks[0].Outputs[] | select( .OutputKey | contains("CognitoUserPoolID"))' | jq -r ".OutputValue"))
retrieve_client_id:
	$(eval CLIENT_ID := $(shell aws cloudformation describe-stacks --stack-name serverlessSystemStack --region $(AWS_REGION) | jq -r '.Stacks[0].Outputs[] | select( .OutputKey | contains("CognitoClientID"))' | jq -r ".OutputValue"))
retrieve_client_secret:
	$(eval CLIENT_SECRET := $(shell aws cognito-idp describe-user-pool-client --user-pool-id $(USER_POOL_ID) --client-id $(CLIENT_ID)  --region $(AWS_REGION) | jq -r ".UserPoolClient.ClientSecret"))
create_secret_hash:
	$(eval message := $(shell echo $(USERNAME)$(CLIENT_ID)))
	$(eval SECRET_HASH := $(shell echo -n $(message) | openssl dgst -sha256 -hmac $(CLIENT_SECRET) -binary | base64))
user_signup:
	aws cognito-idp sign-up \
    --client-id $(CLIENT_ID) \
    --secret-hash $(SECRET_HASH) \
    --username $(USERNAME) \
    --password $(PASSWORD) \
    --user-attributes Name=email,Value=$(USERNAME) \
    --region $(AWS_REGION)
retrieve_token:
	$(eval IDTOKEN := $(shell aws cognito-idp admin-initiate-auth --user-pool-id $(USER_POOL_ID) --client-id $(CLIENT_ID) --auth-flow ADMIN_NO_SRP_AUTH -auth-parameters USERNAME=$(USERNAME),PASSWORD=$(PASSWORD),SECRET_HASH=$(SECRET_HASH) --region $(AWS_REGION) | jq -r ".AuthenticationResult.IdToken"))
test_get_orders:
	curl --location --request POST '$(APIGATEWAY_ENDPOINT)' --header 'Authorization: Bearer $(IDTOKEN)' --header 'Content-Type: application/json' --data-raw '{"user_id": "Burger_18","quantity": "3","restaurant_id": "Restaurant 4"}'