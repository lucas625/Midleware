package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lucas625/Middleware/utils"
	"github.com/streadway/amqp"
)

func main() {
	numberOfCalls := 10000 // the number of server calls

	calc := utils.InitCalcValues(make([]float64, numberOfCalls, numberOfCalls)) // object to store the rtts

	// conecting to mom server
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

	// Preparing to read messages from server
	msgFromServer, err := ch.Consume(
		replyQueue.Name, // queue
		"",              // consumer
		true,            // autoAck
		false,           // exclusive
		false,           // noLocal
		false,           // noWait
		nil,             // args
	)
	utils.PrintError(err, "Failed to consume from client.")
	fmt.Println("Client on!")
	// Running
	for i := 0; i < numberOfCalls; i++ {
		// Publishing request
		requestMsg := utils.Message{Client: 0, Value: i}
		requestMsgBytes, err := json.Marshal(requestMsg)
		utils.PrintError(err, "Failed to convert to json.")

		err = ch.Publish(
			"",                // exchange
			requestQueue.Name, // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(requestMsgBytes),
			},
		)
		initialTime := time.Now() //calculating time
		utils.PrintError(err, "Failed to publish message.")

		msg := <-msgFromServer // receiving the response

		endTime := float64(time.Now().Sub(initialTime).Milliseconds()) // RTT
		utils.AddValue(&calc, endTime)                                 // pushing to the stored values

		// getting the response
		var msgResponse utils.Message
		err = json.Unmarshal(msg.Body, &msgResponse)
		utils.PrintError(err, "Failed to parse json.")
		fmt.Println(msgResponse)

		time.Sleep(25 * time.Millisecond) // used to evaluate

	}

	// evaluating
	avrg := utils.CalcAverage(&calc)
	stdv := utils.CalcStandardDeviation(&calc, avrg)

	utils.PrintEvaluation(avrg, stdv)
}
