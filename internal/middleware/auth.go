package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ccany/internal/crypto"
	"ccany/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware authentication middleware
type AuthMiddleware struct {
	db            *database.Database
	jwtSecret     []byte
	logger        *logrus.Logger
	cryptoService *crypto.CryptoService
}

// NewAuthMiddleware creates authentication middleware
func NewAuthMiddleware(db *database.Database, jwtSecret string, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		db:            db,
		jwtSecret:     []byte(jwtSecret),
		logger:        logger,
		cryptoService: db.CryptoService,
	}
}

// JWTClaims JWT payload
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthRequired middleware that requires authentication
func (am *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or invalid authorization token",
			})
			c.Abort()
			return
		}

		claims, err := am.validateToken(token)
		if err != nil {
			am.logger.WithError(err).Error("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// verify user exists
		user, err := am.db.Client.User.Get(context.Background(), claims.UserID)
		if err != nil {
			am.logger.WithError(err).WithField("user_id", claims.UserID).Error("User not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// check if user is active
		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User account is disabled",
			})
			c.Abort()
			return
		}

		// store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user", user)

		c.Next()
	}
}

// AdminRequired middleware that requires admin privileges
func (am *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// first check if authenticated
		if _, exists := c.Get("user_id"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// check user role
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts token from request
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	// get from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// get from Cookie
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}

	// get from query parameter
	return c.Query("token")
}

// validateToken validates JWT token
func (am *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GenerateToken generates JWT token
func (am *AuthMiddleware) GenerateToken(userID string, username, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ccany",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.jwtSecret)
}

// RefreshToken refreshes JWT token
func (am *AuthMiddleware) RefreshToken(tokenString string) (string, error) {
	claims, err := am.validateToken(tokenString)
	if err != nil {
		return "", err
	}

	// check if token is about to expire (remaining time less than 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return tokenString, nil // no need to refresh
	}

	// generate new token
	return am.GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// RateLimitMiddleware simple rate limiting middleware
func (am *AuthMiddleware) RateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
	// simple memory storage, production should use Redis
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// clean up expired records
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < duration {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// check if limit exceeded
		if len(clients[clientIP]) >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// record current request
		clients[clientIP] = append(clients[clientIP], now)

		c.Next()
	}
}

// GetCurrentUser gets current user information
func GetCurrentUser(c *gin.Context) (userID string, username string, role string, ok bool) {
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	} else {
		return "", "", "", false
	}

	if uname, exists := c.Get("username"); exists {
		username = uname.(string)
	}

	if urole, exists := c.Get("user_role"); exists {
		role = urole.(string)
	}

	return userID, username, role, true
}
