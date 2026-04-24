package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"radimaster/internal/models"
	"radimaster/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClassroomHandler struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewClassroomHandler(db *gorm.DB, hub *ws.Hub) *ClassroomHandler {
	return &ClassroomHandler{db: db, hub: hub}
}

// Create starts a new live classroom session for a medical case.
func (h *ClassroomHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		CaseID uint   `json:"case_id" binding:"required"`
		Title  string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user is a doctor
	var user models.User
	if h.db.First(&user, userID).Error != nil || user.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only doctors can create classrooms"})
		return
	}

	// Verify case exists
	var mc models.MedicalCase
	if h.db.First(&mc, req.CaseID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	classroom := models.Classroom{
		CaseID:   req.CaseID,
		DoctorID: userID,
		Title:    req.Title,
		Code:     generateCode(),
		IsActive: true,
	}
	h.db.Create(&classroom)
	h.db.Preload("Case").Preload("Doctor").First(&classroom, classroom.ID)

	c.JSON(http.StatusCreated, classroom)
}

// Join returns classroom details for students to join.
func (h *ClassroomHandler) Join(c *gin.Context) {
	code := c.Param("code")

	var classroom models.Classroom
	if err := h.db.Where("code = ? AND is_active = ?", code, true).
		Preload("Case.Phases", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Preload("Case.Phases.Slices", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Preload("Case.Doctor").
		Preload("Doctor").
		First(&classroom).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or inactive"})
		return
	}

	// Get online users
	roomID := fmt.Sprintf("classroom:%s", code)
	onlineUsers := h.hub.GetOnlineUsers(roomID)

	c.JSON(http.StatusOK, gin.H{
		"classroom":    classroom,
		"online_users": onlineUsers,
	})
}

// List returns active classrooms.
func (h *ClassroomHandler) List(c *gin.Context) {
	var classrooms []models.Classroom
	h.db.Where("is_active = ?", true).
		Preload("Case").
		Preload("Doctor").
		Order("created_at DESC").
		Find(&classrooms)

	c.JSON(http.StatusOK, classrooms)
}

// End deactivates a classroom session.
func (h *ClassroomHandler) End(c *gin.Context) {
	code := c.Param("code")
	userID := c.GetUint("userID")

	var classroom models.Classroom
	if h.db.Where("code = ? AND doctor_id = ?", code, userID).First(&classroom).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	h.db.Model(&classroom).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "Classroom ended"})
}

func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return string(code)
}
