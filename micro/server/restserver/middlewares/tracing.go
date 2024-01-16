package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func TracingHandler(service string) gin.HandlerFunc {
	return otelgin.Middleware(service)
}
