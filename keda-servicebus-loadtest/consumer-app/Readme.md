go mod init consumer
go mod tidy
go build -o consumer

export SERVICEBUS_QUEUE_NAME=your-queue-name
export SERVICEBUS_CONNECTION_STRING=your-connection-string
./consumer

docker buildx build --push --platform linux/arm/v7,linux/arm64/v8,linux/amd64 -t ams0/consumer-app:0.6 .

docker run -e SERVICEBUS_CONNECTION_STRING=${SERVICEBUS_CONNECTION_STRING} -e SERVICEBUS_QUEUE_NAME=${SERVICEBUS_QUEUE_NAME}  ams0/consumer-app:0.6

