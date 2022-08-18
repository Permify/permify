package permission

// CheckResponse -
type CheckResponse struct {
	Can            bool        `json:"can"`
	RemainingDepth int32       `json:"remaining_depth"`
	Decisions      interface{} `json:"decisions"`
}
