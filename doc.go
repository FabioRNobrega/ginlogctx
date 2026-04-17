// Package ginlogctx enriches Logrus logs with request-scoped fields for Gin handlers.
//
// Automatic enrichment applies only to log entries emitted on the active request
// goroutine while the middleware is executing. Background or detached goroutines
// must propagate request information explicitly if they need the same fields.
//
// Middleware provides request ID handling out of the box and enriches logs with
// request_id by default. Additional request-scoped fields can be registered
// through Config.Fields or Config.AdditionalFields.
package ginlogctx
