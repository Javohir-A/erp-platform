package middleware

import (
	"net/http"
	"strings"

	authv1 "erp/platform/genproto/erp/auth/v1"
	"erp/platform/gateway/services"

	"github.com/gin-gonic/gin"
)

const CtxUserID = "user_id"
const CtxRole = "role"

func RequireAuth(cl *services.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tok := strings.TrimSpace(h[7:])
		resp, err := cl.Auth.ValidateToken(c.Request.Context(), &authv1.ValidateTokenRequest{Token: tok})
		if err != nil || !resp.GetValid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxUserID, resp.GetUserId())
		c.Set(CtxRole, resp.GetRole())
		c.Next()
	}
}
