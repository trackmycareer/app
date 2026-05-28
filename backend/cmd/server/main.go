package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bhcloudlabs/trackmy-career/internal/admin"
	"github.com/bhcloudlabs/trackmy-career/internal/auth"
	"github.com/bhcloudlabs/trackmy-career/internal/certification"
	"github.com/bhcloudlabs/trackmy-career/internal/config"
	"github.com/bhcloudlabs/trackmy-career/internal/database"
	"github.com/bhcloudlabs/trackmy-career/internal/export"
	"github.com/bhcloudlabs/trackmy-career/internal/gamification"
	"github.com/bhcloudlabs/trackmy-career/internal/job"
	"github.com/bhcloudlabs/trackmy-career/internal/logger"
	"github.com/bhcloudlabs/trackmy-career/internal/middleware"
	"github.com/bhcloudlabs/trackmy-career/internal/profile"
	"github.com/bhcloudlabs/trackmy-career/internal/settings"
	"github.com/bhcloudlabs/trackmy-career/internal/skill"
	"github.com/bhcloudlabs/trackmy-career/internal/tag"
	"github.com/bhcloudlabs/trackmy-career/internal/user"
	"github.com/bhcloudlabs/trackmy-career/internal/win"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	logger.Init(cfg.Env)

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	// Seed admin user
	database.SeedAdmin(ctx, pool, cfg.AdminEmail, cfg.AdminPassword)

	// Repositories
	userRepo := user.NewRepository(pool)
	settingsRepo := settings.NewRepository(pool)
	tagRepo := tag.NewRepository(pool)
	winRepo := win.NewRepository(pool)
	jobRepo := job.NewRepository(pool)
	certRepo := certification.NewRepository(pool)
	skillRepo := skill.NewRepository(pool)
	gamificationRepo := gamification.NewRepository(pool)

	// Services
	adminService := admin.NewService(pool)
	exportService := export.NewService(pool)
	gamificationService := gamification.NewService(gamificationRepo)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	oauthManager := auth.NewOAuthManager(
		userRepo, settingsRepo, jwtManager,
		cfg.OAuthGoogleClientID, cfg.OAuthGoogleClientSecret,
		cfg.OAuthGitHubClientID, cfg.OAuthGitHubClientSecret,
		cfg.OAuthRedirectBase,
	)
	authService := auth.NewService(userRepo, jwtManager)
	winService := win.NewService(winRepo)
	jobService := job.NewService(jobRepo)
	certService := certification.NewService(certRepo)
	skillService := skill.NewService(skillRepo)

	// Handlers
	adminHandler := admin.NewHandler(adminService, userRepo)
	authHandler := auth.NewHandler(authService, oauthManager, settingsRepo, cfg.FrontendURL)
	exportHandler := export.NewHandler(exportService)
	userHandler := user.NewHandler(userRepo)
	settingsHandler := settings.NewHandler(settingsRepo)
	tagHandler := tag.NewHandler(tagRepo)
	winHandler := win.NewHandler(winService, gamificationService)
	jobHandler := job.NewHandler(jobService, gamificationService)
	certHandler := certification.NewHandler(certService, gamificationService)
	skillHandler := skill.NewHandler(skillService, gamificationService)
	gamificationHandler := gamification.NewHandler(gamificationService)
	profileHandler := profile.NewHandler(userRepo, gamificationRepo, winRepo, jobRepo, certRepo, skillRepo)

	// Router
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "healthy"})
	})

	// API v1
	v1 := router.Group("/api/v1")

	// Auth routes (public)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.GET("/providers", authHandler.Providers)
		authGroup.GET("/:provider", authHandler.OAuthInitiate)
		authGroup.GET("/:provider/callback", authHandler.OAuthCallback)
	}

	// Public profile route (no auth required)
	v1.GET("/profiles/:username", profileHandler.GetPublicProfile)

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(jwtManager))
	{
		// User
		protected.GET("/user/me", userHandler.GetMe)
		protected.PUT("/user/me", userHandler.UpdateMe)
		protected.PUT("/user/me/password", userHandler.ChangePassword)

		// Profile settings
		protected.GET("/user/me/profile", profileHandler.GetProfileSettings)
		protected.PUT("/user/me/profile", profileHandler.UpdateProfileSettings)

		// Tags
		protected.GET("/tags", tagHandler.List)
		protected.POST("/tags", tagHandler.Create)
		protected.PUT("/tags/:id", tagHandler.Update)
		protected.DELETE("/tags/:id", tagHandler.Delete)

		// Wins
		protected.GET("/wins", winHandler.List)
		protected.POST("/wins", winHandler.Create)
		protected.GET("/wins/:id", winHandler.Get)
		protected.PUT("/wins/:id", winHandler.Update)
		protected.DELETE("/wins/:id", winHandler.Delete)

		// Jobs
		protected.GET("/jobs", jobHandler.List)
		protected.POST("/jobs", jobHandler.Create)
		protected.GET("/jobs/:id", jobHandler.Get)
		protected.PUT("/jobs/:id", jobHandler.Update)
		protected.DELETE("/jobs/:id", jobHandler.Delete)

		// Certifications
		protected.GET("/certifications", certHandler.List)
		protected.POST("/certifications", certHandler.Create)
		protected.GET("/certifications/:id", certHandler.Get)
		protected.PUT("/certifications/:id", certHandler.Update)
		protected.PATCH("/certifications/:id/status", certHandler.UpdateStatus)
		protected.DELETE("/certifications/:id", certHandler.Delete)

		// Skills
		protected.GET("/skills", skillHandler.List)
		protected.POST("/skills", skillHandler.Create)
		protected.GET("/skills/:id", skillHandler.Get)
		protected.PUT("/skills/:id", skillHandler.Update)
		protected.DELETE("/skills/:id", skillHandler.Delete)
		protected.POST("/skills/:id/evidence", skillHandler.AddEvidence)
		protected.DELETE("/skills/:id/evidence/:evidenceId", skillHandler.RemoveEvidence)

		// Export
		exportGroup := protected.Group("/export")
		{
			exportGroup.GET("/json", exportHandler.ExportJSON)
			exportGroup.GET("/markdown", exportHandler.ExportMarkdown)
		}

		// Gamification
		gamificationGroup := protected.Group("/gamification")
		{
			gamificationGroup.GET("/progress", gamificationHandler.GetProgress)
			gamificationGroup.GET("/badges", gamificationHandler.ListBadges)
			gamificationGroup.GET("/heatmap", gamificationHandler.GetHeatmap)
			gamificationGroup.GET("/streak", gamificationHandler.GetStreak)
		}
	}

	// Admin routes
	adminGroup := protected.Group("/admin")
	adminGroup.Use(middleware.AdminRequired())
	{
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.GET("/users/:id", adminHandler.GetUser)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
		adminGroup.PATCH("/users/:id/admin", adminHandler.ToggleAdmin)
		adminGroup.GET("/stats", adminHandler.GetStats)
		adminGroup.GET("/settings", settingsHandler.GetSettings)
		adminGroup.PUT("/settings", settingsHandler.UpdateSettings)

		// Badge management
		adminGroup.GET("/badges", gamificationHandler.AdminListBadges)
		adminGroup.POST("/badges", gamificationHandler.AdminCreateBadge)
		adminGroup.PUT("/badges/:id", gamificationHandler.AdminUpdateBadge)
		adminGroup.DELETE("/badges/:id", gamificationHandler.AdminDeleteBadge)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
