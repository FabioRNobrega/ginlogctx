// Package ginlogctx enriches Logrus logs with request-scoped fields for Gin handlers.
//
// Automatic enrichment applies only to log entries emitted on the active request
// goroutine while the middleware is executing. Background or detached goroutines
// must propagate request information explicitly if they need the same fields.
//
// The request ID must already be available on the Gin context, so
// requestid.New(...) should run before ginlogctx.Middleware(...).
package ginlogctx
