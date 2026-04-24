package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"radimaster/internal/config"
	"radimaster/internal/database"
	"radimaster/internal/handlers"
	"radimaster/internal/middleware"
	"radimaster/internal/services"
	"radimaster/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	db := database.Connect(cfg)
	services.Seed(db, cfg.MediaDir)

	hub := ws.NewHub(cfg.RedisURL)
	go hub.Run()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health (no rate limit)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "radimaster", "version": "2.0.0"})
	})

	// Static medical images — no rate limit (CT scans can have 300+ slices)
	r.Static("/media", cfg.MediaDir)

	// WebSocket — no rate limit (persistent connection)
	wsH := handlers.NewWSHandler(hub, cfg, db)
	r.GET("/ws/classrooms/:code", wsH.HandleClassroomWS)

	// API routes — rate limited
	api := r.Group("/api")
	api.Use(middleware.RateLimit(200, time.Minute))

	// Auth
	authH := handlers.NewAuthHandler(db, cfg)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.GET("/me", middleware.Auth(cfg.JWTSecret), authH.Me)
	}

	// Cases (public)
	casesH := handlers.NewCasesHandler(db)
	api.GET("/cases", casesH.List)
	api.GET("/cases/modalities", casesH.ListModalities)
	api.GET("/cases/:id", casesH.Get)

	// Classrooms (protected)
	classroomH := handlers.NewClassroomHandler(db, hub)
	cr := api.Group("/classrooms", middleware.Auth(cfg.JWTSecret))
	{
		cr.POST("", classroomH.Create)
		cr.GET("", classroomH.List)
		cr.GET("/:code", classroomH.Join)
		cr.PUT("/:code/end", classroomH.End)
	}

	// Admin metrics (protected)
	metricsH := handlers.NewMetricsHandler(db, hub)
	api.GET("/admin/metrics", middleware.Auth(cfg.JWTSecret), metricsH.GetMetrics)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 RadiMaster V2 starting on http://localhost:%s", cfg.Port)
		log.Printf("📡 WebSocket: ws://localhost:%s/ws/classrooms/:code", cfg.Port)
		log.Printf("🏥 Media: http://localhost:%s/media/", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⏳ Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Forced shutdown: %v", err)
	}
	log.Println("✅ Server stopped")
}
