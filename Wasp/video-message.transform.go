package wasp

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// IngestMessage defines the video ingest message structure
type IngestMessage struct {
	Ingest    string                 `json:"ingest"`
	IngestID  string                 `json:"ingestId"`
	Timestamp string                 `json:"timestamp"`
	Payload   string                 `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
	ThingID   string                 `json:"thingId"`
	Type      string                 `json:"type"`
}

// OutputMessage defines the output message structure
type OutputMessage struct {
	ThingID   string
	Type      string
	Timestamp string
	Value     string
	Metadata  map[string]interface{}
}

// TransformVideoMessages processes the messages as part of the message relay
func TransformVideoMessages(msg []byte) ([]byte, error) {
	// Transform messages here
	// Unmarshal
	var (
		ingest = &IngestMessage{}
		err    error
	)

	if err = json.Unmarshal(msg, ingest); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal ingest")
	}
	// Modify values
	// Create new message
	output := &OutputMessage{
		ThingID:   ingest.ThingID,
		Type:      ingest.Type,
		Timestamp: ingest.Timestamp,
		Value:     ingest.Payload,
		Metadata:  ingest.Metadata,
	}
	// Marshal
	outputMsg, err := json.Marshal(output)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal output message")
	}

	// return
	return outputMsg, nil
}
