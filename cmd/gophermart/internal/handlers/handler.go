package handlers

import (
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

func NewURLHandler(service *services.Service, secretKey string) *Handler {
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
