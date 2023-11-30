package models

type Proxy struct {
	Active  bool   `json:"active"`
	Target  string `json:"target"`
	URLPart string `json:"urlpart"`
	Verbose bool   `json:"verbose"`
}
