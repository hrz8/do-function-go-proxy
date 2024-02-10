package core

import (
	"fmt"
	"net/http"
)

// GatewayTimeout returns a dafault Gateway Timeout (504) response
func GatewayTimeout() *DigitalOceanHTTPResponse {
	return &DigitalOceanHTTPResponse{StatusCode: http.StatusGatewayTimeout}
}

// NewLoggedError generates a new error and logs it to stdout
func NewLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}
