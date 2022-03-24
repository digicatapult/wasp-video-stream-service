package util

import "go.uber.org/zap"

// SaramaZapLogger is a zap logger for sarama
type SaramaZapLogger struct{}

// Print will Print a log message to zap info
func (sz SaramaZapLogger) Print(v ...interface{}) {
	zap.S().Info(v)
}

// Printf will Printf a log message to zap info
func (sz SaramaZapLogger) Printf(format string, v ...interface{}) {
	zap.S().Infof(format, v)
}

// Println will Println a log message to zap info
func (sz SaramaZapLogger) Println(v ...interface{}) {
	zap.S().Info(v)
}
