package main

import (
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/digicatapult/wasp-video-stream-service/util"
	"github.com/digicatapult/wasp-video-stream-service/wasp"
	"github.com/digicatapult/wasp-video-stream-service/websocket"
)

func main() {
	inTopicName := util.GetEnv(util.InTopicNameKey, "video")
	kafkaBrokersVar := util.GetEnv(util.KafkaBrokersKey, "localhost:9092")
	kafkaBrokers := strings.Split(kafkaBrokersVar, ",")
	hostAddress := util.GetEnv(util.HostAddressKey, "localhost:9999")

	util.ConfigureLogging()

	sarama.Logger = util.SaramaZapLogger{}

	// TODO define best size for this
	msgChan := make(chan []byte, 100000)

	wsController := websocket.NewController(msgChan)

	zap.S().Infof("Starting server on: '%s'", hostAddress)
	srv := &http.Server{Addr: hostAddress}

	router := mux.NewRouter()
	router.HandleFunc(`/ws`, wsController.HandleWs)

	srv.Handler = router
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.S().With("error", err).Fatal("http.ListenAndServe")
		}
		zap.S().Info("Exiting http.Server")
	}()

	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			zap.S().Fatalf("consumer error %s", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(inTopicName, 0, 0)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			zap.S().Fatalf("msgUnmarshalled error %s", err)
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
			zap.S().Debugf("Consumed message offset %d", msg.Offset)

			msgUnmarshalled, err := wasp.VideoMessage(msg)
			if err != nil {
				zap.S().Fatalf("msgUnmarshalled error %s", err)
			}

			msgChan <- msgUnmarshalled

			consumed++
		case <-signals:
			break ConsumerLoop
		}
	}

	zap.S().Debugf("Consumed: %d", consumed)
}
