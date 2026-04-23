<p align="center">
	<img src="./docs/package-logo-gpt-generated.png" width="300" height="300" alt="ginlogctx logo" />
</p>

<h1 align="center">ginlogctx</h1>
<p align="center">
    <em>Request-scoped Logrus fields for Gin, with sensible defaults and flexible custom context.</em>
</p>

<p align="center">
    <a href="https://goreportcard.com/report/github.com/FabioRNobrega/ginlogctx" style="text-decoration: none;">
        <img src="https://goreportcard.com/badge/github.com/FabioRNobrega/ginlogctx" alt="Go Report Card">
    </a>
    <a href="https://opensource.org/licenses/MIT" style="text-decoration: none;">
        <img src="https://img.shields.io/badge/License-MIT-brightgreen.svg" alt="MIT License">
    </a>
    <a href="https://pkg.go.dev/github.com/FabioRNobrega/ginlogctx" style="text-decoration: none;">
        <img src="https://pkg.go.dev/badge/github.com/FabioRNobrega/ginlogctx.svg" alt="GoDoc">
    </a>
    <a href="https://github.com/FabioRNobrega/ginlogctx/pulls" style="text-decoration: none;">
        <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions welcome">
    </a>
</p>

`ginlogctx` is a small Go package that gives Gin services an out-of-the-box
request ID flow and enriches `logrus` entries with request-scoped fields inside
handlers.

Under the hood, it builds on the idea behind `github.com/gin-contrib/requestid`
and packages that behavior together with Logrus request-scoped enrichment, so a
service can:
- generate or reuse a request ID for every Gin request
- attach that `request_id` automatically to application logs
- extend the same request context with custom fields such as `user_id`

The goal is to cover the common "I need correlated logs per request" use case
without forcing small and medium projects into the complexity of a full
observability stack.

Out of the box it adds:
- `request_id`

And it lets applications add any extra request-scoped fields they want, such as:
- `user_id`
- `product_id`
- `tenant_id`
- `account_id`

It is designed for teams that already use:
- `github.com/gin-gonic/gin`
- `github.com/sirupsen/logrus`

If you are comparing this package with a tracing stack, see:
- [OpenTelemetry vs ginlogctx](./docs/OPENTELEMETRY.md)

## Installation

```shell
go get github.com/FabioRNobrega/ginlogctx
```

Or add it to your `go.mod`:

```go
require github.com/FabioRNobrega/ginlogctx
```

## Run Tests With Docker

If you do not want to install Go locally, you can run the package tests with:

```shell
make docker-test
```

The Docker setup mounts the repository into a Go container and keeps module and
build caches in named Docker volumes so repeated test runs are faster.

## Features

- Provides request ID handling out of the box and adds `request_id` automatically
- Enriches plain `logrus.Info/Error/Warn/...` calls through a Logrus hook
- Supports custom request-scoped fields through resolvers
- Includes optional request completion logging
- Keeps the built-in request completion log focused on HTTP fields instead of caller metadata
- Preserves explicitly set log fields instead of overwriting them
- Keeps setup small and easy to drop into existing Gin services

## Why Use It

`ginlogctx` is a good fit when you want:
- request-level correlation in logs
- a simple `request_id` story for Gin services
- custom fields like `user_id`, `tenant_id`, or `product_id`
- log-based tracking in tools such as Datadog without introducing tracing or APM first

It is intentionally a lighter approach than full distributed tracing. For many
services, especially internal APIs and smaller systems, correlated logs are
enough to answer:
- which logs belong to this request?
- which user triggered it?
- what happened across my services if I forward the same request ID?

If later you need full spans, distributed traces, baggage, and cross-process
trace visualization, you can still adopt OpenTelemetry on top of or alongside
this approach.

## How It Works

`ginlogctx` binds request fields for the lifetime of the active Gin request and uses a Logrus hook to inject them into log entries emitted on that same request goroutine.

This means:
- It works automatically for logs emitted during normal handler execution
- It does not automatically follow spawned goroutines
- Background work should propagate context explicitly if you want the same fields there

## Quick Start

This is the minimal setup. It gives you `request_id` out of the box.

```go
package main

import (
	"github.com/FabioRNobrega/ginlogctx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	ginlogctx.Install(logrus.StandardLogger(), ginlogctx.DefaultConfig())

	r := gin.New()
	r.Use(ginlogctx.Middleware(ginlogctx.DefaultConfig()))

	r.GET("/ping", func(c *gin.Context) {
		logrus.Info("inside handler")
		c.JSON(200, gin.H{"message": "pong"})
	})

	_ = r.Run(":8080")
}
```

Example log output:

```json
{
  "level": "info",
  "msg": "inside handler",
  "request_id": "9f5275d0-3fbc-47b4-9aa9-9f29767e4f6e",
  "time": "2026-04-17T12:00:00Z"
}
```

## Custom Fields

The recommended extension point is `Config.Fields`.

Each field defines:
- the log key
- how to resolve its value from the current `*gin.Context`

### Add `user_id`

```go
cfg := ginlogctx.DefaultConfig()
cfg.Fields = []ginlogctx.Field{
	{
		Key: "user_id",
		Resolve: func(c *gin.Context) (any, bool) {
			userID := c.GetHeader("X-User-ID")
			return userID, userID != ""
		},
	},
}
```

### Add `product_id`

```go
cfg := ginlogctx.DefaultConfig()
cfg.Fields = []ginlogctx.Field{
	{
		Key: "product_id",
		Resolve: func(c *gin.Context) (any, bool) {
			productID := c.Param("product_id")
			return productID, productID != ""
		},
	},
}
```

### Add multiple fields

```go
cfg := ginlogctx.DefaultConfig()
cfg.Fields = []ginlogctx.Field{
	{
		Key: "user_id",
		Resolve: func(c *gin.Context) (any, bool) {
			userID := c.GetHeader("X-User-ID")
			return userID, userID != ""
		},
	},
	{
		Key: "product_id",
		Resolve: func(c *gin.Context) (any, bool) {
			productID := c.Param("product_id")
			return productID, productID != ""
		},
	},
	{
		Key: "tenant_id",
		Resolve: func(c *gin.Context) (any, bool) {
			tenantID := c.GetHeader("X-Tenant-ID")
			return tenantID, tenantID != ""
		},
	},
}
```

## Full Example

```go
package main

import (
	"github.com/FabioRNobrega/ginlogctx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := ginlogctx.DefaultConfig()
	cfg.Fields = []ginlogctx.Field{
		{
			Key: "user_id",
			Resolve: func(c *gin.Context) (any, bool) {
				userID := c.GetHeader("X-User-ID")
				return userID, userID != ""
			},
		},
		{
			Key: "product_id",
			Resolve: func(c *gin.Context) (any, bool) {
				productID := c.Param("product_id")
				return productID, productID != ""
			},
		},
	}

	ginlogctx.Install(logrus.StandardLogger(), cfg)

	r := gin.New()
	r.Use(ginlogctx.Middleware(cfg))

	r.GET("/products/:product_id", func(c *gin.Context) {
		logrus.WithField("operation", "fetch_product").Info("handling request")
		c.JSON(200, gin.H{"ok": true})
	})

	_ = r.Run(":8080")
}
```

Possible output:

```json
{
  "level": "info",
  "msg": "handling request",
  "operation": "fetch_product",
  "product_id": "p-123",
  "request_id": "4c784c9d-ec4b-4556-a650-0eaf8111ef0d",
  "user_id": "u-456",
  "time": "2026-04-17T12:00:00Z"
}
```

## Request Completion Logs

By default, `ginlogctx` also emits a request completion log with the message `request completed` and these fields:
- `method`
- `path`
- `status`
- `durationMs`

You can customize both the message and the level:

```go
cfg := ginlogctx.DefaultConfig()
cfg.RequestLogMessage = "http request finished"
cfg.RequestLogLevel = logrus.DebugLevel
```

Or disable it:

```go
cfg := ginlogctx.DefaultConfig()
cfg.IncludeRequestLog = false
```

This only disables the built-in request completion log emitted by `ginlogctx`.
Request-scoped fields such as `request_id` and your custom fields still continue to
be attached to the other `logrus` entries emitted during the request.

## API

Main types and functions:

- `ginlogctx.DefaultConfig() Config`
- `ginlogctx.Install(logger *logrus.Logger, cfg Config)`
- `ginlogctx.NewHook(cfg Config) logrus.Hook`
- `ginlogctx.Middleware(cfg Config) gin.HandlerFunc`
- `type ginlogctx.Field`
- `type ginlogctx.Config`

Useful `Config` fields:

- `IncludeRequestLog` enables or disables the built-in request completion log
- `RequestLogMessage` customizes the request completion message
- `RequestLogLevel` customizes the level used for the request completion log
- `Fields` registers custom request-scoped fields such as `user_id`

## Notes

- `request_id` is the only built-in field
- Custom fields are application-defined
- Explicit fields already present on a log entry are not overwritten by the hook
- Empty custom field values are ignored
- The package is intended for `logrus` standard/global logging patterns

## Contributing

Contributions are very welcome, whether they are:
- bug fixes
- better tests
- docs improvements
- examples
- API refinements

If you are using `ginlogctx` in a real Gin service and want to improve its ergonomics, that is especially useful feedback.

For the repository Git workflow, branch naming, commit prefixes, and tag conventions, see:

- [Contributing Guide](./docs/CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)

### Contributors

![contributors](https://contrib.rocks/image?repo=FabioRNobrega/ginlogctx)


## License

MIT [© FabioRNobrega](./LICENSE)
