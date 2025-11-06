package handler

import (
	"crypto/rand"
	"encoding/base64"
	//"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.appointy.com/admin-deletion-dashboard/internal/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authConfig *auth.Config
	sessions   map[string]string // state -> used to prevent CSRF (in production, use Redis)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authConfig *auth.Config) *AuthHandler {
	return &AuthHandler{
		authConfig: authConfig,
		sessions:   make(map[string]string),
	}
}

// HandleLogin initiates the OAuth2 login flow
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	// Generate random state
	state, err := generateRandomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	// Store state (in production, store in Redis with expiration)
	h.sessions[state] = "pending"

	// Get OAuth2 URL
	url := h.authConfig.GetLoginURL(state)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// HandleCallback handles the OAuth2 callback
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	// Get authorization code and state
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	// Validate state (CSRF protection)
	if _, exists := h.sessions[state]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}
	delete(h.sessions, state) // Remove used state

	// Exchange code for token
	token, err := h.authConfig.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
		return
	}

	// Get user info
	userInfo, err := h.authConfig.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Validate email domain
	if err := auth.ValidateAppointyEmail(userInfo.Email); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "only @appointy.com emails are allowed"})
		return
	}

	// Verify email is verified
	if !userInfo.VerifiedEmail {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	// Generate JWT
	jwtToken, err := h.authConfig.GenerateJWT(userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"email":   userInfo.Email,
			"name":    userInfo.Name,
			"picture": userInfo.Picture,
		},
	})
}

// HandleMe returns the current user's info
func (h *AuthHandler) HandleMe(c *gin.Context) {
	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	name, _ := c.Get("user_name")

	c.JSON(http.StatusOK, gin.H{
		"email": email,
		"name":  name,
	})
}

// generateRandomState generates a random state string for CSRF protection
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleLogout logs out the user (client should delete the token)
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}
