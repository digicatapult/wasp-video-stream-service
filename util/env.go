package util

import "os"

const (
	// InTopicNameKey defines the environment variable key for InTopicName
	InTopicNameKey = "IN_TOPIC_NAME_KEY"

	// OutTopicNameKey defines the environment variable key for OutTopicName
	OutTopicNameKey = "OUT_TOPIC_NAME_KEY"

	// KafkaBrokersKey defines the environment variable key for KafkaBrokers
	KafkaBrokersKey = "KAFKA_BROKERS"

	// EnvKey defines the environment variable key for Env
	EnvKey = "ENV"

	// LogLevelKey defines the environment variable key for LogLevel
	LogLevelKey = "LOG_LEVEL"
)

// GetEnv will lookup a environment variable or return the default
func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}