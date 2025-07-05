package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const cookieName = "user_id"

func signUserID(userID string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

func validateCookie(userID, signature string, secretKey string) bool {
	return hmac.Equal([]byte(signUserID(userID, secretKey)), []byte(signature))
}

func SetAuthCookie(c *gin.Context, userID uuid.UUID, secretKey string) {
	signed := userID.String() + "|" + signUserID(userID.String(), secretKey)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "user_id",
		Value:    signed,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	})
}

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(cookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.Split(cookie.Value, "|")
		if len(parts) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, sig := parts[0], parts[1]
		if !validateCookie(userID, sig, secretKey) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
