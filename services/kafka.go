package services

import (
	"github.com/Shopify/sarama"
)

// Payload defines the data contained in
type Payload struct {
	ID      string
	FrameNo int
	Data    []byte
}

// KafkaOperations defines operations for kafka messaging
type KafkaOperations interface {
	PayloadQueue() chan<- *Payload
}

// KafkaService implements kafka message functionality
type KafkaService struct {
	sp sarama.SyncProducer

	payloads chan *Payload
}

// KafkaMessage defines the message structure
type KafkaMessage struct {
	Ingest    string                 `json:"ingest"`
	IngestID  string                 `json:"ingestId"`
	Timestamp string                 `json:"timestamp"`
	Payload   string                 `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewKafkaService will instantiate an instance using the producer provided
func NewKafkaService(sp sarama.SyncProducer) *KafkaService {
	return &KafkaService{
		sp: sp,

		payloads: make(chan *Payload),
	}
}

// PayloadQueue provides access to load a payload object into the queue for sending
func (k *KafkaService) PayloadQueue() chan<- *Payload {
	return k.payloads
}
