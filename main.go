package main

import (
	"github.com/Shopify/sarama"
)

func main() {

	inTopicName := "payloads.simpleH264"
	outTopicName := "video"
	kafkaBrokers := []string{"localhost:9092"}

	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		panic(err)
	}

	

}