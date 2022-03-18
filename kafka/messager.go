package kafka

import (
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RelayTransform aliases the function for transforming relayed messages
type RelayTransform func(msg []byte) ([]byte, error)

// Messager implements kafka message functionality
type Messager struct {
	sp sarama.SyncProducer
	c  sarama.Consumer

	rt RelayTransform
}

// NewMessager will instantiate an instance using the producer provided
func NewMessager() *Messager {
	return &Messager{}
}

// WithConsumer will set the consumer instance on the messager
func (m *Messager) WithConsumer(c sarama.Consumer) *Messager {
	m.c = c

	return m
}

// WithProducer will set the producer instance on the messager
func (m *Messager) WithProducer(p sarama.SyncProducer) *Messager {
	m.sp = p

	return m
}

// WithRelayTransform will set the relay transform func instance on the messager
func (m *Messager) WithRelayTransform(rt RelayTransform) *Messager {
	m.rt = rt

	return m
}

// SendMessage can send a message to the
func (m *Messager) SendMessage(topic, mKey string, mValue []byte) error {
	if m.sp == nil {
		panic("nil producer")
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(mKey),
		Value: sarama.StringEncoder(mValue),
	}

	partition, offset, err := m.sp.SendMessage(msg)
	if err != nil {
		err = errors.Wrap(err, "problem sending message")
		zap.S().Errorf("error sending msg %s - %s (%d, %d", msg.Key, err, partition, offset)

		return err
	}

	zap.S().Debugf("Message sent to partition %d, offset %d", partition, offset)

	return nil
}

// Listen will listen indefinitely to the kafka bus for messages on a specific topic
func (m *Messager) Listen(topic string, received chan []byte, signals chan bool) {
	if m.c == nil {
		panic("nil consumer")
	}

	partitionConsumer, err := m.c.ConsumePartition(topic, 0, 0)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	consumed := 0

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed %s message offset %d\n", msg.Topic, msg.Offset)
			consumed++
			received <- msg.Value
		case <-signals:
			return
		}
	}
}

// Relay will start listening for messages, transforming, and sending onwards
func (m *Messager) Relay(inTopic, outTopic string, stop chan os.Signal) error {
	receivedMsgs := make(chan []byte)
	signals := make(chan bool)

	m.validateRelay()

	go m.Listen(inTopic, receivedMsgs, signals)

	go func() {
		for {
			select {
			case msg := <-receivedMsgs:
				m.executeTransform(msg, outTopic)
			case <-stop:
				signals <- true

				close(receivedMsgs)

				return
			}
		}
	}()

	return nil
}

func (m *Messager) executeTransform(msg []byte, outTopic string) {
	transformed, err := m.rt(msg)
	if err != nil {
		zap.S().Error(err)
	}

	err = m.SendMessage(outTopic, "", transformed)
	if err != nil {
		zap.S().Error(err)
	}
}

func (m *Messager) validateRelay() {
	if m.c == nil {
		panic("nil consumer")
	}

	if m.sp == nil {
		panic("nil producer")
	}

	if m.rt == nil {
		panic("nil relay transform")
	}
}
