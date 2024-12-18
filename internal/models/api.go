package models

// a standard response structure for the APIs
type APIResponse struct {
	DataFormat DataFormat  `json:"-" xml:"-"`
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// a standard response structure for the APIs
type SuccessAPIResponse struct {
	StatusCode int         `example:200`
	Message    string      `example: "WITHDRAWAL processing initialized"`
	Data       interface{} `example: {}`
}

// a standard response structure for the APIs
type BadRequestAPIResponse struct {
	StatusCode int    `example:400`
	Message    string `example: "Invalid request body"`
}

// a standard response structure for the APIs
type InternalErrorAPIResponse struct {
	StatusCode int    `example:500`
	Message    string `example:"Internal Error"`
}

type DataFormat string

const (
	XML  DataFormat = "application/xml"
	JSON DataFormat = "application/json"
)
