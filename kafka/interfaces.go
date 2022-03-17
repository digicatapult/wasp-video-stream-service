package kafka

// Producer defines the producer operations
type Producer interface {
	SendMessage(topic, msg string, value []byte) error
}

// Consumer defines the consumer operations
type Consumer interface {
	Listen(topic string, received chan []byte)
}
