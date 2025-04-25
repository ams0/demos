package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func main() {
	queueName := os.Getenv("SERVICEBUS_QUEUE_NAME")
	if queueName == "" {
		if len(os.Args) > 1 {
			queueName = os.Args[1]
		} else {
			fmt.Println("Error: Queue name must be provided via SERVICEBUS_QUEUE_NAME environment variable or as a command-line argument.")
			os.Exit(1)
		}
	}

	// Read processing delay from environment variable, default to 0ms
	processingDelayMsStr := os.Getenv("PROCESSING_DELAY_MS")
	processingDelayMs := 0 // Default value is 0
	var err error
	if processingDelayMsStr != "" {
		processingDelayMs, err = strconv.Atoi(processingDelayMsStr)
		if err != nil {
			fmt.Printf("Error parsing PROCESSING_DELAY_MS: %v. Using default 0ms.\n", err)
			processingDelayMs = 0
		}
	}
	fmt.Printf("Using processing delay: %dms\n", processingDelayMs)
	delayDuration := time.Duration(processingDelayMs) * time.Millisecond

	connStr := os.Getenv("SERVICEBUS_CONNECTION_STRING")
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connStr))
	if err != nil {
		fmt.Println("namespace: ", err)
		panic(err)
	}

	fmt.Println("connecting to queue: ", queueName)
	q, err := ns.NewQueue(queueName)
	if err != nil {
		// handle queue creation error
		fmt.Println("create queue: ", err)
		// It might be better to exit if the queue cannot be accessed
		panic(err)
	}

	fmt.Println("setting up listener")
	var messageHandler servicebus.HandlerFunc = func(ctx context.Context, msg *servicebus.Message) error {
		fmt.Println("received message: ", string(msg.Data))

		// Introduce processing delay
		if delayDuration > 0 {
			fmt.Printf("processing delay: sleeping for %v\n", delayDuration)
			time.Sleep(delayDuration)
		}

		fmt.Println("processing complete for message: ", string(msg.Data))
		return msg.Complete(ctx)
	}

	// Use a cancellable context for the receiver
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start receiving messages in a separate goroutine
	go func() {
		err = q.Receive(ctx, messageHandler)
		if err != nil {
			// Log the error unless it's a context cancellation
			if ctx.Err() == nil {
				fmt.Println("listener error: ", err)
				// Consider if this should cause the application to exit
				// For now, just log it. The main loop will still wait for signals.
			}
		}
	}()

	fmt.Println("listening...")

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Signal received, cancel the context to stop the receiver gracefully
	fmt.Println("Received shutdown signal, stopping listener...")
	cancel()

	// Optional: Add a small delay to allow the receiver to shut down cleanly
	// Adjust the duration as needed
	time.Sleep(2 * time.Second)
	fmt.Println("Exiting.")
}
