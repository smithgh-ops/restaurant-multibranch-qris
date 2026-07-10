package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/auth"
)

// contextKey type prevents collisions in gin.Keys map.
const (
	KeyUserID         = "user_id"
	KeyOrganizationID = "organization_id"
	KeyEmail          = "email"
	KeyUserName       = "user_name"
)

// Authenticate returns a middleware that validates the JWT access token from
// the Authorization: ****** header and populates gin context keys.
func Authenticate(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak ditemukan"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseAccessToken(jwtSecret, tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid atau sudah kadaluarsa"})
			return
		}

		c.Set(KeyUserID, claims.UserID)
		c.Set(KeyOrganizationID, claims.OrganizationID)
		c.Set(KeyEmail, claims.Email)
		c.Set(KeyUserName, claims.Name)
		c.Next()
	}
}

// RequireRoles returns a middleware that checks whether the authenticated user
// has at least one of the required roles across any branch (or globally).
// It relies on the roles being pre-loaded into the context; callers that need
// branch-scoped checks should use HasBranchRole instead.
//
// NOTE: For routes that require database role checks, inject the auth repository
// into the handler and perform the check there. This middleware is a lightweight
// shortcut for simple global role gates.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		// Roles are stored as []string in context when populated by a richer middleware.
		raw, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses ditolak"})
			return
		}
		roles, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses ditolak"})
			return
		}
		for _, r := range roles {
			if _, ok := allowedSet[r]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses ditolak"})
	}
}

// OrgIDFromContext extracts the organization_id set by the Authenticate middleware.
func OrgIDFromContext(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(KeyOrganizationID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

// UserIDFromContext extracts the user_id set by the Authenticate middleware.
func UserIDFromContext(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(KeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
