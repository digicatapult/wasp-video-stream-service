package integration_test

import (
	"log"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/digicatapult/wasp-simple-h264-payload-parser/kafka"
	"github.com/digicatapult/wasp-simple-h264-payload-parser/wasp"
)

func TestWaspTransformRelay(t *testing.T) {
	inTopicName := "payloads.simpleH264"
	outTopicName := "video"
	kafkaBrokers := []string{"localhost:9092"}

	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		panic(err)
	}

	producer, err := sarama.NewSyncProducer(kafkaBrokers, nil)
	if err != nil {
		zap.S().Fatalf("problem initiating producer: %s", err)
	}

	relayer := kafka.NewMessager().
		WithConsumer(consumer).
		WithProducer(producer).
		WithRelayTransform(wasp.TransformVideoMessages)

	relayer.Relay(inTopicName, outTopicName, nil)

	wg := &sync.WaitGroup{}
	consumed := 0
	go func() {
		partitionConsumer, err := consumer.ConsumePartition(outTopicName, 0, 0)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := partitionConsumer.Close(); err != nil {
				log.Fatalln(err)
			}
		}()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				log.Printf("Consumed %s message offset %d\n", msg.Topic, msg.Offset)
				consumed++
				wg.Done()
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		producer.SendMessage(&sarama.ProducerMessage{
			Topic: inTopicName,
			Key:   nil,
			Value: nil,
		})
	}

	wg.Wait()
	assert.Equal(t, 1000, consumed)
}
