package responses

// Check -
type Check struct {
	Can            bool        `json:"can"`
	RemainingDepth int32       `json:"remaining_depth"`
	Decisions      interface{} `json:"decisions"`
}
