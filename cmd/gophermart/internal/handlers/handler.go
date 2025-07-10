package handlers

import (
	// "io"

	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/middlewares"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository/customErrors"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
)

type Handler struct {
	service   *services.Service
	secretKey string
}

func NewHandler(service *services.Service, secretKey string) *Handler {
	return &Handler{service: service, secretKey: secretKey}
}

type AuthRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(req.Login, req.Password)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	middlewares.SetAuthCookie(c, user.ID, h.secretKey)
	c.Status(http.StatusOK)
}

func (h *Handler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByLogin(req.Login)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	middlewares.SetAuthCookie(c, user.ID, h.secretKey)
	c.Status(http.StatusOK)
}

func (h *Handler) UploadOrder(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	orderNumberRaw, err := io.ReadAll(c.Request.Body)
	if err != nil || len(orderNumberRaw) == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	orderNumber := string(orderNumberRaw)

	if !services.IsValidLuhn(orderNumber) {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	statusCode, err := h.service.SaveNewOrder(c.Request.Context(), userID, orderNumber)
	if err != nil {
		switch statusCode {
		case http.StatusConflict:
			c.AbortWithStatus(http.StatusConflict)
			return
		case http.StatusOK:
			c.Status(http.StatusOK)
			return
		default:
			log.Printf("UploadOrder error: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	h.service.EnqueueOrderForProcessing(orderNumber)

	c.Status(http.StatusAccepted)
}

func (h *Handler) GetOrders(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	log.Printf("UuserID: %v", userID)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	orders, err := h.service.GetUserOrders(c.Request.Context(), userID)
	log.Printf("orders: %v", orders)
	if err != nil {
		log.Printf("GetUserOrders error: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var req models.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err := h.service.Withdraw(c.Request.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case customErrors.ErrInsufficientBalance:
			c.AbortWithStatus(http.StatusPaymentRequired)
		case customErrors.ErrInvalidOrderNumber:
			c.AbortWithStatus(http.StatusUnprocessableEntity)
		default:
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) GetWithdrawals(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	list, err := h.service.GetWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *Handler) GetUserBalance(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	balance, err := h.service.GetUserBalance(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, balance)
}
