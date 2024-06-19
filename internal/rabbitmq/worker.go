package rabbitmq

import (
	"encoding/json"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	customLog "golang-api-starter/internal/modules/log"
	"log"
	"time"
)

type UploadRequest struct {
	File []byte `json:"file"`
}

var rbChanns = map[string]func(){"test_queue": HandleTestQueue, "log_queue": HandleLogQueue}

func HandleTestQueue() {
	rabbitMQ, err := NewRabbitMQ(GetUrl(), "test_queue")
	if err != nil {
		logger.Fatalf(err.Error())
	}
	defer rabbitMQ.Close()

	msgs, err := rabbitMQ.channel.Consume(
		rabbitMQ.queue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Test worker started. Waiting for messages...")

	for msg := range msgs {
		var uploadRequest UploadRequest
		err := json.Unmarshal(msg.Body, &uploadRequest)
		if err != nil {
			log.Println(err)
			continue
		}

		// Process the file upload here
		logger.Infof("Received file upload: %d bytes\n", len(uploadRequest.File))

		// Simulate processing time
		time.Sleep(2 * time.Second)

		logger.Infof("File upload processed successfully")
	}
}

func HandleLogQueue() {
	rabbitMQ, err := NewRabbitMQ(GetUrl(), "log_queue")
	if err != nil {
		logger.Fatalf(err.Error())
	}
	defer rabbitMQ.Close()

	msgs, err := rabbitMQ.channel.Consume(
		rabbitMQ.queue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Log worker started. Waiting for logs...")

	for msg := range msgs {
		var log = new(customLog.Log)
		if err := json.Unmarshal(msg.Body, log); err != nil {
			logger.Errorf("failed to Unmarshal log, err: %+v", err)
		}

		// Process the file upload here
		logger.Infof("Received log: %d bytes\n", *log)

		// Simulate processing time
		time.Sleep(2 * time.Second)

		// create log to database,
		customLog.Srvc.Create([]*customLog.Log{log})
		logger.Infof("log processed successfully")
	}
}

func RunWorker() {
	// Open a dummy channel to hold this RunWorker without exit
	forever := make(chan bool)

	for chanName, handler := range rbChanns {
		logger.Infof("handling queue: %+v", chanName)
		go handler()
	}

	<-forever
}
