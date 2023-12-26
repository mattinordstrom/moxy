package httphandling

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/config"
	"github.com/mattinordstrom/moxy/utils"
)

type WebSocketMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Settings struct {
	Port         int      `json:"port"`
	DefaultRoute string   `json:"defaultRoute"`
	PayloadFiles []string `json:"payloadFiles"`
	PayloadPath  string   `json:"payloadPath"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	currentWSConnection  *websocket.Conn
	connMutex            sync.Mutex
	closeMessageDeadLine = 5
)

func handleWebSocket(resWriter http.ResponseWriter, req *http.Request) {
	if currentWSConnection != nil {
		closeConnectionWithMessage(currentWSConnection, "ws_takeover")
	}

	connMutex.Lock()
	defer connMutex.Unlock()

	conn, err := upgrader.Upgrade(resWriter, req, nil)
	if err != nil {
		utils.LogError("WS connection error: ", err)

		return
	}

	currentWSConnection = conn

	// Keep the connection alive, handle incoming messages or disconnection
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			conn.Close()

			currentWSConnection = nil

			break
		}
	}
}

func handleAdminReq(resWriter http.ResponseWriter, req *http.Request) {
	reqURL := req.URL.Path

	switch reqURL {
	case "/moxyadminui/mockdef", "/moxyadminui/proxydef":
		jsonName := reqURL[strings.LastIndex(reqURL, "/")+1:]

		if req.Method == http.MethodPost {
			jsonData, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}
			defer req.Body.Close()

			var data interface{}

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

			const filePermission = 0o644

			errr := os.WriteFile(jsonName+".json", jsonData, filePermission)
			if errr != nil {
				http.Error(resWriter, errr.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			http.ServeFile(resWriter, req, jsonName+".json")
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

		if _, err := resWriter.Write(jsonResponse); err != nil {
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
				File    string      `json:"file"`
				Payload interface{} `json:"payload"`
			}

			var data RequestData

			err = json.Unmarshal(jsonData, &data)
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusBadRequest)

				return
			}

			// log.Println("File path: " + data.File)

			jsonData, err = json.MarshalIndent(data.Payload, "", "  ")
			if err != nil {
				http.Error(resWriter, err.Error(), http.StatusInternalServerError)

				return
			}

			const filePermission = 0o644

			errr := os.WriteFile(data.File, jsonData, filePermission)
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

			http.ServeFile(resWriter, req, config.AppConfig.Defaults.PayloadArchivePath+value[0])
			// log.Println(req.Method + " " + fmt.Sprintf(value[0]))
		}
	default:
		return
	}
}

func closeConnectionWithMessage(conn *websocket.Conn, message string) {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, message)

	deadline := time.Now().Add(time.Second * time.Duration(closeMessageDeadLine))
	if err := conn.SetWriteDeadline(deadline); err != nil {
		utils.LogError("Error SetWriteDeadline: ", err)
	}

	if err := conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
		utils.LogError("Error WriteMessage close msg: ", err)
	}

	conn.Close()
}

func updateAdminWithLatest(evtStr string, evtType string) {
	if currentWSConnection != nil {
		msg := WebSocketMessage{
			Type:    evtType,
			Message: evtStr,
		}

		jsonMsg, jErr := json.Marshal(msg)
		if jErr != nil {
			utils.LogError("Error: json marshal websocket msg ", jErr)
		}

		err := currentWSConnection.WriteMessage(websocket.TextMessage, jsonMsg)
		if err != nil {
			utils.LogError("Error: WriteMessage websocket ", err)
		}
	}
}
