package main

import (
	"log"
	"strconv"

	mgo "gopkg.in/mgo.v2"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cp16net/hod-test-app/common"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func fib(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func serveFib(conn *amqp.Connection) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			n, err := strconv.Atoi(string(d.Body))
			failOnError(err, "Failed to convert body to integer")

			common.Logger.Info(" [.] fib(%d)", n)
			response := fib(n)

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(strconv.Itoa(response)),
				})
			failOnError(err, "Failed to publish a message")

			d.Ack(false)
		}
	}()

	common.Logger.Info(" [*] Awaiting RPC requests")
}

// Log struct for log data
type Log struct {
	Message string
}

func logHandler(conn *amqp.Connection, svcMongo *cfenv.Service) {
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

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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

	go func() {
		for d := range msgs {
			common.Logger.Info(" [x] %s", d.Body)
			uri, ok := svcMongo.CredentialString("uri")
			if !ok {
				common.Logger.Error("could not get the mongo connection uri string")
			}
			// mongo connection uri should be in the form of:
			// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
			mongoConn, err := mgo.Dial(uri)
			if err != nil {
				common.Logger.Error("failed to connect to mongo: ", err)
			}
			defer mongoConn.Close()
			mongoConn.SetMode(mgo.Monotonic, true)
			c := mongoConn.DB("logs").C("gologger")
			err = c.Insert(&Log{string(d.Body)})
			if err != nil {
				common.Logger.Fatal(err)
			}
		}
	}()

	common.Logger.Info(" [*] Waiting for logs. To exit press CTRL+C")
}

func main() {
	appEnv, _ := cfenv.Current()
	svcRabbitmq, err := appEnv.Services.WithName("cp16net-rabbitmq")
	if err != nil {
		panic("failed to get the cp16net-rabbitmq service details")
	}
	uri, ok := svcRabbitmq.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for rabbitmq")
	}
	svcMongo, err := appEnv.Services.WithName("cp16net-mongo")
	if err != nil {
		panic("failed to get the cp16net-mongo service details")
	}
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	serveFib(conn)
	logHandler(conn, svcMongo)
	forever := make(chan bool)
	<-forever
}
