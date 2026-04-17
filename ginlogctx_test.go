package ginlogctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const testUserIDHeader = "X-User-ID"

const (
	testRequestID1 = "11111111-1111-1111-1111-111111111111"
	testRequestID2 = "22222222-2222-2222-2222-222222222222"
	testRequestID3 = "33333333-3333-3333-3333-333333333333"
	testRequestID4 = "44444444-4444-4444-4444-444444444444"
	testHandlerDelay = 15 * time.Millisecond
)

func TestHookAddsScopedFields(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
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
	req.Header.Set(defaultRequestIDHeader, testRequestID1)
	req.Header.Set(testUserIDHeader, "user-1")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if got, want := resp.Code, http.StatusNoContent; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}

	logEntry := findEntry(t, entries, "inside handler")
	logCapturedEntry(t, "scoped handler log", logEntry)
	if logEntry[defaultRequestIDField] != testRequestID1 {
		t.Fatalf("request_id = %v, want %s", logEntry[defaultRequestIDField], testRequestID1)
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
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
	router.Use(Middleware(DefaultConfig()))
	router.GET("/hook", func(c *gin.Context) {
		logger.WithField(defaultRequestIDField, "explicit").Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Set(defaultRequestIDHeader, testRequestID2)
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
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
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
		time.Sleep(testHandlerDelay)
		c.Status(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/done", nil)
	req.Header.Set(defaultRequestIDHeader, testRequestID3)
	req.Header.Set(testUserIDHeader, "user-3")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	logEntry := findEntry(t, entries, defaultRequestLogMsg)
	logCapturedEntry(t, "request completion log", logEntry)
	if logEntry["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %s", logEntry["method"], http.MethodGet)
	}
	if logEntry["path"] != "/done" {
		t.Fatalf("path = %v, want /done", logEntry["path"])
	}
	if got := int(logEntry["status"].(float64)); got != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", got, http.StatusAccepted)
	}
	if got := int64(logEntry["durationMs"].(float64)); got < testHandlerDelay.Milliseconds() {
		t.Fatalf("durationMs = %d, want at least %d", got, testHandlerDelay.Milliseconds())
	}
	if logEntry[defaultRequestIDField] != testRequestID3 {
		t.Fatalf("request_id = %v, want %s", logEntry[defaultRequestIDField], testRequestID3)
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
	setupStandardLoggerForTest(t, logger)

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
	router.Use(Middleware(cfg))
	router.GET("/hook", func(c *gin.Context) {
		logger.Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Set(defaultRequestIDHeader, testRequestID4)
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
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
	router.Use(Middleware(DefaultConfig()))
	router.GET("/done", func(c *gin.Context) {
		logger.Info("inside handler")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/done", nil)
	req.Header.Set(defaultRequestIDHeader, testRequestID4)
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
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
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
			reqID := fmt.Sprintf("00000000-0000-0000-0000-%012d", i)
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

	logCapturedEntry(t, "concurrent request completion log", findEntry(t, entries, defaultRequestLogMsg))
}

func TestMiddlewareGeneratesRequestIDOutOfBox(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, entries := newTestLogger()
	Install(logger, DefaultConfig())
	setupStandardLoggerForTest(t, logger)

	router := gin.New()
	router.Use(Middleware(DefaultConfig()))
	router.GET("/generated", func(c *gin.Context) {
		logger.Info("generated request id")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/generated", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	requestID := resp.Header().Get(defaultRequestIDHeader)
	if requestID == "" {
		t.Fatalf("response %s header is empty", defaultRequestIDHeader)
	}

	logEntry := findEntry(t, entries, "generated request id")
	logCapturedEntry(t, "generated request id log", logEntry)
	if logEntry[defaultRequestIDField] != requestID {
		t.Fatalf("request_id = %v, want %s", logEntry[defaultRequestIDField], requestID)
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

func logCapturedEntry(t *testing.T, label string, entry map[string]any) {
	t.Helper()

	encoded, err := json.Marshal(entry)
	if err != nil {
		t.Logf("%s: <unable to marshal: %v>", label, err)
		return
	}

	t.Logf("%s: %s", label, encoded)
}

func setupStandardLoggerForTest(t *testing.T, logger *logrus.Logger) {
	t.Helper()

	original := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalHooks := logrus.StandardLogger().Hooks
	originalReportCaller := logrus.StandardLogger().ReportCaller
	originalLevel := logrus.StandardLogger().Level

	logrus.SetOutput(logger.Out)
	logrus.SetFormatter(logger.Formatter)
	logrus.SetReportCaller(logger.ReportCaller)
	logrus.SetLevel(logger.Level)
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
		logrus.SetLevel(originalLevel)
		logrus.StandardLogger().ReplaceHooks(originalHooks)
	})
}
