package rabbitmq

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

const verbose = false

var deliveryCount = 0

const queueName = "oengus.bot" // Also our routing key

func StartListening(dg *discordgo.Session) {
	c, err := NewConsumer("amqp://duncte123:password@localhost:5672/", "amq.topic", "topic", queueName, queueName, "", dg)

	if err != nil {
		log.Fatalf("%s", err)
	}

	SetupCloseHandler(c)

	log.Printf("running until Consumer is done")
	<-c.done

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

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, dg *discordgo.Session) (*Consumer, error) {
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
		return nil, fmt.Errorf("dial: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %s", err)
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
		return nil, fmt.Errorf("queue declare: %s", err)
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
		return nil, fmt.Errorf("queue bind: %s", err)
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
		return nil, fmt.Errorf("queue consume: %s", err)
	}

	go func() {
		//for {
		//	handle(deliveries, c.done)
		//}
		handle(deliveries, c.done, dg)
	}()

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error, dg *discordgo.Session) {
	cleanup := func() {
		log.Printf("handle: deliveries channel closed")
		done <- nil
	}

	defer cleanup()

	for d := range deliveries {
		err := handleIncomingEvent(d.Body, dg)

		if err == nil {
			d.Ack(false)
		} else {
			// Reject failed messages, usually because it's not json
			d.Reject(false)
			fmt.Println("Error handling rmq event", err)
		}

		deliveryCount++
		if verbose {
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
	}
}
