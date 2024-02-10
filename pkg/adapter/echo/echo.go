package echoadapter

import (
	"context"
	"net/http"

	"github.com/hrz8/do-function-go-proxy/core"
	"github.com/hrz8/do-function-go-proxy/pkg/proxy"

	"github.com/labstack/echo/v4"
)

// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type EchoAdapter struct {
	proxy.RequestAccessor

	app *echo.Echo
}

// New creates a new instance of the EchoAdapter object.
// Receives an initialized *echo.Echo object - normally created with echo.New().
// It returns the initialized instance of the EchoAdapter object.
func New(app *echo.Echo) *EchoAdapter {
	return &EchoAdapter{app: app}
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (f *EchoAdapter) ProxyWithContext(ctx context.Context, params core.DigitalOceanParameters) (*core.DigitalOceanHTTPResponse, error) {
	httpRequest, err := f.EventToRequestWithContext(ctx, params.HTTP)
	return f.proxyInternal(httpRequest, err)
}

func (e *EchoAdapter) proxyInternal(req *http.Request, err error) (*core.DigitalOceanHTTPResponse, error) {

	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := proxy.NewProxyResponseWriter()
	e.app.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return &proxyResponse, nil
}
