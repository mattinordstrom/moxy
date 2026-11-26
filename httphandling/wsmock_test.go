package httphandling_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/httphandling"
)

func dialWSMock(t *testing.T, moxy *httptest.Server) *websocket.Conn {
	t.Helper()

	conn, res, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(moxy.URL, "http")+"/moxywsmock", nil)
	if res != nil {
		res.Body.Close()
	}

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { conn.Close() })

	return conn
}

func drain(conn *websocket.Conn) {
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

//nolint:paralleltest // uses the global WS mock clients
func TestWSMockStuckClientDoesNotStallOthers(t *testing.T) {
	httphandling.StartWSMockBroadcaster()

	mux := http.NewServeMux()
	mux.HandleFunc("/moxywsmock", httphandling.HandleWSMock)

	moxy := httptest.NewServer(mux)
	t.Cleanup(moxy.Close)

	dialWSMock(t, moxy) // never reads
	reader := dialWSMock(t, moxy)
	sender := dialWSMock(t, moxy)
	drain(sender)

	received := make(chan string, 1)

	go func() {
		for {
			_, msg, err := reader.ReadMessage()
			if err != nil {
				return
			}

			if bytes.HasPrefix(msg, []byte("last")) {
				received <- string(msg)

				return
			}
		}
	}()

	// Enough to fill the stuck client's socket buffers and its 256 message queue
	filler := bytes.Repeat([]byte("x"), 128<<10)
	for range 400 {
		if err := sender.WriteMessage(websocket.TextMessage, filler); err != nil {
			t.Fatal(err)
		}
	}

	if err := sender.WriteMessage(websocket.TextMessage, []byte("last")); err != nil {
		t.Fatal(err)
	}

	select {
	case <-received:
	case <-time.After(10 * time.Second):
		t.Fatal("reading client stopped getting messages while another client was stuck")
	}
}
