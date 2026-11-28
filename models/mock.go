package models

import "encoding/json"

type Mock struct {
	Active          bool            `json:"active"`
	Comment         string          `json:"comment,omitempty"`
	FreezeTimeMS    int             `json:"freezetimems"`
	Method          string          `json:"method"`
	Payload         json.RawMessage `json:"payload"`
	PayloadFromFile string          `json:"payloadFromFile"`
	// Nil means the mock predates the toggle, so the old "file when payload is empty" rule applies
	UsePayloadFromFile *bool  `json:"usePayloadFromFile,omitempty"`
	StatusCode         int    `json:"statuscode"`
	URLPart            string `json:"urlpart"`
}
