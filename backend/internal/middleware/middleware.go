package middleware

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORS handles Cross-Origin Resource Sharing headers.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger logs each HTTP request with latency, status, and path.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		log.Printf("│ %3d │ %13v │ %-7s │ %s", status, latency, method, path)
	}
}

// Auth validates JWT tokens from the Authorization header and sets userID in context.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		c.Set("userID", uint(userID))
		c.Next()
	}
}

// RateLimit provides simple in-memory rate limiting per IP.
// For production, use Redis-backed rate limiting.
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	type clientInfo struct {
		count    int
		lastSeen time.Time
	}
	var mu sync.Mutex
	clients := make(map[string]*clientInfo)

	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			now := time.Now()
			for ip, cl := range clients {
				if now.Sub(cl.lastSeen) > window*2 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		if cl, exists := clients[ip]; exists {
			if now.Sub(cl.lastSeen) > window {
				cl.count = 1
				cl.lastSeen = now
			} else {
				cl.count++
				if cl.count > maxRequests {
					mu.Unlock()
					c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
					c.Abort()
					return
				}
			}
		} else {
			clients[ip] = &clientInfo{count: 1, lastSeen: now}
		}
		mu.Unlock()

		c.Next()
	}
}
