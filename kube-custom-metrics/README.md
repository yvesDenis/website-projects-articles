# Autoscaling streaming application with the Horizontal Pod Autoscaler and custom metrics on Kubernetes

Prerequisites

You should have kubernetes cluster installed. I personally use [kind](https://kind.sigs.k8s.io) 

## Installing Kubernetes cluster with kind:

`kind create cluster --name custom-metrics --config ./kind.yml`

## Installing Custom Metrics Api

Deploy the Metrics Server in the kube-system namespace:

```
kubectl create -f monitoring/metrics-server
After one minute the metric-server starts reporting CPU and memory usage for nodes and pods.
```

View nodes metrics:

`kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes" | jq .`

View pods metrics:

`kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods" | jq .`

Create the monitoring namespace:

`kubectl create -f monitoring/namespaces.yaml`

Deploy Prometheus v2 in the monitoring namespace:

`kubectl create -f monitoring/prometheus`

Deploy the Prometheus custom metrics API adapter:

`kubectl create -f monitoring/prometheus-adapter`

List the custom metrics provided by Prometheus:

`kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .`

## Deploy the streaming application

The application consists of an activeMq Server and a Spring boot application working the jms template.
The spring boot app exposes an endpoint from which messages are sent by an external client. This messages are then sent fromatted, sent to the queue, consumed and finally logged in the app.

To make it work well , the deployemnt order is crucial:

### Deploying the ActiveMQ server

```
kubectl apply -f ./deployemnt-activemq.yml

After few minutes, in order to make it interact with the spring boot app, you shoud create a queue channel. To do so ActiveMq has a web interface, you can access it via port-forwarding:

kubectl port-forward svc/activemq 8161:8161

You can visit the ActiveMQ UI iterface at http://localhost:8161

```

### Deploying the Sprint boot app

After deploting the Streaming server, you can now deploy the spring boot app:

`kubectl apply -f /deployment-app.yml`

You could send requests after opening the 8080 port.

```
kubectl port-forward svc/demo-app 8080:8080

You can post messages to the queue by via:

curl --location 'http://localhost:8080/message/deffo/counter/3' \
--header 'Content-Type: application/json' \
--data '{
    "playerName": "denis",
    "amount": "4000",
    "transactionLogId": "4"
}'
```

### Deploying the HPA

You can scale the application in proportion to the number of messages in the queue with the Horizontal Pod Autoscaler. You can deploy the HPA with:

`kubectl apply -f ./hpa.yml`

You should be able to see the number of pending messages from http://localhost:8080/metrics and from the custom metrics endpoint:

kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/messages" | jq .
Autoscaling workers

You may need to wait three minutes before you can see more pods joining the deployment with:

kubectl get pods
The autoscaler will remove pods from the deployment every 5 minutes.

You can inspect the event and triggers in the HPA with:

kubectl get hpa spring-boot-hpa
Notes

The configuration for metrics and metrics server is configured to run on minikube only.

You won't be able to run the same YAML files for metrics and custom metrics server on your cluster or EKS, GKE, AKS, etc.

Also, there are secrets checked in the repository to deploy the Prometheus adapter.

In production, you should generate your own secrets and (possibly) not check them into version control.