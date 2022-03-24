package main

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"

	"github.com/digicatapult/wasp-video-stream-service/util"
	"github.com/digicatapult/wasp-video-stream-service/wasp"
)

func main() {
	inTopicName := util.GetEnv(util.InTopicNameKey, "video")
	kafkaBrokersVar := util.GetEnv(util.KafkaBrokersKey, "localhost:9092")
	kafkaBrokers := strings.Split(kafkaBrokersVar, ",")

	cfg := zap.NewDevelopmentConfig()
	if util.GetEnv(util.EnvKey, "") == "production" {
		cfg = zap.NewProductionConfig()

		lvl, err := zap.ParseAtomicLevel(util.GetEnv(util.LogLevelKey, "debug"))
		if err != nil {
			panic("invalid log level")
		}

		log.Printf("setting level: %s", lvl.String())

		cfg.Level = lvl
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("error initializing the logger")
	}

	defer func() {
		syncErr := logger.Sync()
		if err != nil {
			log.Printf("error whilst syncing zap: %s\n", syncErr)
		}
	}()

	zap.ReplaceGlobals(logger)

	sarama.Logger = util.SaramaZapLogger{}

	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(inTopicName, 0, 0)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumed := 0
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed message offset %d\n", msg.Offset)

			msgUnmarshalled, err := wasp.VideoMessage(msg)
			if err != nil {
				zap.S().Fatalf("msgUnmarshalled error %s", err)
			}
			log.Printf("Consumed message payload %s\n", msgUnmarshalled)

			consumed++
		case <-signals:
			break ConsumerLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
}
