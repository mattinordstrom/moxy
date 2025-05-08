package httphandling

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/utils"
)

var upgraderWSMock = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte)
	lock      = sync.Mutex{}
)

func handleWebSocketWSMock(w http.ResponseWriter, r *http.Request) {
	conn, err := upgraderWSMock.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)

		return
	}

	defer conn.Close()

	lock.Lock()
	clients[conn] = true
	lock.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			lock.Lock()
			delete(clients, conn)
			lock.Unlock()

			break
		}
		broadcast <- msg
	}
}

func handleWSMockMessages() {
	for {
		msg := <-broadcast

		fmt.Printf("WS Mock received msg: %s\n", msg)
		updateAdminWithLatest(string(msg), utils.EventTypeWSMock, map[string]interface{}{})

		lock.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		lock.Unlock()
	}
}
