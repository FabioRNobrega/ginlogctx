package ginlogctx

import (
	"time"

	ginrequestid "github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Middleware binds request-scoped fields for the lifetime of the current Gin
// request.
//
// The middleware captures request_id plus any configured custom fields, makes
// them available to the hook for logs emitted on the current request goroutine,
// and optionally emits a request completion log after the handler chain
// finishes.
func Middleware(cfg Config) gin.HandlerFunc {
	cfg = normalizeConfig(cfg)

	return func(c *gin.Context) {
		fields := scopedFieldsForRequest(c, cfg)
		unbind := bindFields(fields)
		defer unbind()

		startedAt := time.Now()
		c.Next()

		if !cfg.IncludeRequestLog {
			return
		}

		logFields := logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"durationMs": time.Since(startedAt).Milliseconds(),
		}

		logrus.StandardLogger().
			WithFields(logFields).
			Log(cfg.RequestLogLevel, cfg.RequestLogMessage)
	}
}

func scopedFieldsForRequest(c *gin.Context, cfg Config) logrus.Fields {
	fields := logrus.Fields{}

	requestID := ginrequestid.Get(c)
	if requestID == "" && cfg.RequestIDHeader != "" {
		requestID = c.GetHeader(cfg.RequestIDHeader)
	}
	if requestID != "" {
		fields[cfg.RequestIDField] = requestID
	}

	for _, field := range cfg.Fields {
		if field.Key == "" || field.Resolve == nil {
			continue
		}

		value, ok := field.Resolve(c)
		if !ok || isEmptyFieldValue(value) {
			continue
		}

		fields[field.Key] = value
	}

	if cfg.AdditionalFields != nil {
		for key, value := range cfg.AdditionalFields(c) {
			if key == "" || isEmptyFieldValue(value) {
				continue
			}
			fields[key] = value
		}
	}

	if len(fields) == 0 {
		return nil
	}

	return fields
}

func isEmptyFieldValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == ""
	default:
		return false
	}
}
