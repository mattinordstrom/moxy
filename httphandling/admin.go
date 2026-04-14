package httphandling

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/config"
	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

type WebSocketMessage struct {
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Extras    map[string]any `json:"extras,omitempty"`
	Timestamp string         `json:"timestamp"`
}

type Settings struct {
	Port         int      `json:"port"`
	DefaultRoute string   `json:"defaultRoute"`
	PayloadFiles []string `json:"payloadFiles"`
	PayloadPath  string   `json:"payloadPath"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

var (
	currentWSConnection  *WebSocketConnection
	connMutex            sync.Mutex
	closeMessageDeadLine = 5
	wsSendBuffer         = 1024
)

type WebSocketConnection struct {
	Conn  *websocket.Conn
	Send  chan []byte
	mutex sync.Mutex
}

func NewWebSocketConnection(conn *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		Conn: conn,
		Send: make(chan []byte, wsSendBuffer),
	}
}

func (wc *WebSocketConnection) enqueueMessage(message []byte) {
	select {
	case wc.Send <- message:
	default:
		log.Println("Send channel is full, dropping message")
	}
}

func (wc *WebSocketConnection) writePump() {
	for {
		message, ok := <-wc.Send
		if !ok {
			// Channel closed, stop the goroutine
			return
		}

		wc.mutex.Lock()
		err := wc.Conn.WriteMessage(websocket.TextMessage, message)
		wc.mutex.Unlock()

		if err != nil {
			break
		}
	}
}

func handleWebSocket(resWriter http.ResponseWriter, req *http.Request) {
	connMutex.Lock()
	oldConn := currentWSConnection
	connMutex.Unlock()

	if oldConn != nil {
		closeConnectionWithMessage(oldConn.Conn, "ws_takeover")
	}

	conn, err := upgrader.Upgrade(resWriter, req, nil)
	if err != nil {
		utils.LogError("WS connection error: ", err)

		return
	}

	connMutex.Lock()

	currentWSConnection = NewWebSocketConnection(conn)
	go currentWSConnection.writePump()
	connMutex.Unlock()
}

func handleAdminReq(resWriter http.ResponseWriter, req *http.Request) {
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

			err = json.Unmarshal(jsonData, &data)
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

			errr := os.WriteFile(jsonName+".json", jsonData, filePermission)
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

			err = json.Unmarshal(jsonData, &data)
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

			errr := os.WriteFile("ui/settings.json", jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			http.ServeFile(resWriter, req, "ui/settings.json")
		}
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
			Port:         Port,
			DefaultRoute: DefaultRoute,
			PayloadFiles: payloadFiles,
			PayloadPath:  config.AppConfig.Defaults.PayloadArchivePath,
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

			err = json.Unmarshal(jsonData, &data)
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

			jsonData, err = json.MarshalIndent(payload, "", "  ")
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}

			const filePermission = 0o600
			errr := os.WriteFile(filePath, jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			queryParams := req.URL.Query()

			value, exists := queryParams["file"]
			if !exists {
				log.Println("Parameter 'file' not found")
				http.Error(resWriter, "Parameter 'file' not found", http.StatusInternalServerError)

				return
			}

			basePath, err := filepath.Abs(config.AppConfig.Defaults.PayloadArchivePath)
			if err != nil {
				http.Error(resWriter, "Invalid archive path", http.StatusInternalServerError)

				return
			}

			requestedPath := filepath.Join(basePath, value[0])
			resolvedPath, err := filepath.Abs(requestedPath)
			if err != nil || !strings.HasPrefix(resolvedPath, basePath) {
				http.Error(resWriter, "Invalid file path", http.StatusBadRequest)

				return
			}

			http.ServeFile(resWriter, req, resolvedPath)
			// log.Println(req.Method + " " + fmt.Sprintf(value[0]))
		}
	default:
		return
	}
}

func closeConnectionWithMessage(conn *websocket.Conn, message string) {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, message)

	deadline := time.Now().Add(time.Second * time.Duration(closeMessageDeadLine))

	err := conn.SetWriteDeadline(deadline)
	if err != nil {
		utils.LogError("Error SetWriteDeadline: ", err)
	}

	err = conn.WriteMessage(websocket.CloseMessage, msg)
	if err != nil {
		utils.LogError("Error WriteMessage close msg: ", err)
	}

	conn.Close()
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
