package fiberadapter

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/hrz8/do-function-go-proxy/core"
	"github.com/hrz8/do-function-go-proxy/pkg/proxy"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
)

// FiberAdapter makes it easy to send API Gateway proxy events to a fiber.App.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the *fiber.Ctx
type FiberAdapter struct {
	proxy.RequestAccessor

	app *fiber.App
}

// New creates a new instance of the FiberAdapter object.
// Receives an initialized *fiber.App object - normally created with fiber.New().
// It returns the initialized instance of the FiberAdapter object.
func New(app *fiber.App) *FiberAdapter {
	return &FiberAdapter{app: app}
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (f *FiberAdapter) ProxyWithContext(ctx context.Context, params core.DigitalOceanParameters) (*core.DigitalOceanHTTPResponse, error) {
	httpRequest, err := f.EventToRequestWithContext(ctx, params.HTTP)
	return f.proxyInternal(httpRequest, err)
}

func (f *FiberAdapter) proxyInternal(req *http.Request, err error) (*core.DigitalOceanHTTPResponse, error) {
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	resp := proxy.NewProxyResponseWriter()
	f.adaptor(resp, req)

	proxyResponse, err := resp.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return &proxyResponse, nil
}

func (f *FiberAdapter) adaptor(w http.ResponseWriter, r *http.Request) {
	// New fasthttp request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Convert net/http -> fasthttp request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, utils.StatusMessage(fiber.StatusInternalServerError), fiber.StatusInternalServerError)
		return
	}
	req.Header.SetContentLength(len(body))
	_, _ = req.BodyWriter().Write(body)

	req.Header.SetMethod(r.Method)
	req.SetRequestURI(r.RequestURI)
	req.SetHost(r.Host)
	for key, val := range r.Header {
		for _, v := range val {
			switch key {
			case fiber.HeaderHost,
				fiber.HeaderContentType,
				fiber.HeaderUserAgent,
				fiber.HeaderContentLength,
				fiber.HeaderConnection:
				req.Header.Set(key, v)
			default:
				req.Header.Add(key, v)
			}
		}
	}

	// We need to make sure the net.ResolveTCPAddr call works as it expects a port
	addrWithPort := r.RemoteAddr
	if !strings.Contains(r.RemoteAddr, ":") {
		addrWithPort = r.RemoteAddr + ":80" // assuming a default port
	}

	remoteAddr, err := net.ResolveTCPAddr("tcp", addrWithPort)
	if err != nil {
		fmt.Printf("could not resolve TCP address for addr %s\n", r.RemoteAddr)
		log.Println(err)
		http.Error(w, utils.StatusMessage(fiber.StatusInternalServerError), fiber.StatusInternalServerError)
		return
	}

	// New fasthttp Ctx
	var fctx fasthttp.RequestCtx
	fctx.Init(req, remoteAddr, nil)

	// Pass RequestCtx to Fiber router
	f.app.Handler()(&fctx)

	// Set response headers
	fctx.Response.Header.VisitAll(func(k, v []byte) {
		w.Header().Add(utils.UnsafeString(k), utils.UnsafeString(v))
	})

	// Set response statusCode
	w.WriteHeader(fctx.Response.StatusCode())

	// Set response body
	_, _ = w.Write(fctx.Response.Body())
}
