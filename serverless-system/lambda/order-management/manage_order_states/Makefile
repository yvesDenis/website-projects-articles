file := ./events/definition.asl.txt

default: docker_network docker_run aws_step_function_create_machine aws_sam_build aws_lambda_invoke

local_stack_servces_start:
	docker-compose up -d
aws_dynamodb_create_table:
	aws dynamodb create-table --table-name order-table --attribute-definitions AttributeName=user_id,AttributeType=S AttributeName=id,AttributeType=S --key-schema AttributeName=user_id,KeyType=HASH AttributeName=id,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 --endpoint-url http://localhost:4566
aws_sns_create_topic:
	aws sns create-topic --name topic-test --endpoint-url http://localhost:4566
aws_step_function_create_machine:
#This target generates a json output with the state machine arn necessary for the last target
	aws stepfunctions create-state-machine --name manageStateOrderAsl --definition $(shell cat ${file}) --role-arn "arn:aws:iam::000000000000:role/DummyRole" --endpoint-url http://localhost:4566
aws_sam_build:
#Installing samlocal is a prerequisite to run successfully this target
	samlocal build -t template_local.yaml --docker-network networktest  --parameter-overrides StateMachineArn=arn:aws:states:ca-central-1:000000000000:stateMachine:manageStateOrderAsl OrderSnsTopicArn=arn:aws:sns:ca-central-1:000000000000:topic-test
aws_sam_deploy_on_localstack:
	samlocal deploy --stack-name deploy-lambda --resolve-s3 --region ca-central-1
aws_stepfunction_start_execution_save_order:
	aws stepfunctions start-execution --state-machine-arn arn:aws:states:ca-central-1:000000000000:stateMachine:manageStateOrderAsl --input "$(shell cat ${file})" --endpoint-url http://localhost:4566
aws_local_lambda_invoke:
	samlocal invoke --docker-network networktest -e events/event.json


arn:aws:dynamodb:ca-central-1:000000000000:table/order-table
"TopicArn": "arn:aws:sns:ca-central-1:000000000000:topic-test"
{
    "stateMachineArn": "arn:aws:states:ca-central-1:000000000000:stateMachine:manageStateOrderAsl",
    "creationDate": "2023-01-10T23:06:43.205000-05:00"
}