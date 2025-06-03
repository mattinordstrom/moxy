package models

import "encoding/json"

type Mock struct {
	Active          bool            `json:"active"`
	FreezeTimeMS    int             `json:"freezetimems"`
	Method          string          `json:"method"`
	Payload         json.RawMessage `json:"payload"`
	PayloadFromFile string          `json:"payloadFromFile"`
	StatusCode      int             `json:"statuscode"`
	URLPart         string          `json:"urlpart"`
}
