package models

type Proxy struct {
	Active  bool   `json:"active"`
	Comment string `json:"comment,omitempty"`
	Target  string `json:"target"`
	URLPart string `json:"urlpart"`
	Verbose bool   `json:"verbose"`
}
