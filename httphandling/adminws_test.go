package httphandling_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/httphandling"
)

func dialAdmin(t *testing.T, moxy *httptest.Server, header http.Header) (*websocket.Conn, error) {
	t.Helper()

	conn, res, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(moxy.URL, "http")+"/moxyws", header)
	if res != nil {
		res.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("dial admin websocket: %w", err)
	}

	t.Cleanup(func() { conn.Close() })

	return conn, nil
}

func startAdmin(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/moxyws", httphandling.HandleAdminWebSocket)

	moxy := httptest.NewServer(mux)
	t.Cleanup(moxy.Close)

	return moxy
}

func readAdminMessage(t *testing.T, conn *websocket.Conn) (string, error) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}

	_, msg, err := conn.ReadMessage()

	return string(msg), err
}

func waitForAdminConnection(t *testing.T, want bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for httphandling.HasAdminConnection() != want {
		if time.Now().After(deadline) {
			t.Fatalf("admin connection present = %v, want %v", !want, want)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

//nolint:paralleltest // uses the global admin connection
func TestAdminWebSocketTakeoverWhileSending(t *testing.T) {
	moxy := startAdmin(t)

	first, err := dialAdmin(t, moxy, nil)
	if err != nil {
		t.Fatal(err)
	}

	waitForAdminConnection(t, true)

	var senders sync.WaitGroup
	stop := make(chan struct{})

	for range 4 {
		senders.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
					httphandling.UpdateAdminWithLatest("event", "proxy", nil)
				}
			}
		})
	}

	second, err := dialAdmin(t, moxy, nil)
	if err != nil {
		t.Fatal(err)
	}

	for {
		_, err = readAdminMessage(t, first)
		if err != nil {
			break
		}
	}

	close(stop)
	senders.Wait()

	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) || closeErr.Text != "ws_takeover" {
		t.Fatalf("first tab got %v, want close with ws_takeover", err)
	}

	httphandling.UpdateAdminWithLatest("after takeover", "proxy", nil)

	for {
		msg, err := readAdminMessage(t, second)
		if err != nil {
			t.Fatalf("second tab: %v", err)
		}

		if strings.Contains(msg, "after takeover") {
			break
		}
	}
}

//nolint:paralleltest // uses the global admin connection
func TestAdminWebSocketBlockedOriginKeepsCurrentTab(t *testing.T) {
	moxy := startAdmin(t)

	current, err := dialAdmin(t, moxy, nil)
	if err != nil {
		t.Fatal(err)
	}

	waitForAdminConnection(t, true)

	if _, err := dialAdmin(t, moxy, http.Header{"Origin": {"http://evil.example"}}); err == nil {
		t.Fatal("cross-origin admin websocket was accepted")
	}

	httphandling.UpdateAdminWithLatest("still connected", "proxy", nil)

	msg, err := readAdminMessage(t, current)
	if err != nil || !strings.Contains(msg, "still connected") {
		t.Fatalf("current tab got %q, %v", msg, err)
	}
}

//nolint:paralleltest // uses the global admin connection
func TestAdminWebSocketForgetsClosedTab(t *testing.T) {
	moxy := startAdmin(t)

	conn, err := dialAdmin(t, moxy, nil)
	if err != nil {
		t.Fatal(err)
	}

	waitForAdminConnection(t, true)

	conn.Close()

	waitForAdminConnection(t, false)
}
