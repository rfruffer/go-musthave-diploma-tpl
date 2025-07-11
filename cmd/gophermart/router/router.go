package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/handlers"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/middlewares"
)

type Router struct {
	Handler   *handlers.Handler
	SecretKey string
}

func SetupRouter(rt Router) http.Handler {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	r := gin.New()

	middlewares.InitLogger(sugar)
	r.Use(middlewares.GinLoggingMiddleware())
	r.Use(gin.Recovery())

	r.POST("/api/user/register", rt.Handler.Register)
	r.POST("/api/user/login", rt.Handler.Login)

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware(rt.SecretKey))

	auth.POST("/api/user/orders", rt.Handler.UploadOrder)
	auth.GET("/api/user/orders", rt.Handler.GetOrders)

	auth.POST("/api/user/balance/withdraw", rt.Handler.Withdraw)
	auth.GET("/api/user/withdrawals", rt.Handler.GetWithdrawals)

	auth.GET("/api/user/balance", rt.Handler.GetUserBalance)

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "invalid request")
	})

	r.NoMethod(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "invalid request")
	})

	return r
}
