FROM golang:1.19.3-alpine3.17 as stage 

LABEL mantainer="yvesdenis@gmail.com" \
      verson="1.0.0"  \
      description="Simple http server"
    
WORKDIR /app

COPY ./deploy-app-aws-ecs/base-app/ /app/base-app/

WORKDIR /app/base-app

#libc is missing when you package your application to Docker image FROM scratch
#CGO_ENABLED is set to 0 to disable the cgo tool
RUN CGO_ENABLED=0 go build -o main-app

FROM scratch as runtime

COPY --from=stage /app/base-app/main-app  /

EXPOSE 8080

ENTRYPOINT [ "./main-app" ]




