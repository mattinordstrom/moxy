package httphandling

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/utils"
)

var upgraderWSMock = websocket.Upgrader{
	CheckOrigin: wsMockOriginAllowed,
}

const (
	wsMockClientQueue  = 256
	wsMockWriteTimeout = 10 * time.Second
)

var (
	clients   = make(map[*wsMockClient]struct{})
	broadcast = make(chan []byte)
	lock      = sync.Mutex{}
)

// Each client has its own queue and writer, so one that stops reading can't stall the broadcast for the others.
type wsMockClient struct {
	conn      *websocket.Conn
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func (client *wsMockClient) close() {
	client.closeOnce.Do(func() { close(client.done) })
}

func (client *wsMockClient) writePump() {
	defer client.conn.Close()

	for {
		select {
		case <-client.done:
			return
		case msg := <-client.send:
			err := client.conn.SetWriteDeadline(time.Now().Add(wsMockWriteTimeout))
			if err == nil {
				err = client.conn.WriteMessage(websocket.TextMessage, msg)
			}

			if err != nil {
				client.close()

				return
			}
		}
	}
}

func handleWebSocketWSMock(w http.ResponseWriter, r *http.Request) {
	conn, err := upgraderWSMock.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)

		return
	}

	client := &wsMockClient{conn: conn, send: make(chan []byte, wsMockClientQueue), done: make(chan struct{})}

	lock.Lock()
	clients[client] = struct{}{}
	lock.Unlock()

	go client.writePump()

	defer func() {
		lock.Lock()
		delete(clients, client)
		lock.Unlock()

		client.close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		broadcast <- msg
	}
}

func handleWSMockMessages() {
	for {
		msg := <-broadcast

		fmt.Print(utils.ColorYellow + utils.SanitizeTerminal(string(msg)) + utils.ColorReset + "\n")
		updateAdminWithLatest(string(msg), utils.EventTypeWSMock, map[string]any{})

		lock.Lock()
		recordEvent(LogEntry{Type: utils.EventTypeWSMock, Message: fmt.Sprintf("%d bytes broadcast to %d clients", len(msg), len(clients))})

		for client := range clients {
			select {
			case client.send <- msg:
			default:
				log.Println("Dropping a WS mock client that stopped reading")
				delete(clients, client)
				client.close()
			}
		}
		lock.Unlock()
	}
}
