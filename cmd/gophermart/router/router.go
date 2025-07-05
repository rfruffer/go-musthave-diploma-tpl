package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/handlers"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/middlewares"
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

	/*
	   POST /api/user/register — регистрация пользователя;
	   POST /api/user/login — аутентификация пользователя;
	   POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
	   GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
	   GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
	   POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	   GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
	*/

	middlewares.InitLogger(sugar)
	r.Use(middlewares.GinLoggingMiddleware())
	r.Use(gin.Recovery())

	r.POST("/api/user/register", rt.Handler.Register)
	r.POST("/api/user/login", rt.Handler.Login)

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware(rt.SecretKey))

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "invalid request")
	})

	r.NoMethod(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "invalid request")
	})

	return r
}
