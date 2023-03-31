package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"log"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type LogMessage struct {
	Application string `json:"application"`
	Level       string `json:"level"`
	Message     string `json:"message"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Register HTTP handlers
	http.HandleFunc("/log", logHandler(logger))
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Register AMQP consumer
	amqpConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer amqpConn.Close()

	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare(
		"log_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to declare a queue", zap.Error(err))
	}

	msgs, err := amqpChannel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to register a consumer", zap.Error(err))
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var logMessage LogMessage
			if err := json.Unmarshal(d.Body, &logMessage); err != nil {
				logger.Error("Failed to unmarshal log message", zap.Error(err))
				continue
			}
			saveLog(logger, logMessage)
		}
	}()

	logger.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func logHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var logMessage LogMessage
		err = json.Unmarshal(body, &logMessage)
		if err != nil {
			logger.Error("Failed to unmarshal log message", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		saveLog(logger, logMessage)
		w.WriteHeader(http.StatusAccepted)
	}
}

func saveLog(logger *zap.Logger, logMessage LogMessage) {
	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Failed to open log file", zap.Error(err))
		return
	}
	defer logFile.Close()

	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("%s [%s] [%s] %s\n",
		timestamp, logMessage.Application, logMessage.Level, logMessage.Message)
	_, err = logFile.WriteString(logLine)
	if err != nil {
		logger.Error("Failed to write log", zap.Error(err))
		return
	}

	switch logMessage.Level {
	case "debug":
		logger.Debug(logMessage.Message, zap.String("application", logMessage.Application))
	case "info":
		logger.Info(logMessage.Message, zap.String("application", logMessage.Application))
	case "warn":
		logger.Warn(logMessage.Message, zap.String("application", logMessage.Application))
	case "error":
		logger.Error(logMessage.Message, zap.String("application", logMessage.Application))
	default:
		logger.Info("Received log with unsupported level",
			zap.String("level", logMessage.Level),
			zap.String("application", logMessage.Application),
		)
	}
}
