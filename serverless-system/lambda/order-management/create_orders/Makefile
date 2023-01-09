file := ./events/definition.asl.txt

default: docker_network docker_run aws_step_function_create_machine aws_sam_build aws_lambda_invoke

docker_network:
	docker network create networktest
docker_run:
	docker run -d -p 8000:8083 --network networktest --name stepfunctionlocal -e AWS_DEFAULT_REGION=ca-central-1 amazon/aws-stepfunctions-local
aws_step_function_create_machine:
#This target generates a json output with the state machine arn necessary for the last taregt
	aws stepfunctions create-state-machine --name createOrderAsl --definition $(shell cat ${file}) --role-arn "arn:aws:iam::012345678901:role/DummyRole" --endpoint-url http://localhost:8000 --region ca-central-1
aws_sam_build:
	sam build -t template_local.yml --docker-network networktest 
#You can generate the events.json file with the command : sam local generate-event sqs receive-message --body "{\"user_id\": \"Burger_2\",\"quantity\": \"3\",\"restaurant_id\": \"Restaurant 3\"}"
aws_lambda_invoke:
	AWS_REGION=ca-central-1 sam local invoke --docker-network networktest  -e events/event.json --parameter-overrides 'ParameterKey=StateMachineArn,ParameterValue=${stateMachineArn}'