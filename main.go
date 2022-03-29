package main

import (
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/digicatapult/wasp-video-stream-service/util"
	"github.com/digicatapult/wasp-video-stream-service/wasp"
)

func main() {
	inTopicName := util.GetEnv(util.InTopicNameKey, "video")
	kafkaBrokersVar := util.GetEnv(util.KafkaBrokersKey, "localhost:9092")
	kafkaBrokers := strings.Split(kafkaBrokersVar, ",")
	hostAddress := util.GetEnv(util.HostAddressKey, "localhost:9999")

	util.ConfigureLogging()

	sarama.Logger = util.SaramaZapLogger{}

	srv := &http.Server{Addr: hostAddress}
	router := mux.NewRouter()

	msgChan := make(chan *wasp.Message)

	router.HandleFunc("/{streamID}", streamHandler).Methods("GET")
	router.HandleFunc("/{streamID}/{segName:output[0-9]+.ts}", streamHandler).Methods("GET")

	// TODO define best size for this
	// msgChan := make(chan []byte, 100000)
	// wsController := websocket.NewController(msgChan)
	// zap.S().Infof("Starting server on: '%s'", hostAddress)
	// srv := &http.Server{Addr: hostAddress}
	// router := mux.NewRouter()
	// router.HandleFunc(`/ws`, wsController.HandleWs)

	srv.Handler = router
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.S().With("error", err).Fatal("http.ListenAndServe")
		}
		zap.S().Info("Exiting http.Server")
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop
		zap.S().Info("closing down")
		os.Exit(0)
	}()

	go consumeVideoMessages(kafkaBrokers, inTopicName, msgChan)
	storeVideoSegments(msgChan)
}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	// streamID, err := strconv.Atoi(vars["streamID"])
	// if err != nil {
	// 	response.WriteHeader(http.StatusNotFound)
	// 	return
	// }

	segName, ok := vars["segName"]
	// zap.S().Info(segName)
	if !ok {
		writePlaylist(response, "http://localhost:9999/stream_test")
		return
	}

	seg, ok := mpegTsStore[segName]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	response.Write(seg)
}

func writePlaylist(w http.ResponseWriter, base string) {
	for _, l := range m3u8Playlist {

		match, err := regexp.Match("output[0-9]+.ts", []byte(l))
		if err != nil {
			zap.S().Errorf("problem checking playlist lines: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if match {
			l = path.Join(base, l)
		}

		w.Write([]byte(l))
		w.Write([]byte("\n"))
	}
}

var (
	m3u8Playlist = []string{}
	mpegTsStore  = map[string][]byte{}
)

func storeVideoSegments(msgChan chan *wasp.Message) {
	for msg := range msgChan {
		p := msg.Payload

		switch p.Type {
		case "meta":
			if len(m3u8Playlist) > 0 {
				newlines := getLastLines(msg.Payload.Data)
				// zap.S().Infof("adding new lines to playlist: '%s'", newlines)
				// TODO: check file exists first
				m3u8Playlist = append(m3u8Playlist, newlines...)
				continue
			}
			m3u8Playlist = strings.Split(string(msg.Payload.Data), "\n")
		case "data":
			mpegTsStore[p.Filename] = p.Data
		default:
			zap.S().Warnf("unknown message payload type '%s'", p.Type)
		}
	}
}

func getLastLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	return lines[len(lines)-3 : len(lines)-1]
}

func consumeVideoMessages(kafkaBrokers []string, inTopicName string, msgChan chan *wasp.Message) {
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
			// zap.S().Debugf("Consumed message offset %d", msg.Offset)

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
