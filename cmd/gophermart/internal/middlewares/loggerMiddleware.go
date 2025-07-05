package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

type responseData struct {
	status int
	size   int
}

func InitLogger(logger *zap.SugaredLogger) {
	sugar = *logger
}

func GinLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		respData := responseData{
			status: c.Writer.Status(),
			size:   c.Writer.Size(),
		}

		sugar.Infow("incoming request",
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"status", respData.status,
			"duration", duration,
			"size", respData.size,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
