package core

type DigitalOceanHTTPRequest struct {
	Headers         map[string]string `json:"headers"`
	Path            string            `json:"path"`
	Method          string            `json:"method"`
	Body            string            `json:"body"`
	QueryString     string            `json:"queryString"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type DigitalOceanHTTPResponse struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

type DigitalOceanParameters struct {
	Headers map[string]string       `json:"__ow_headers"`
	Path    string                  `json:"__ow_path"`
	Method  string                  `json:"__ow_method"`
	Body    string                  `json:"__ow_body"`
	Query   string                  `json:"__ow_query"`
	HTTP    DigitalOceanHTTPRequest `json:"http"`
}
