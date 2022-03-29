package wasp

import (
	"encoding/base64"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Payload defines the data contained in
type Payload struct {
	ID          string
	FrameNo     int
	Filename    string
	Type        string
	Data        []byte `json:"-"`
	EncodedData string `json:"Data"`
}

// Message defines the video message structure
type Message struct {
	Payload *Payload `json:"Value"`
}

// VideoMessage unmarshalling
func VideoMessage(msg *sarama.ConsumerMessage) (*Message, error) {
	// Unmarshal
	var (
		egress = &Message{}
		err    error
	)
	// decoded, err := base64.StdEncoding.DecodeString(string(msg.Value))
	// zap.S().With("msg v", string(msg.Value)).Info("msg v")
	// if err != nil {
	// 	zap.S().Fatalf("problem with unmarshalling message: %v", err)
	// }

	// unquoted, err := strconv.Unquote(string(msg.Value))
	// if err != nil {
	// 	zap.S().Info(string(msg.Value))
	// 	zap.S().Fatalf("problem with unquoting message: %v", err)
	// }

	if err = json.Unmarshal(msg.Value, egress); err != nil {
		// zap.S().With("decoded", decoded).Info("decoded")
		return nil, errors.Wrap(err, "unable to unmarshal video message")
	}

	egress.Payload.Data, err = base64.StdEncoding.DecodeString(egress.Payload.EncodedData)
	if err != nil {
		zap.S().Fatalf("problem with unmarshalling message: %v", err)
	}

	return egress, nil
}
