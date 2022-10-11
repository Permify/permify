package common

// HTTPSuccessResponse -
type HTTPSuccessResponse struct {
	Data interface{} `json:"data"`
}

// Message -
type Message struct {
	Message string `json:"message"`
}

// HTTPErrorResponse -
type HTTPErrorResponse struct {
	Errors interface{} `json:"errors"`
}

// SuccessResponse -
func SuccessResponse(data interface{}) HTTPSuccessResponse {
	return HTTPSuccessResponse{Data: data}
}

// MResponse -
func MResponse(message string) Message {
	return Message{
		Message: message,
	}
}

// ValidationResponse -
func ValidationResponse(data interface{}) HTTPErrorResponse {
	return HTTPErrorResponse{Errors: data}
}
