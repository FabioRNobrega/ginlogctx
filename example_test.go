package ginlogctx_test

import (
	"github.com/FabioRNobrega/ginlogctx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ExampleInstall() {
	logger := logrus.StandardLogger()
	ginlogctx.Install(logger, ginlogctx.DefaultConfig())

	router := gin.New()
	router.Use(ginlogctx.Middleware(ginlogctx.DefaultConfig()))
}

func ExampleMiddleware_customFields() {
	cfg := ginlogctx.DefaultConfig()
	cfg.RequestIDField = "trace_id"
	cfg.RequestLogMessage = "http request finished"
	cfg.Fields = []ginlogctx.Field{
		{
			Key: "product_id",
			Resolve: func(c *gin.Context) (any, bool) {
				productID := c.GetHeader("X-Product-ID")
				return productID, productID != ""
			},
		},
	}

	router := gin.New()
	router.Use(ginlogctx.Middleware(cfg))
}
