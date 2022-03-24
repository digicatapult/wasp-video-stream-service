package wasp

import (
	"encoding/base64"
	"encoding/json"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

// Message defines the video message structure
type Message struct {
	Value string
}

// VideoMessage unmarshalling
func VideoMessage(msg *sarama.ConsumerMessage) ([]byte, error) {
	// Unmarshal
	var (
		egress = &Message{}
		err    error
	)

	if err = json.Unmarshal(msg.Value, egress); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal video message")
	}

	egressDecoded, err := base64.StdEncoding.DecodeString(egress.Value)
	if err != nil {
		zap.S().Fatalf("problem with unmarshalling message: %v", err)
	}

	return egressDecoded, nil
}
