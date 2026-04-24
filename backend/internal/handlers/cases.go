package handlers

import (
	"net/http"
	"strconv"

	"radimaster/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CasesHandler struct {
	db *gorm.DB
}

func NewCasesHandler(db *gorm.DB) *CasesHandler {
	return &CasesHandler{db: db}
}

// List returns all public cases with optional filtering by modality and search.
func (h *CasesHandler) List(c *gin.Context) {
	modality := c.Query("modality")
	search := c.Query("search")

	query := h.db.Model(&models.MedicalCase{}).Where("is_public = ?", true)

	if modality != "" {
		query = query.Where("modality = ?", modality)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR diagnosis LIKE ? OR body_part LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var cases []models.MedicalCase
	query.Preload("Doctor").
		Preload("Phases", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Order("created_at DESC").
		Find(&cases)

	// Calculate slice counts per phase
	for i := range cases {
		for j := range cases[i].Phases {
			var count int64
			h.db.Model(&models.Slice{}).Where("phase_id = ?", cases[i].Phases[j].ID).Count(&count)
			cases[i].Phases[j].SliceCount = int(count)
		}
	}

	c.JSON(http.StatusOK, cases)
}

// Get returns a single case with all phases and slice metadata.
func (h *CasesHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var mc models.MedicalCase
	if err := h.db.
		Preload("Doctor").
		Preload("Phases", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Preload("Phases.Slices", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		First(&mc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Increment view count
	h.db.Model(&mc).Update("view_count", gorm.Expr("view_count + 1"))

	c.JSON(http.StatusOK, mc)
}

// ListModalities returns distinct modalities for filtering.
func (h *CasesHandler) ListModalities(c *gin.Context) {
	var modalities []string
	h.db.Model(&models.MedicalCase{}).
		Where("is_public = ?", true).
		Distinct("modality").
		Pluck("modality", &modalities)

	c.JSON(http.StatusOK, modalities)
}
