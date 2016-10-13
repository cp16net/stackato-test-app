package main

import (
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"

	"github.com/cp16net/hod-test-app/common"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Log struct for log data
type Log struct {
	Message string
}

func main() {
	appEnv, _ := cfenv.Current()
	svcRabbitmq, err := appEnv.Services.WithName("cp16net-rabbitmq")
	if err != nil {
		panic("failed to get the cp16net-rabbitmq service details")
	}
	svcMongo, err := appEnv.Services.WithName("cp16net-mongo")
	if err != nil {
		panic("failed to get the cp16net-mongo service details")
	}
	mongouri, ok := svcMongo.CredentialString("uri")
	if !ok {
		common.Logger.Error("could not get the mongo connection uri string")
	}

	// TODO: this is a hack for the mongodb uri
	// made PR to fix this in csm
	mongouri = mongouri[:len(mongouri)-1]

	mongodbname, ok := svcMongo.CredentialString("db")
	if !ok {
		panic("failed to get the credential name of db for mongo")
	}

	rabbitmquri, ok := svcRabbitmq.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for rabbitmq")
	}
	conn, err := amqp.Dial(rabbitmquri)
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

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			common.Logger.Info(" [x] %s", d.Body)
			common.Logger.Debug("mongouri: ", mongouri)
			// mongo connection uri should be in the form of:
			// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
			mongoConn, err := mgo.Dial(mongouri)
			if err != nil {
				common.Logger.Error("failed to connect to mongo: ", err)
			}
			defer mongoConn.Close()
			// mongoConn.SetMode(mgo.Monotonic, true)
			c := mongoConn.DB(mongodbname).C("gologger")
			err = c.Insert(&Log{string(d.Body)})
			if err != nil {
				common.Logger.Fatal(err)
			}
		}
	}()

	common.Logger.Info(" [*] Waiting for logs...")
	<-forever
}
