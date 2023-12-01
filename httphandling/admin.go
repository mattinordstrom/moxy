package httphandling

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/config"
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
		return true // Adjust this as needed
	},
}

var updateClient *websocket.Conn // The client to receive updates

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		return
	}
	updateClient = conn // Store the connection

	// Keep the connection alive, handle incoming messages or disconnection
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			conn.Close()
			updateClient = nil

			break
		}
	}
}

func handleAdminReq(req *http.Request, resWriter http.ResponseWriter) {
	reqURL := fmt.Sprint(req.URL)

	switch reqURL {
	case "/moxyadminui/mockdef", "/moxyadminui/proxydef":
		jsonName := reqURL[strings.LastIndex(reqURL, "/")+1:]
		if req.Method == "POST" {
			// fmt.Println("----------- mockdef proxydef POST -------------")
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

			errr := os.WriteFile(jsonName+".json", jsonData, 0644)
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
			log.Printf("Error reading dir: %v", err)
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
			log.Printf("Error writing response: %v", err)
		}
	case "/moxyadminui":
		http.ServeFile(resWriter, req, "ui/index.html")
	default:
		return
	}
}

func updateAdminWithLatest(evtStr string, evtType string) {
	if updateClient != nil {
		msg := WebSocketMessage{
			Type:    evtType,
			Message: evtStr,
		}

		jsonMsg, jErr := json.Marshal(msg)
		if jErr != nil {
			fmt.Println("Error: json marshal websocket msg")
		}

		err := updateClient.WriteMessage(websocket.TextMessage, jsonMsg)
		if err != nil {
			fmt.Println("Error: WriteMessage websocket")
		}
	}
}
