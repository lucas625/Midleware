package main

import (
	"encoding/json"
	"fmt"

	"github.com/lucas625/Middleware/utils"
	"github.com/streadway/amqp"
)

func main() {
	// Connecting to rabbitmq server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	utils.PrintError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Creating a channel
	ch, err := conn.Channel()
	utils.PrintError(err, "Failed to open a channel.")
	defer ch.Close()

	// Creating queues
	requestQueue, err := ch.QueueDeclare(
		"request", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	utils.PrintError(err, "Failed to declare a queue.")

	replyQueue, err := ch.QueueDeclare(
		"response", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	utils.PrintError(err, "Failed to declare a queue.")

	// Preparing to read messages from client
	msgFromClient, err := ch.Consume(
		requestQueue.Name, // queue
		"",                // consumer
		true,              // autoAck
		false,             // exclusive
		false,             // noLocal
		false,             // noWait
		nil,               // args
	)
	utils.PrintError(err, "Failed to consume from client.")
	fmt.Println("Server on!")

	// server on
	for d := range msgFromClient {
		// receiving request
		var msgRequest utils.Message
		err := json.Unmarshal(d.Body, &msgRequest)
		utils.PrintError(err, "Failed to parse json.")

		// processing request
		fmt.Println(msgRequest)

		// replying
		var replyMsg utils.Message
		replyMsg.Client = msgRequest.Client
		replyMsg.Value = utils.DecodeMessage(&msgRequest)
		replyMsgBytes, err := json.Marshal(replyMsg)
		utils.PrintError(err, "Failed to convert to json.")

		err = ch.Publish(
			"",              // exchange
			replyQueue.Name, // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(replyMsgBytes),
			},
		)
		utils.PrintError(err, "Failed to publish message.")

	}

}
