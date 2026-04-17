package ginlogctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ginrequestid "github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const testUserIDHeader = "X-User-ID"

func TestHookAddsScopedFields(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	cfg := DefaultConfig()
	cfg.Fields = []Field{
		{
			Key: "user_id",
			Resolve: func(c *gin.Context) (any, bool) {
				userID := c.GetHeader(testUserIDHeader)
				return userID, userID != ""
			},
		},
	}
	router.Use(Middleware(cfg))
	router.GET("/hook", func(c *gin.Context) {
		logger.Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Set(defaultRequestIDHeader, "req-1")
	req.Header.Set(testUserIDHeader, "user-1")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if got, want := resp.Code, http.StatusNoContent; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}

	logEntry := findEntry(t, entries, "inside handler")
	if logEntry[defaultRequestIDField] != "req-1" {
		t.Fatalf("request_id = %v, want req-1", logEntry[defaultRequestIDField])
	}
	if logEntry["user_id"] != "user-1" {
		t.Fatalf("user_id = %v, want user-1", logEntry["user_id"])
	}
}

func TestHookDoesNotOverwriteExplicitFields(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	router.Use(Middleware(DefaultConfig()))
	router.GET("/hook", func(c *gin.Context) {
		logger.WithField(defaultRequestIDField, "explicit").Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Set(defaultRequestIDHeader, "req-2")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	logEntry := findEntry(t, entries, "inside handler")
	if logEntry[defaultRequestIDField] != "explicit" {
		t.Fatalf("request_id = %v, want explicit", logEntry[defaultRequestIDField])
	}
}

func TestHookOutsideRequestScope(t *testing.T) {
	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	logger.Info("outside request")

	logEntry := findEntry(t, entries, "outside request")
	if _, exists := logEntry[defaultRequestIDField]; exists {
		t.Fatalf("request_id unexpectedly present")
	}
}

func TestMiddlewareRequestCompletedLog(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	logger.SetReportCaller(true)
	Install(logger, DefaultConfig())
	original := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalHooks := logrus.StandardLogger().Hooks
	originalReportCaller := logrus.StandardLogger().ReportCaller
	logrus.SetOutput(logger.Out)
	logrus.SetFormatter(logger.Formatter)
	logrus.SetReportCaller(logger.ReportCaller)
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
	for _, level := range logrus.AllLevels {
		for _, hook := range logger.Hooks[level] {
			logrus.StandardLogger().AddHook(hook)
		}
	}
	t.Cleanup(func() {
		logrus.SetOutput(original)
		logrus.SetFormatter(originalFormatter)
		logrus.SetReportCaller(originalReportCaller)
		logrus.StandardLogger().ReplaceHooks(originalHooks)
	})

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	cfg := DefaultConfig()
	cfg.Fields = []Field{
		{
			Key: "user_id",
			Resolve: func(c *gin.Context) (any, bool) {
				userID := c.GetHeader(testUserIDHeader)
				return userID, userID != ""
			},
		},
	}
	router.Use(Middleware(cfg))
	router.GET("/done", func(c *gin.Context) {
		c.Status(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/done", nil)
	req.Header.Set(defaultRequestIDHeader, "req-3")
	req.Header.Set(testUserIDHeader, "user-3")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	logEntry := findEntry(t, entries, defaultRequestLogMsg)
	if logEntry["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %s", logEntry["method"], http.MethodGet)
	}
	if logEntry["path"] != "/done" {
		t.Fatalf("path = %v, want /done", logEntry["path"])
	}
	if got := int(logEntry["status"].(float64)); got != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", got, http.StatusAccepted)
	}
	if logEntry[defaultRequestIDField] != "req-3" {
		t.Fatalf("request_id = %v, want req-3", logEntry[defaultRequestIDField])
	}
	if logEntry["user_id"] != "user-3" {
		t.Fatalf("user_id = %v, want user-3", logEntry["user_id"])
	}
	if _, exists := logEntry["file"]; exists {
		t.Fatalf("file unexpectedly present on request completion log")
	}
	if _, exists := logEntry["func"]; exists {
		t.Fatalf("func unexpectedly present on request completion log")
	}
}

func TestMiddlewareSupportsCustomFields(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	cfg := DefaultConfig()
	cfg.Fields = []Field{
		{
			Key: "product_id",
			Resolve: func(c *gin.Context) (any, bool) {
				productID := c.GetHeader("X-Product-ID")
				return productID, productID != ""
			},
		},
	}

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	router.Use(Middleware(cfg))
	router.GET("/hook", func(c *gin.Context) {
		logger.Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Set(defaultRequestIDHeader, "req-product")
	req.Header.Set("X-Product-ID", "product-7")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	logEntry := findEntry(t, entries, "inside handler")
	if logEntry["product_id"] != "product-7" {
		t.Fatalf("product_id = %v, want product-7", logEntry["product_id"])
	}
}

func TestBindingClearedAfterRequest(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	router.Use(Middleware(DefaultConfig()))
	router.GET("/done", func(c *gin.Context) {
		logger.Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/done", nil)
	req.Header.Set(defaultRequestIDHeader, "req-4")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	logger.Info("outside handler")

	logEntry := findEntry(t, entries, "outside handler")
	if _, exists := logEntry[defaultRequestIDField]; exists {
		t.Fatalf("request_id unexpectedly present after request")
	}
}

func TestConcurrentRequestsDoNotLeakFields(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())

	router := gin.New()
	router.Use(ginrequestid.New(ginrequestid.WithCustomHeaderStrKey(defaultRequestIDHeader)))
	router.Use(Middleware(DefaultConfig()))
	router.GET("/parallel", func(c *gin.Context) {
		logger.WithField("message_id", c.GetHeader(defaultRequestIDHeader)).Info("parallel")
		c.Status(http.StatusNoContent)
	})

	const requests = 12
	var wg sync.WaitGroup
	for i := range requests {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/parallel", nil)
			reqID := fmt.Sprintf("req-%d", i)
			req.Header.Set(defaultRequestIDHeader, reqID)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
		}(i)
	}
	wg.Wait()

	seen := map[string]string{}
	for _, entry := range entries.snapshot() {
		if entry["msg"] != "parallel" {
			continue
		}
		msgID := entry["message_id"].(string)
		reqID := entry[defaultRequestIDField].(string)
		seen[msgID] = reqID
	}

	if len(seen) != requests {
		t.Fatalf("seen %d request logs, want %d", len(seen), requests)
	}
	for key, value := range seen {
		if key != value {
			t.Fatalf("message_id %q was enriched with request_id %q", key, value)
		}
	}
}

type capturedEntries struct {
	mu      sync.Mutex
	entries []map[string]any
}

func (c *capturedEntries) Write(p []byte) (int, error) {
	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(p), &entry); err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.entries = append(c.entries, entry)
	c.mu.Unlock()

	return len(p), nil
}

func (c *capturedEntries) snapshot() []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]map[string]any, len(c.entries))
	copy(out, c.entries)
	return out
}

func newTestLogger() (*logrus.Logger, *capturedEntries) {
	writer := &capturedEntries{}
	logger := logrus.New()
	logger.SetOutput(writer)
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger, writer
}

func findEntry(t *testing.T, entries *capturedEntries, message string) map[string]any {
	t.Helper()

	for _, entry := range entries.snapshot() {
		if entry["msg"] == message {
			return entry
		}
	}

	t.Fatalf("log entry %q not found", message)
	return nil
}
