package consumers

// Notification -
type Notification struct {
	Entity  string                 `json:"entity"`
	Action  string                 `json:"action"`
	OldData map[string]interface{} `json:"old_data"`
	NewData map[string]interface{} `json:"new_data"`
}
