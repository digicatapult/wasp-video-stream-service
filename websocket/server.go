// Package websocket implements a websocket reader/writer. The
// implementation heavily based on the VW Video Streaming server
// (https://github.com/timdrysdale/vw).
//
// The code was originally taken on 2022-03-24 from the below address
// https://github.com/timdrysdale/vw/blob/6dbe6d197ddaf1578edbfb969a73c7ec45d29e98/cmd/handleWs.go
package websocket

import (
	"net/http"
	"sync"
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
	clients    map[string]*WsHandlerClient
	clientLock *sync.RWMutex
	msgChan    chan []byte
}

// NewController will initialise a new instance
func NewController(msgChan chan []byte) *Controller {
	c := &Controller{
		clients:    map[string]*WsHandlerClient{},
		clientLock: &sync.RWMutex{},
		msgChan:    msgChan,
	}

	go c.ForwardMessages()

	return c
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

	client.Conn.SetCloseHandler(func(code int, text string) error {
		c.clientLock.Lock()
		delete(c.clients, client.ID)
		c.clientLock.Unlock()

		zap.S().Infof("client %s disconnected", client.ID)

		close(client.Messages)
		return nil
	})

	c.clientLock.Lock()
	c.clients[client.ID] = client
	c.clientLock.Unlock()
	zap.S().Infof("client added: %s", client.ID)
	zap.S().Debugf("current clients: %#v", c.clients)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// ForwardMessages will iterate messages and send them  to connected clients
func (c *Controller) ForwardMessages() {
	for msg := range c.msgChan {
		c.clientLock.RLock()
		for _, client := range c.clients {
			client.Messages <- msg
		}
		c.clientLock.RUnlock()
	}
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
		// close channels when we leave this function
		// (in the event of an error or connection closing)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		zap.S().Errorf("problem setting read deadline: %s", err)
		return
	}

	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			zap.S().Warnf("problem setting read deadline in pong handler: %s", err)
			return err
		}
		return nil
	})

	for {
		// TODO Decide on reader implementation
		mt, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.S().Errorf("error: %v", err)
			}
			break
		}
		zap.S().With("type", mt, "data", data).Debug("received a client message that are unhandled")
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WsHandlerClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		// TODO Test channel closing behaviour
		case chunk, ok := <-c.Messages:
			// TODO investigate write deadlines
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				zap.S().Errorf("problem setting write deadline; exiting writer: %s", err)
				return
			}

			if !ok {
				// The hub closed the channel.
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					zap.S().Warnf("problem write close message: %s", err)
					return
				}
				zap.S().Debugf("message channel closed, exiting write goroutine")
				return
			}

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				zap.S().Warnf("problem obtaining next writer; exiting write goroutine: %s", err)
				return
			}

			if _, err := w.Write(chunk); err != nil {
				zap.S().Warnf("problem writing chunk: %s", err)
			}
			c.ChunkCount++

			size := len(chunk)

			// Add queued chunks to the current websocket message, without delimiter.
			n := len(c.Messages)
			for i := 0; i < n; i++ {
				zap.S().Debug("writing follow up")
				followOnMessage := <-c.Messages
				if _, err := w.Write(followOnMessage); err != nil {
					zap.S().Warnf("problem writing followup chunk: %s", err)
				}
				size += len(followOnMessage)
				c.ChunkCount++
			}

			if err := w.Close(); err != nil {
				zap.S().Warnf("problem closing writer; exiting write goroutine: %s", err)
				return
			}
			if c.ChunkCount%100 == 0 {
				zap.S().Debugf("chunks sent to client '%s': %d", c.ID, c.ChunkCount)
			}
		case <-ticker.C:
			// TODO investigate write deadlines
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				zap.S().Warnf("problem setting write deadline after timeout: %s", err)
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				zap.S().Debugf("problem sending ping after reaching WriteDeadline: %s", err)
				return
			}
		}
	}
}
