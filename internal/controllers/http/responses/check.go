package responses

type Check struct {
	Can            bool        `json:"can"`
	RemainingDepth int         `json:"remaining_depth"`
	Decisions      interface{} `json:"decisions"`
}
