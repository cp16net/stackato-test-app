package rabbitmq

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cp16net/hod-test-app/common"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		common.Logger.Fatalf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// FibonacciRPC call to amqp
func FibonacciRPC(n int) (res int, err error) {
	conn, err := amqp.Dial(amqpuri)
	// conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrID := randomString(32)

	err = ch.Publish(
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			ReplyTo:       q.Name,
			Body:          []byte(strconv.Itoa(n)),
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrID == d.CorrelationId {
			res, err = strconv.Atoi(string(d.Body))
			failOnError(err, "Failed to convert body to integer")
			break
		}
	}

	return
}

var amqpuri string

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	appEnv, _ := cfenv.Current()
	svc, err := appEnv.Services.WithName("cp16net-rabbitmq")
	if err != nil {
		panic("failed to get the cp16net-rabbitmq service details")
	}
	uri, ok := svc.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for rabbitmq")
	}
	amqpuri = uri
}

// Log struct for log data
type Log struct {
	Message string
}

// WriteLogs writes number of log messages to amqp to be stored
func WriteLogs(num int) {
	conn, err := amqp.Dial(amqpuri)
	// conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	for index := 0; index < num; index++ {
		message := randomString(64)
		err = ch.Publish(
			"logs", // exchange
			"",     // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			})
		failOnError(err, "Failed to publish a message")
		common.Logger.Infof(" [x] Sent %s", message)
	}

}
