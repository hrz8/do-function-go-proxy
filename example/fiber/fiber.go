package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/hrz8/do-function-go-proxy/core"
	"github.com/hrz8/do-function-go-proxy/example"
	fiberadapter "github.com/hrz8/do-function-go-proxy/pkg/adapter/fiber"
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

	app := fiber.New()

	// using group to configure basePath
	// since DO function must have a base path for each function
	router := app.Group(fmt.Sprintf("/%s%s", namespace, path))

	router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World Fiber!")
	})

	router.Get("/ping", func(c *fiber.Ctx) error {
		ok := &example.PingResponse{Ok: true}
		return c.JSON(ok)
	})

	adapter = fiberadapter.New(app)
	return adapter.ProxyWithContext(ctx, params)
}
