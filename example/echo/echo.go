package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hrz8/do-function-go-proxy/core"
	"github.com/hrz8/do-function-go-proxy/example"
	echoadapter "github.com/hrz8/do-function-go-proxy/pkg/adapter/echo"
	"github.com/hrz8/do-function-go-proxy/pkg/proxy"
	"github.com/labstack/echo/v4"
)

var app *echo.Echo
var adapter example.Adapter

func Main(ctx context.Context, params core.DigitalOceanParameters) (*core.DigitalOceanHTTPResponse, error) {
	// https://domain.com/$namespace
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		return &core.DigitalOceanHTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal server error",
		}, nil
	}

	// $namespace
	namespace := os.Getenv("FUNCTION_NAMESPACE")
	if namespace == "" {
		return &core.DigitalOceanHTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal server error",
		}, nil
	}

	pCtx := proxy.NewContext(ctx, baseUrl).Background()
	path := pCtx.Value("trailing_path").(string)

	app = echo.New()

	// using group to configure basePath
	// since DO function must have a base path for each function
	router := app.Group(fmt.Sprintf("/%s%s", namespace, path))

	router.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World Echo!")
	})

	router.GET("/ping", func(c echo.Context) error {
		ok := &example.PingResponse{Ok: true}
		return c.JSON(200, ok)
	})

	adapter = echoadapter.New(app)
	return adapter.ProxyWithContext(ctx, params)
}
