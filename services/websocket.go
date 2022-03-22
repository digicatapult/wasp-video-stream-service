package services

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
)

func ChunksToWebsocket(rtmpURL string) {
	pipeReader, pipeWriter := io.Pipe()
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)
	videoChunks := make(chan []byte, 10000)
	chunkCount = map[string]int{}

	go sendVideoChunks(videoChunks)
	go readByteBufferToChannel(pipeReader, 1000, videoChunks, videoSendWaitGroup)

	done := make(chan error)

	go func() {
		err := ffmpeg.Input(rtmpURL).
			Output("pipe:", ffmpeg.KwArgs{"f": "mpegts"}).
			WithOutput(pipeWriter).
			Run()
		if err != nil {
			zap.S().Fatalf("problem with ffmpeg: %v", err)
		}
		done <- err
	}()

	err := <-done
	zap.S().Infof("Done (waiting for completion of send): %s", err)
	// videoSendWaitGroup.Wait()
	shutdown <- true
}

var (
	// videoChunks chan []byte
	clients []*WsHandlerClient
	close   chan bool
)

func sendVideoChunks(videoChunks chan []byte) {
	chunkCount := 0
	for {
		select {
		case chunk := <-videoChunks:
			if chunkCount%100 == 0 {
				zap.S().Infof("chunks forwarded: %d", chunkCount)
			}
			for _, c := range clients {
				c.Messages <- chunk
			}
			chunkCount++
		}
	}
}

func readByteBufferToChannel(reader io.Reader, bufferSize int, bufferChannel chan []byte, waitGroup *sync.WaitGroup) {
	frameSize := 1000
	frameCount := 0
	buf := make([]byte, frameSize)

	for {
		if frameCount%100 == 0 {
			zap.S().Infof("frames consumed: %d", frameCount)
		}
		count, err := io.ReadFull(reader, buf)
		frameCount++

		switch {
		case count == 0 || errors.Is(err, io.EOF):
			zap.S().Debug("end of stream reached")

			return
		case count != frameSize:
			zap.S().Infof("end of stream reached, sending short chunk: %d, %s", count, err)
		case err != nil:
			zap.S().Errorf("read error: %d, %s", count, err)

			if count == 0 {
				return
			}
		}

		bufCopy := make([]byte, frameSize)
		copy(bufCopy, buf)

		waitGroup.Add(1)
		bufferChannel <- bufCopy

	}
}

// 4096 Bytes is the approx average message size
// this number does not limit message size
// So for key frames we just make a few more syscalls
// null subprotocol required by Chrome
// TODO restrict CheckOrigin
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	Subprotocols:    []string{"null"},
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().With("error", err).Error("Failed upgrading to websocket connection in wsHandler")
		return
	}

	// vars := mux.Vars(r)
	// topic := vars["feed"]

	// messageClient := &hub.Client{
	// 	Hub:   app.Hub.Hub,
	// 	Name:  uuid.New().String()[:3],
	// 	Send:  make(chan hub.Message),
	// 	Stats: hub.NewClientStats(),
	// 	Topic: topic,
	// }

	client := &WsHandlerClient{
		ID:         uuid.NewV4().String(),
		Messages:   make(chan []byte),
		Conn:       conn,
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.Header.Get("X-Forwarded-For"),
	}

	clients = append(clients, client)
	chunkCount[client.ID] = 0

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.readPump()
	go client.writePump(close)
	zap.S().Info("client added")
}

type WsHandlerClient struct {
	Messages   chan []byte
	Conn       *websocket.Conn
	UserAgent  string // r.UserAgent()
	RemoteAddr string // r.Header.Get("X-Forwarded-For")
	ID         string
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (10MB)
	// Typical key frame at 640x480 is 60 * 188B ~= 11kB
	maxMessageSize = 1024 * 1024 * 10
)

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WsHandlerClient) readPump() {
	// defer func() {
	// 	c.Messages.Hub.Unregister <- c.Messages
	// 	c.Conn.Close()
	// }()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		mt, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.S().Errorf("error: %v", err)
			}
			break
		}

		// t := time.Now()

		zap.S().With("type", mt, "data", data).Info("client message")
		zap.S().Infof("%s", data)
		// c.Messages.Hub.Broadcast <- hub.Message{Sender: *c.Messages, Data: data, Type: mt, Sent: t}
	}
}

var chunkCount map[string]int

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WsHandlerClient) writePump(closed <-chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case chunk := <-c.Messages:
			// c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			// if !ok {
			// 	// The hub closed the channel.
			// 	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			// 	return
			// }
			// zap.S().Info("sending stuff")
			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			w.Write(chunk)
			chunkCount[c.ID]++

			size := len(chunk)

			// Add queued chunks to the current websocket message, without delimiter.
			n := len(c.Messages)
			for i := 0; i < n; i++ {
				zap.S().Info("writing follow up")
				followOnMessage := <-c.Messages
				w.Write(followOnMessage)
				size += len(followOnMessage)
				chunkCount[c.ID]++
			}

			if err := w.Close(); err != nil {
				return
			}
			if chunkCount[c.ID]%100 == 0 {
				zap.S().Infof("chunks sent to client '%s': %d", c.ID, chunkCount[c.ID])
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-closed:
			return
		}
	}
}
