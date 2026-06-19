package dto

import "net/http"

type Request struct {
	Body               []byte
	URL                string
	Headers            map[string]string
	Method             string
	Timeout            int
	RetryCount         int
	ExpectedStatusCode int
}

type Response struct {
	Success    bool
	Body       map[string]interface{} `json:"data"`
	Headers    http.Header
	StatusCode int
}

var ErrorResponse struct {
	Error string `json:"error"`
}
