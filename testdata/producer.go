package testdata

import (
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
)

// KafkaMessage defines the message structure
type KafkaMessage struct {
	Ingest    string `json:"ingest"`
	IngestID  string `json:"ingestId"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"`
	Metadata  string `json:"metadata"`
}

// SendMessage will send the given key and message pair
func SendMessage(producer sarama.SyncProducer, topic, mKey string, mValue KafkaMessage) {
	mValueMarshal := &mValue

	mValueMarshalled, errJSONMarshal := json.Marshal(mValueMarshal)
	if errJSONMarshal != nil {
		log.Fatal(errJSONMarshal)

		return
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(mKey),
		Value: sarama.StringEncoder(mValueMarshalled),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("error sending msg %s - %s (%d, %d", msg.Key, err, partition, offset)
	}

	log.Printf("Message sent to partition %d, offset %d", partition, offset)
}
