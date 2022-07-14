package responses

type HTTPSuccessResponse struct {
	Data interface{} `json:"data"`
}

type Message struct {
	Message string `json:"message"`
}

type HTTPErrorResponse struct {
	Errors interface{} `json:"errors"`
}

func SuccessResponse(data interface{}) HTTPSuccessResponse {
	return HTTPSuccessResponse{Data: data}
}

func MResponse(message string) Message {
	return Message{
		Message: message,
	}
}

func ValidationResponse(data interface{}) HTTPErrorResponse {
	return HTTPErrorResponse{Errors: data}
}
