package httphandling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/config"
	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
	"github.com/mattinordstrom/moxy/version"
)

type WebSocketMessage struct {
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Extras    map[string]any `json:"extras,omitempty"`
	Timestamp string         `json:"timestamp"`
}

type Settings struct {
	Port          int      `json:"port"`
	DefaultRoute  string   `json:"defaultRoute"`
	Insecure      bool     `json:"defaultInsecure"`
	PayloadFiles  []string `json:"payloadFiles"`
	PayloadPath   string   `json:"payloadPath"`
	MaxLogEntries *int     `json:"maxLogEntries"`
	Version       string   `json:"version"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: adminWebSocketAllowed,
}

var (
	currentWSConnection  *WebSocketConnection
	connMutex            sync.Mutex
	closeMessageDeadLine = 5
	wsSendBuffer         = 1024
)

type WebSocketConnection struct {
	// Only writePump writes to Conn, since gorilla/websocket allows a single concurrent writer
	Conn        *websocket.Conn
	send        chan []byte
	done        chan struct{}
	closeOnce   sync.Once
	closeReason string
}

func NewWebSocketConnection(conn *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		Conn: conn,
		send: make(chan []byte, wsSendBuffer),
		done: make(chan struct{}),
	}
}

func (wc *WebSocketConnection) enqueueMessage(message []byte) {
	select {
	case <-wc.done:
	case wc.send <- message:
	default:
		log.Println("Send channel is full, dropping message")
	}
}

func (wc *WebSocketConnection) close(reason string) {
	wc.closeOnce.Do(func() {
		wc.closeReason = reason
		close(wc.done)
	})
}

func (wc *WebSocketConnection) writePump() {
	defer wc.Conn.Close()

	for {
		select {
		case <-wc.done:
			if wc.closeReason != "" {
				writeCloseMessage(wc.Conn, wc.closeReason)
			}

			return
		case message := <-wc.send:
			if err := wc.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				wc.close("")
				forgetWSConnection(wc)

				return
			}
		}
	}
}

// Reading is what notices a closed tab, and it also handles the client's ping and close frames.
func (wc *WebSocketConnection) readPump() {
	for {
		if _, _, err := wc.Conn.ReadMessage(); err != nil {
			wc.close("")
			forgetWSConnection(wc)

			return
		}
	}
}

func forgetWSConnection(wc *WebSocketConnection) {
	connMutex.Lock()
	if currentWSConnection == wc {
		currentWSConnection = nil
	}
	connMutex.Unlock()
}

func handleWebSocket(resWriter http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(resWriter, req, nil)
	if err != nil {
		utils.LogError("WS connection error: ", err)

		return
	}

	wsConn := NewWebSocketConnection(conn)

	connMutex.Lock()
	oldConn := currentWSConnection
	currentWSConnection = wsConn
	connMutex.Unlock()

	if oldConn != nil {
		oldConn.close("ws_takeover")
	}

	go wsConn.writePump()
	go wsConn.readPump()
}

func handleAdminReq(resWriter http.ResponseWriter, req *http.Request) {
	if !guardAdminRequest(resWriter, req) {
		return
	}

	reqURL := req.URL.Path

	switch reqURL {
	case "/moxyadminui/mockdef", "/moxyadminui/proxydef":
		jsonName := "mockdef"
		if reqURL == "/moxyadminui/proxydef" {
			jsonName = "proxydef"
		}

		if req.Method == http.MethodPost {
			jsonData, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}
			defer req.Body.Close()

			var data any
			if reqURL == "/moxyadminui/mockdef" && !config.AppConfig.Admin.AutoSortJSONKeys {
				data = &[]models.Mock{}
			}

			err = unmarshalKeepNumbers(jsonData, &data)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusBadRequest)

				return
			}

			jsonData, err = json.MarshalIndent(data, "", "  ")
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}

			const filePermission = 0o600

			errr := utils.WriteFileAtomic(jsonName+".json", jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			http.ServeFile(resWriter, req, jsonName+".json")
		}
	case "/moxyadminui/uisettings":
		if req.Method == http.MethodPost {
			jsonData, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}
			defer req.Body.Close()

			var data map[string]any

			err = unmarshalKeepNumbers(jsonData, &data)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusBadRequest)

				return
			}

			jsonData, err = json.MarshalIndent(data, "", "  ")
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}

			const filePermission = 0o600

			errr := utils.WriteFileAtomic("ui/settings.json", jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			http.ServeFile(resWriter, req, "ui/settings.json")
		}
	case "/moxyadminui/log":
		handleEventLogReq(resWriter, req)
	case "/moxyadminui/settings":
		payloadFilesInDir, err := os.ReadDir(config.AppConfig.Defaults.PayloadArchivePath)
		if err != nil {
			utils.LogError("Error reading Payload Archive Path:", err)

			http.Error(resWriter, err.Error(), http.StatusInternalServerError)

			return
		}

		var payloadFiles []string
		for _, entry := range payloadFilesInDir {
			payloadFiles = append(payloadFiles, entry.Name())
		}

		data := Settings{
			Port:          Port,
			DefaultRoute:  DefaultRoute,
			Insecure:      DefaultInsecure,
			PayloadFiles:  payloadFiles,
			PayloadPath:   config.AppConfig.Defaults.PayloadArchivePath,
			MaxLogEntries: config.AppConfig.Admin.MaxLogEntries,
			Version:       version.Version,
		}

		jsonResponse, err := json.Marshal(data)
		if err != nil {
			http.Error(resWriter, err.Error(), http.StatusInternalServerError)

			return
		}

		resWriter.Header().Set("Content-Type", "application/json")

		_, err = resWriter.Write(jsonResponse)
		if err != nil {
			// Since headers are already sent, we can't send a new HTTP status code.
			// Log the error instead.
			utils.LogError("Error writing response: ", err)
		}
	case "/moxyadminui":
		http.ServeFile(resWriter, req, "ui/index.html")
	case "/moxyadminui/editpayloadfile":
		if req.Method == http.MethodPost {
			jsonData, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}
			defer req.Body.Close()

			type RequestData struct {
				File    string `json:"file"`
				Payload any    `json:"payload"`
			}

			type RequestDataNoAutoSort struct {
				File    string          `json:"file"`
				Payload json.RawMessage `json:"payload"`
			}

			var data any
			if !config.AppConfig.Admin.AutoSortJSONKeys {
				data = &RequestDataNoAutoSort{}
			} else {
				data = &RequestData{}
			}

			err = unmarshalKeepNumbers(jsonData, &data)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusBadRequest)

				return
			}

			var payload any

			var filePath string

			if !config.AppConfig.Admin.AutoSortJSONKeys {
				reqData, ok := data.(*RequestDataNoAutoSort)
				if !ok {
					http.Error(resWriter, "Invalid data type", http.StatusBadRequest)

					return
				}

				payload = reqData.Payload
				filePath = reqData.File
			} else {
				reqData, ok := data.(*RequestData)
				if !ok {
					http.Error(resWriter, "Invalid data type", http.StatusBadRequest)

					return
				}

				payload = reqData.Payload
				filePath = reqData.File
			}

			absPath, err := filepath.Abs(filePath)
			if err != nil {
				http.Error(resWriter, "Invalid file path", http.StatusBadRequest)

				return
			}

			root, relPath, err := resolveInArchive(absPath)
			if err != nil {
				writeArchivePathError(resWriter, err)

				return
			}
			defer root.Close()

			jsonData, err = json.MarshalIndent(payload, "", "  ")
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}

			const filePermission = 0o600
			errr := root.WriteFile(relPath, jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}

			utils.ForgetCache()
		} else {
			queryParams := req.URL.Query()

			value, exists := queryParams["file"]
			if !exists {
				log.Println("Parameter 'file' not found")
				http.Error(resWriter, "Parameter 'file' not found", http.StatusInternalServerError)

				return
			}

			root, relPath, err := resolveInArchive(value[0])
			if err != nil {
				writeArchivePathError(resWriter, err)

				return
			}
			defer root.Close()

			http.ServeFileFS(resWriter, req, root.FS(), filepath.ToSlash(relPath)) //nolint:gosec // root.FS() confines access to the archive
			// log.Println(req.Method + " " + fmt.Sprintf(value[0]))
		}
	default:
		return
	}
}

var errTrailingJSON = errors.New("unexpected data after the JSON value")

// UseNumber keeps big ints and number formatting intact when the keys are re-sorted through a map.
func unmarshalKeepNumbers(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	err := decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	_, err = decoder.Token()
	if !errors.Is(err, io.EOF) {
		return errTrailingJSON
	}

	return nil
}

var errOutsideArchive = errors.New("path is outside the payload archive")

func resolveInArchive(path string) (*os.Root, string, error) {
	basePath, err := filepath.Abs(config.AppConfig.Defaults.PayloadArchivePath)
	if err != nil {
		return nil, "", fmt.Errorf("resolve archive path: %w", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(basePath, path)
	}

	relPath, err := filepath.Rel(basePath, path)
	if err != nil || !filepath.IsLocal(relPath) {
		return nil, "", errOutsideArchive
	}

	// os.Root also stops symlinks inside the archive from escaping it
	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, "", fmt.Errorf("open archive: %w", err)
	}

	return root, relPath, nil
}

func writeArchivePathError(resWriter http.ResponseWriter, err error) {
	if errors.Is(err, errOutsideArchive) {
		http.Error(resWriter, "Invalid file path", http.StatusBadRequest)

		return
	}

	http.Error(resWriter, "Invalid archive path", http.StatusInternalServerError)
}

func writeCloseMessage(conn *websocket.Conn, reason string) {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason)
	deadline := time.Now().Add(time.Second * time.Duration(closeMessageDeadLine))

	err := conn.WriteControl(websocket.CloseMessage, msg, deadline)
	if err != nil {
		utils.LogError("Error WriteMessage close msg: ", err)
	}
}

func updateAdminWithLatest(evtStr string, evtType string, extras map[string]any) {
	connMutex.Lock()
	wsConn := currentWSConnection
	connMutex.Unlock()

	if wsConn != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		msg := WebSocketMessage{
			Type:      evtType,
			Message:   evtStr,
			Extras:    extras,
			Timestamp: timestamp,
		}

		jsonMsg, jErr := json.Marshal(msg)
		if jErr != nil {
			utils.LogError("Error: json marshal websocket msg ", jErr)
		}

		wsConn.enqueueMessage(jsonMsg)
	}
}
