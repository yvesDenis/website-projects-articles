FROM public.ecr.aws/lambda/provided:al2 as build
# install compiler
RUN yum install -y golang
RUN go env -w GOPROXY=direct
# cache dependencies
COPY ./serverless-system/lambda/create_orders/go.mod ./serverless-system/lambda/create_orders/go.sum ./
RUN go mod download
# build
COPY ./serverless-system/lambda/create_orders/create_order.go ./
RUN go build -o /main
# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /main /main
ENTRYPOINT [ "/main" ]           