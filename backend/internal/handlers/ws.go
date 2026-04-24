package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"radimaster/internal/config"
	"radimaster/internal/models"
	"radimaster/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WSHandler struct {
	hub *ws.Hub
	cfg *config.Config
	db  *gorm.DB
}

func NewWSHandler(hub *ws.Hub, cfg *config.Config, db *gorm.DB) *WSHandler {
	return &WSHandler{hub: hub, cfg: cfg, db: db}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (h *WSHandler) HandleClassroomWS(c *gin.Context) {
	code := c.Param("code")

	var classroom models.Classroom
	if h.db.Where("code = ? AND is_active = ?", code, true).First(&classroom).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	tokenStr := c.Query("token")
	if tokenStr == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.User
	h.db.First(&user, userID)
	userName := user.Name
	if userName == "" {
		userName = fmt.Sprintf("User %d", userID)
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	roomID := fmt.Sprintf("classroom:%s", code)
	client := ws.NewClient(h.hub, conn, roomID, userID, userName, classroom.ID)
	h.hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
