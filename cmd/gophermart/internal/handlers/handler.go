package handlers

import (
	// "io"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/middlewares"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
)

type Handler struct {
	service   *services.Service
	secretKey string
	// DeleteChan chan async.DeleteTask
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
