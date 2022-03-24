// Package websocket implements a websocket reader/writer. The
// implementation heavily based on the VW Video Streaming server
// (https://github.com/timdrysdale/vw).
//
// The code was originally taken on 2022-03-24 from the below address
// https://github.com/timdrysdale/vw/blob/6dbe6d197ddaf1578edbfb969a73c7ec45d29e98/cmd/handleWs.go
package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

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

// Controller defines a websocket controller for managing websocket connections
type Controller struct {
	clients []*WsHandlerClient
}

// HandleWs handles a websocket connection and records a client instance
func (c *Controller) HandleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().With("error", err).Error("Failed upgrading to websocket connection in wsHandler")
		return
	}

	client := &WsHandlerClient{
		ID:         uuid.NewV4().String(),
		Messages:   make(chan []byte),
		Conn:       conn,
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.Header.Get("X-Forwarded-For"),
		ChunkCount: 0,
	}

	c.clients = append(c.clients, client)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	// TODO add close channel
	// go client.writePump(app.closed)
	go client.writePump(nil)
	go client.readPump()
}

// WsHandlerClient defines a client connection and message transfer
type WsHandlerClient struct {
	Messages   chan []byte
	Conn       *websocket.Conn
	UserAgent  string // r.UserAgent()
	RemoteAddr string // r.Header.Get("X-Forwarded-For")
	ID         string
	ChunkCount int
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WsHandlerClient) readPump() {
	defer func() {
		// TODO close message channel
		// c.Messages.Hub.Unregister <- c.Messages
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// TODO Decide on reader implementation
		//
		// mt, data, err := c.Conn.ReadMessage()
		// if err != nil {
		// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		// 		log.Errorf("error: %v", err)
		// 	}
		// 	break
		// }

		// t := time.Now()

		// c.Messages.Hub.Broadcast <- hub.Message{Sender: *c.Messages, Data: data, Type: mt, Sent: t}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WsHandlerClient) writePump(closed <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case chunk := <-c.Messages:
			// TODO investigate write deadlines
			// c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			// if !ok {
			// 	// The hub closed the channel.
			// 	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			// 	return
			// }

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			w.Write(chunk)
			c.ChunkCount++

			size := len(chunk)

			// Add queued chunks to the current websocket message, without delimiter.
			n := len(c.Messages)
			for i := 0; i < n; i++ {
				zap.S().Debug("writing follow up")
				followOnMessage := <-c.Messages
				w.Write(followOnMessage)
				size += len(followOnMessage)
				c.ChunkCount++
			}

			if err := w.Close(); err != nil {
				return
			}
			if c.ChunkCount%100 == 0 {
				zap.S().Infof("chunks sent to client '%s': %d", c.ID, c.ChunkCount)
			}
		case <-ticker.C:
			// TODO investigate write deadlines
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-closed:
			return
		}
	}
}
