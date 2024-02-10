package proxy

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/hrz8/do-function-go-proxy/core"
)

// RequestAccessor objects give access to custom API Gateway properties
// in the request.
type RequestAccessor struct {
	stripBasePath string
}

// EventToRequestWithContext converts an API Gateway proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetAPIGatewayContextFromContext, GetStageVarsFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessor) EventToRequestWithContext(ctx context.Context, req core.DigitalOceanHTTPRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(ctx, req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return httpRequest, nil
}

// EventToRequest converts an API Gateway proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessor) EventToRequest(ctx context.Context, req core.DigitalOceanHTTPRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.Path

	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	serverAddress := ctx.Value("app_host").(string)
	path = serverAddress + path

	if len(req.QueryString) > 0 {
		path += "?" + req.QueryString
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.Method),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.Method, req.Path)
		log.Println(err)
		return nil, err
	}

	httpRequest.RemoteAddr = req.Headers["do-connecting-ip"]

	if req.Headers["cookie"] != "" {
		httpRequest.Header.Add("Cookie", req.Headers["cookie"])
	}

	singletonHeaders, headers := splitSingletonHeaders(req.Headers)

	for headerKey, headerValue := range singletonHeaders {
		httpRequest.Header.Add(headerKey, headerValue)
	}

	for headerKey, headerValue := range headers {
		for _, val := range strings.Split(headerValue, ",") {
			httpRequest.Header.Add(headerKey, strings.Trim(val, " "))
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

// splitSingletonHeaders splits the headers into single-value headers and other,
// multi-value capable, headers.
// Returns (single-value headers, multi-value-capable headers)
func splitSingletonHeaders(headers map[string]string) (map[string]string, map[string]string) {
	singletons := make(map[string]string)
	multitons := make(map[string]string)
	for headerKey, headerValue := range headers {
		if ok := singletonHeaders[textproto.CanonicalMIMEHeaderKey(headerKey)]; ok {
			singletons[headerKey] = headerValue
		} else {
			multitons[headerKey] = headerValue
		}
	}

	return singletons, multitons
}

// singletonHeaders is a set of headers, that only accept a single
// value which may be comma separated (according to RFC 7230)
var singletonHeaders = map[string]bool{
	"Content-Type":        true,
	"Content-Disposition": true,
	"Content-Length":      true,
	"User-Agent":          true,
	"Referer":             true,
	"Host":                true,
	"Authorization":       true,
	"Proxy-Authorization": true,
	"If-Modified-Since":   true,
	"If-Unmodified-Since": true,
	"From":                true,
	"Location":            true,
	"Max-Forwards":        true,
}
