package common

type ErrResponse struct {
	Code             string      `json:"code"`
	Message          string      `json:"msg"`
	ValidationErrors interface{} `json:"validation_errors,omitempty"`
	Details          interface{} `json:"details,omitempty"`
}

type ValidationError struct {
	FailedField string
	Tag         string
	Value       string
}

//Catchall error

var UnexpectedError = ErrResponse{
	Code:    "Z500",
	Message: "Unexpected Error",
}
var BadRequest = ErrResponse{
	Code:    "Z400",
	Message: "Bad Request",
}

type NonceValue struct {
	Value string `json:"value"`
}
