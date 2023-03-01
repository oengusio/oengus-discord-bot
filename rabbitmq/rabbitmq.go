package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

var autoAck = false
var verbose = true
var deliveryCount = 0
var lifetime = 0 * time.Second
var queueName = "oengus.bot"

func StartListening() {
	c, err := NewConsumer("amqp://duncte123:password@localhost:5672/", "amq.topic", "topuc", queueName, queueName, "")
	if err != nil {
		log.Fatalf("%s", err)
	}

	SetupCloseHandler(c)

	if lifetime > 0 {
		log.Printf("running for %s", lifetime)
		time.Sleep(lifetime)
	} else {
		log.Printf("running until Consumer is done")
		<-c.done
	}

	log.Printf("shutting down")

	if err := c.Shutdown(); err != nil {
		log.Fatalf("error during shutdown: %s", err)
	}
}

func SetupCloseHandler(consumer *Consumer) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Ctrl+C pressed in Terminal")
		if err := consumer.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
		os.Exit(0)
	}()
}

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName("sample-consumer")
	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	/*if err = c.channel.ExchangeDeclare(
	  	exchange,     // name of the exchange
	  	exchangeType, // type
	  	true,         // durable
	  	false,        // delete when complete
	  	false,        // internal
	  	false,        // noWait
	  	nil,          // arguments
	  ); err != nil {
	  	return nil, fmt.Errorf("Exchange Declare: %s", err)
	  }

	  log.Printf("declared Exchange, declaring Queue %q", queueName)*/
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queueName, // name of the queue
		key,       // bindingKey
		exchange,  // sourceExchange
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queueName, // name
		c.tag,     // consumerTag,
		false,     // autoAck
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go func() {
		for {
			handle(deliveries, c.done)
		}
	}()

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	cleanup := func() {
		log.Printf("handle: deliveries channel closed")
		done <- nil
	}

	defer cleanup()

	for d := range deliveries {
		deliveryCount++
		if verbose == true {
			log.Printf(
				"got %dB delivery: [%v] %q",
				len(d.Body),
				d.DeliveryTag,
				d.Body,
			)
		} else {
			if deliveryCount%65536 == 0 {
				log.Printf("delivery count %d", deliveryCount)
			}
		}
		if autoAck == false {
			d.Ack(false)
		}
	}
}
