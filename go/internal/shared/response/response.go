package response

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorPayload struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

type Response struct {
	Success bool          `json:"success"`
	Data    any           `json:"data,omitempty"`
	Meta    any           `json:"meta,omitempty"`
	Error   *ErrorPayload `json:"error,omitempty"`
}

func Success(data any, meta any) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

func Fail(code, message string, details []ErrorDetail) Response {
	if details == nil {
		details = []ErrorDetail{}
	}
	return Response{
		Success: false,
		Error: &ErrorPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

