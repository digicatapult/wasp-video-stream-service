package wasp

import (
	"encoding/json"
	"github.com/Shopify/sarama"

	"github.com/pkg/errors"
)

// Message defines the video message structure
type Message struct {
	Value []byte
}

// VideoMessage unmarshalling
func VideoMessage(msg *sarama.ConsumerMessage) ([]byte, error) {
	// Unmarshal
	var (
		egress = &Message{}
		err    error
	)

	if err = json.Unmarshal(msg.Value, egress); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal ingest")
	}

	return egress.Value, nil
}
