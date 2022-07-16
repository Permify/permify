package responses

type Check struct {
	Can       bool        `json:"can"`
	Decisions interface{} `json:"decisions"`
}
