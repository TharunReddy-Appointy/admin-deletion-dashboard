package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	appointyDomain = "@appointy.com"
)

// Config holds the OAuth and JWT configuration
type Config struct {
	OAuth2Config *oauth2.Config
	JWTSecret    []byte
	RedirectURL  string
}

// Claims represents JWT claims
type Claims struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	jwt.RegisteredClaims
}

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"` // Hosted domain
}

// NewAuthConfig creates a new auth configuration
func NewAuthConfig(clientID, clientSecret, redirectURL, jwtSecret string) *Config {
	return &Config{
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		JWTSecret:   []byte(jwtSecret),
		RedirectURL: redirectURL,
	}
}

// GetLoginURL generates the OAuth2 login URL
func (c *Config) GetLoginURL(state string) string {
	return c.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for a token
func (c *Config) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.OAuth2Config.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from Google
func (c *Config) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := c.OAuth2Config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// ValidateAppointyEmail checks if the email is from appointy.com domain
func ValidateAppointyEmail(email string) error {
	if !strings.HasSuffix(strings.ToLower(email), appointyDomain) {
		return errors.New("only @appointy.com emails are allowed")
	}
	return nil
}

// GenerateJWT generates a JWT token for the authenticated user
func (c *Config) GenerateJWT(email, name, picture string) (string, error) {
	claims := Claims{
		Email:   email,
		Name:    name,
		Picture: picture,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "admin-deletion-dashboard",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.JWTSecret)
}

// ValidateJWT validates and parses a JWT token
func (c *Config) ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Validate that it's still an appointy email
		if err := ValidateAppointyEmail(claims.Email); err != nil {
			return nil, err
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthMiddleware is a Gin middleware that validates JWT tokens
func (c *Config) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract token from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			ctx.Abort()
			return
		}

		// Remove "Bearer " prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			ctx.Abort()
			return
		}

		// Validate token
		claims, err := c.ValidateJWT(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}

		// Set user info in context
		ctx.Set("user_email", claims.Email)
		ctx.Set("user_name", claims.Name)
		ctx.Next()
	}
}

// GetUserEmailFromContext retrieves the authenticated user's email from context
func GetUserEmailFromContext(ctx *gin.Context) (string, error) {
	email, exists := ctx.Get("user_email")
	if !exists {
		return "", errors.New("user email not found in context")
	}
	emailStr, ok := email.(string)
	if !ok {
		return "", errors.New("invalid user email format")
	}
	return emailStr, nil
}
