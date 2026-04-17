package ginlogctx

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	defaultRequestIDField  = "request_id"
	defaultRequestIDHeader = "X-Request-ID"
	defaultRequestLogMsg   = "request completed"
)

// Field describes a custom request-scoped log field.
//
// Resolve is called for each request handled by Middleware. If it returns
// ok=true and a non-empty value, the field is attached to request-scoped logs
// under Key.
type Field struct {
	Key     string
	Resolve func(*gin.Context) (any, bool)
}

// Config controls how request-scoped fields are collected and logged.
//
// By default ginlogctx adds only request_id. Additional request fields such as
// user_id or product_id can be registered through Fields or AdditionalFields.
type Config struct {
	RequestIDField    string
	RequestIDHeader   string
	Fields            []Field
	IncludeRequestLog bool
	RequestLogLevel   logrus.Level
	RequestLogMessage string
	AdditionalFields  func(*gin.Context) logrus.Fields
}

// DefaultConfig returns the default ginlogctx configuration.
//
// The default setup adds request_id, emits a request completion log at info
// level, and uses the X-Request-ID header as the request ID fallback.
func DefaultConfig() Config {
	return Config{
		RequestIDField:    defaultRequestIDField,
		RequestIDHeader:   defaultRequestIDHeader,
		IncludeRequestLog: true,
		RequestLogLevel:   logrus.InfoLevel,
		RequestLogMessage: defaultRequestLogMsg,
	}
}

func normalizeConfig(cfg Config) Config {
	defaults := DefaultConfig()
	if cfg.RequestIDField == "" {
		cfg.RequestIDField = defaults.RequestIDField
	}
	if cfg.RequestIDHeader == "" {
		cfg.RequestIDHeader = defaults.RequestIDHeader
	}
	if cfg.RequestLogMessage == "" {
		cfg.RequestLogMessage = defaults.RequestLogMessage
	}
	if cfg.RequestLogLevel < logrus.PanicLevel || cfg.RequestLogLevel > logrus.TraceLevel {
		cfg.RequestLogLevel = defaults.RequestLogLevel
	}
	if isZeroConfig(cfg) {
		cfg.IncludeRequestLog = defaults.IncludeRequestLog
	}

	return cfg
}

func isZeroConfig(cfg Config) bool {
	return cfg.RequestIDField == defaultRequestIDField &&
		cfg.RequestIDHeader == defaultRequestIDHeader &&
		len(cfg.Fields) == 0 &&
		cfg.RequestLogLevel == logrus.InfoLevel &&
		cfg.RequestLogMessage == defaultRequestLogMsg &&
		cfg.AdditionalFields == nil &&
		!cfg.IncludeRequestLog
}
