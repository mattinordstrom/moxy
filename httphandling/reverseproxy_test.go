package httphandling_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/httphandling"
)

const waitTimeout = 5 * time.Second

func startMoxy(t *testing.T, upstream http.Handler) (*httptest.Server, *httptest.Server) {
	t.Helper()

	upstreamServer := httptest.NewServer(upstream)
	t.Cleanup(upstreamServer.Close)
	useUpstream(t, upstreamServer)

	moxy := httptest.NewServer(http.HandlerFunc(httphandling.RootHandler))
	t.Cleanup(moxy.Close)

	return moxy, upstreamServer
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyWebSocketUpgrade(t *testing.T) {
	upgrader := websocket.Upgrader{}
	moxy, _ := startMoxy(t, http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(resWriter, req, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msgType, msg, err := conn.ReadMessage()
		if err == nil {
			_ = conn.WriteMessage(msgType, append([]byte("echo: "), msg...))
		}
	}))

	conn, res, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(moxy.URL, "http")+"/hmr", nil)
	if err != nil {
		t.Fatalf("expected the upgrade to be proxied, got %v", err)
	}
	defer res.Body.Close()
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatal(err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(waitTimeout))

	_, msg, err := conn.ReadMessage()
	if err != nil || string(msg) != "echo: hello" {
		t.Errorf("expected the upstream echo, got %q %v", msg, err)
	}
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyStreamsEventsBeforeUpstreamFinishes(t *testing.T) {
	release := make(chan struct{})
	moxy, _ := startMoxy(t, http.HandlerFunc(func(resWriter http.ResponseWriter, _ *http.Request) {
		resWriter.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(resWriter, "data: first\n\n")
		_ = http.NewResponseController(resWriter).Flush()
		<-release
	}))
	// Registered after the servers so it runs before they wait for this handler
	t.Cleanup(func() { close(release) })

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, moxy.URL+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := moxy.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { res.Body.Close() })

	line := make(chan string, 1)

	go func() {
		text, _ := bufio.NewReader(res.Body).ReadString('\n')
		line <- text
	}()

	select {
	case got := <-line:
		if got != "data: first\n" {
			t.Errorf("expected the first event, got %q", got)
		}
	case <-time.After(waitTimeout):
		t.Error("expected the first event while the upstream is still open")
	}
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyCancelsUpstreamWhenClientLeaves(t *testing.T) {
	cancelled := make(chan struct{})
	moxy, _ := startMoxy(t, http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		<-req.Context().Done()
		close(cancelled)
	}))

	ctx, cancel := context.WithCancel(t.Context())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, moxy.URL+"/slow", nil)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	if res, err := moxy.Client().Do(req); err == nil {
		res.Body.Close()
	}

	select {
	case <-cancelled:
	case <-time.After(waitTimeout):
		t.Error("expected the upstream request to be cancelled")
	}
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyHeadersAndURLs(t *testing.T) {
	type seenReq struct {
		uri    string
		header http.Header
	}

	seen := make(chan seenReq, 1)
	moxy, upstream := startMoxy(t, http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		seen <- seenReq{uri: req.RequestURI, header: req.Header.Clone()}
	}))

	send := func(t *testing.T, uri string, headers map[string]string) seenReq {
		t.Helper()

		conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", strings.TrimPrefix(moxy.URL, "http://"))
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var raw strings.Builder

		raw.WriteString("GET " + uri + " HTTP/1.1\r\nHost: localhost\r\n")

		for name, value := range headers {
			raw.WriteString(name + ": " + value + "\r\n")
		}

		if _, err := io.WriteString(conn, raw.String()+"\r\n"); err != nil {
			t.Fatal(err)
		}

		select {
		case got := <-seen:
			return got
		case <-time.After(waitTimeout):
			t.Fatal("upstream got no request")
		}

		return seenReq{}
	}

	t.Run("hop-by-hop headers are dropped", func(t *testing.T) {
		got := send(t, "/api/x", map[string]string{"Connection": "X-Hop", "X-Hop": "secret", "X-Keep": "yes"})
		if got.header.Get("X-Hop") != "" || got.header.Get("X-Keep") != "yes" {
			t.Errorf("expected X-Hop dropped and X-Keep kept, got %v", got.header)
		}
	})

	t.Run("client X-Forwarded headers are passed on", func(t *testing.T) {
		got := send(t, "/api/x", map[string]string{"X-Forwarded-Host": "app.example", "X-Forwarded-For": "10.1.2.3"})
		if got.header.Get("X-Forwarded-Host") != "app.example" || got.header.Get("X-Forwarded-For") != "10.1.2.3" {
			t.Errorf("expected the client's X-Forwarded headers, got %v", got.header)
		}
	})

	t.Run("absolute-form URLs keep only the path", func(t *testing.T) {
		got := send(t, "http://localhost/api/abs?q=1", nil)
		if got.uri != "/api/abs?q=1" {
			t.Errorf("expected /api/abs?q=1 upstream, got %q (upstream %s)", got.uri, upstream.URL)
		}
	})
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyStreamOutlivesWriteTimeout(t *testing.T) {
	const writeTimeout = 200 * time.Millisecond

	upstream := httptest.NewServer(http.HandlerFunc(func(resWriter http.ResponseWriter, _ *http.Request) {
		resWriter.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(resWriter, "data: first\n\n")
		_ = http.NewResponseController(resWriter).Flush()
		time.Sleep(3 * writeTimeout)
		_, _ = io.WriteString(resWriter, "data: second\n\n")
	}))
	t.Cleanup(upstream.Close)
	useUpstream(t, upstream)

	moxy := httptest.NewUnstartedServer(http.HandlerFunc(httphandling.RootHandler))
	moxy.Config.WriteTimeout = writeTimeout
	moxy.Start()
	t.Cleanup(moxy.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, moxy.URL+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := moxy.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil || !strings.Contains(string(body), "data: second") {
		t.Errorf("expected the stream to outlive the write timeout, got %q %v", body, err)
	}
}

//nolint:paralleltest // swaps the global default route and forward transport
func TestProxyWebSocketOutlivesReadTimeout(t *testing.T) {
	const readTimeout = 200 * time.Millisecond

	upgrader := websocket.Upgrader{}
	upstream := httptest.NewServer(http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(resWriter, req, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msgType, msg, err := conn.ReadMessage()
		if err == nil {
			_ = conn.WriteMessage(msgType, msg)
		}
	}))
	t.Cleanup(upstream.Close)
	useUpstream(t, upstream)

	moxy := httptest.NewUnstartedServer(http.HandlerFunc(httphandling.RootHandler))
	moxy.Config.ReadTimeout = readTimeout
	moxy.Config.WriteTimeout = readTimeout
	moxy.Start()
	t.Cleanup(moxy.Close)

	conn, res, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(moxy.URL, "http")+"/hmr", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	defer conn.Close()

	time.Sleep(3 * readTimeout)

	if err := conn.WriteMessage(websocket.TextMessage, []byte("still there")); err != nil {
		t.Fatal(err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(waitTimeout))

	_, msg, err := conn.ReadMessage()
	if err != nil || string(msg) != "still there" {
		t.Errorf("expected the socket to outlive the read timeout, got %q %v", msg, err)
	}
}
