# Use the official Golang image to build the application
FROM golang AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the application source code
COPY main.go .

# Copy the Go module files and download dependencies
RUN go mod init consumer
RUN go mod tidy


# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o consumer

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/consumer .

# Expose the port if needed (optional, depending on the app's behavior)
# EXPOSE 8080

# Set the entrypoint to the application
ENTRYPOINT ["./consumer"]
