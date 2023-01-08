default: docker_network docker_run aws_create_table aws_insert_orders aws_lambda_invoke

docker_network:
	docker network create networktest
docker_run:
	docker run -d -p 8000:8000 --network networktest --name dynamodblocal amazon/dynamodb-local
aws_create_table:
	aws dynamodb create-table --table-name order-table --attribute-definitions AttributeName=user_id,AttributeType=S AttributeName=id,AttributeType=S --key-schema AttributeName=user_id,KeyType=HASH AttributeName=id,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 --endpoint-url http://localhost:8000 --no-paginate
aws_insert_orders:
	aws dynamodb put-item --table-name order-table --item file://events/items.json  --endpoint-url http://localhost:8000
aws_build:
	AWS_REGION=ca-central-1 sam build --docker-network networktest -t template_local.yml
aws_lambda_invoke:
	AWS_REGION=ca-central-1 sam local invoke --docker-network networktest -e events/events.json
aws_get_item_test:
	AWS_REGION=ca-central-1 aws dynamodb get-item --table-name order-table --key file://events/get_items.json --endpoint-url http://localhost:8000