package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/admin"
	"github.com/trackmycareer/app/internal/auth"
	"github.com/trackmycareer/app/internal/cloudflare"
	"github.com/trackmycareer/app/internal/customdomain"
	"github.com/trackmycareer/app/internal/mfa"
	"github.com/trackmycareer/app/internal/mfa/backup"
	mfacrypto "github.com/trackmycareer/app/internal/mfa/crypto"
	"github.com/trackmycareer/app/internal/mfa/passkey"
	"github.com/trackmycareer/app/internal/mfa/totp"
	"github.com/trackmycareer/app/internal/certification"
	"github.com/trackmycareer/app/internal/company"
	"github.com/trackmycareer/app/internal/config"
	"github.com/trackmycareer/app/internal/database"
	"github.com/trackmycareer/app/internal/export"
	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/internal/importer"
	"github.com/trackmycareer/app/internal/job"
	"github.com/trackmycareer/app/internal/jobtitle"
	"github.com/trackmycareer/app/internal/linkedaccount"
	"github.com/trackmycareer/app/internal/location"
	"github.com/trackmycareer/app/internal/logger"
	"github.com/trackmycareer/app/internal/mailer"
	"github.com/trackmycareer/app/internal/middleware"
	"github.com/trackmycareer/app/internal/passwordreset"
	"github.com/trackmycareer/app/internal/polar"
	"github.com/trackmycareer/app/internal/profile"
	"github.com/trackmycareer/app/internal/settings"
	"github.com/trackmycareer/app/internal/skill"
	"github.com/trackmycareer/app/internal/storage"
	"github.com/trackmycareer/app/internal/tag"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/internal/verification"
	"github.com/trackmycareer/app/internal/win"
	"github.com/trackmycareer/app/pkg/response"

	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if len(cfg.JWTSecret) < 32 {
		if cfg.Env == "production" {
			log.Fatal("JWT_SECRET must be at least 32 characters; generate one with: openssl rand -base64 48")
		}
		slog.Warn("JWT_SECRET is shorter than 32 characters; this is insecure for production")
	}

	// MFA encryption key
	var mfaEncryptor *mfacrypto.Encryptor
	mfaKey, err := cfg.MFAKeyBytes()
	if err != nil {
		if cfg.Env == "production" {
			log.Fatalf("MFA_ENCRYPTION_KEY: %v", err)
		}
		slog.Warn("MFA features unavailable", "reason", err.Error())
	} else {
		mfaEncryptor, err = mfacrypto.NewEncryptor(mfaKey)
		if err != nil {
			log.Fatalf("initialising MFA encryptor: %v", err)
		}
	}

	logger.Init(cfg.Env)

	storageClient, err := storage.New(
		cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey,
		cfg.S3Bucket, cfg.S3PublicURL, cfg.S3Region, cfg.S3UseSSL,
	)
	if err != nil {
		log.Fatalf("initialising S3 storage: %v", err)
	}

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
	companyRepo := company.NewRepository(pool)
	jobtitleRepo := jobtitle.NewRepository(pool)
	locationRepo := location.NewRepository(pool)
	certRepo := certification.NewRepository(pool)
	skillRepo := skill.NewRepository(pool)
	gamificationRepo := gamification.NewRepository(pool)
	linkedAccountRepo := linkedaccount.NewRepository(pool)
	customDomainRepo := customdomain.NewRepository(pool)
	verificationRepo := verification.NewRepository(pool)

	// Services
	adminService := admin.NewService(pool)
	exportService := export.NewService(pool)
	gamificationService := gamification.NewService(gamificationRepo)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	oauthManager := auth.NewOAuthManager(
		userRepo, settingsRepo, jwtManager,
		cfg.OAuthGoogleClientID, cfg.OAuthGoogleClientSecret,
		cfg.OAuthGitHubClientID, cfg.OAuthGitHubClientSecret,
		cfg.FrontendURL,
	)
	authService := auth.NewService(userRepo, jwtManager)
	mailService := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom, cfg.FrontendURL)
	verificationService := verification.NewService(verificationRepo, userRepo, mailService)
	passwordResetRepo := passwordreset.NewRepository(pool)

	// MFA services
	totpRepo := totp.NewRepository(pool)
	passkeyRepo := passkey.NewRepository(pool)
	backupRepo := backup.NewRepository(pool)

	var mfaService *mfa.Service
	var mfaHandler *mfa.Handler
	if mfaEncryptor != nil {
		frontendParsed, parseErr := url.Parse(cfg.FrontendURL)
		if parseErr != nil {
			log.Fatalf("parsing FRONTEND_URL: %v", parseErr)
		}
		rpID := frontendParsed.Hostname()

		totpService := totp.NewService(totpRepo, mfaEncryptor, "trackmy.career")
		passkeyService, passkeyErr := passkey.NewService(passkeyRepo, rpID, "trackmy.career", cfg.FrontendURL)
		if passkeyErr != nil {
			log.Fatalf("initialising passkey service: %v", passkeyErr)
		}
		backupService := backup.NewService(backupRepo, mfaKey)

		mfaService = mfa.NewService(totpService, passkeyService, backupService, userRepo)
		mfaHandler = mfa.NewHandler(mfaService, userRepo)
	}

	// Password reset service (after MFA so we can wire MFA verification)
	var mfaCodeVerifier passwordreset.MFACodeVerifier
	var passkeyChallengeFunc passwordreset.PasskeyChallengeFunc
	if mfaService != nil {
		mfaCodeVerifier = func(ctx context.Context, userID uuid.UUID, method, code string) error {
			return mfaService.VerifyAnyMethod(ctx, userID, mfa.Method(method), code)
		}
		passkeyChallengeFunc = func(ctx context.Context, userID uuid.UUID) (any, error) {
			passkeys, err := mfaService.Passkey().ListByUserID(ctx, userID)
			if err != nil {
				return nil, err
			}
			return mfaService.Passkey().BeginAuthentication(ctx, userID, passkeys)
		}
	}
	passwordResetService := passwordreset.NewService(passwordResetRepo, userRepo, mailService, cfg.FrontendURL, mfaCodeVerifier)

	winService := win.NewService(winRepo)
	jobService := job.NewService(jobRepo)
	certService := certification.NewService(certRepo)
	skillService := skill.NewService(skillRepo)
	companyCache := gocache.New(5*time.Minute, 10*time.Minute)
	companyClient := company.NewCompaniesHouseClient(cfg.CompaniesHouseAPIKey, cfg.CompaniesHouseBaseURL)
	companyService := company.NewService(companyRepo, companyCache, companyClient)
	jobtitleService := jobtitle.NewService(jobtitleRepo)
	locationCache := gocache.New(5*time.Minute, 10*time.Minute)
	photonClient := location.NewPhotonClient(cfg.PhotonBaseURL)
	locationService := location.NewService(locationRepo, locationCache, photonClient)
	linkedAccountService := linkedaccount.NewService(linkedAccountRepo)

	// Cloudflare for SaaS (custom domains)
	var cfClient *cloudflare.Client
	if cfg.CloudflareEnabled() {
		cfClient = cloudflare.NewClient(cfg.CloudflareAPIToken, cfg.CloudflareZoneID, cfg.CloudflareCustomFallback)
	}
	customDomainService := customdomain.NewService(customDomainRepo, cfClient, cfg.CloudflareCustomFallback)

	// Import
	importerService := importer.NewService(jobRepo, certRepo, skillRepo, winRepo, userRepo, gamificationService)

	// Handlers
	importerHandler := importer.NewHandler(importerService)
	adminHandler := admin.NewHandler(adminService, userRepo)
	authHandler := auth.NewHandler(authService, oauthManager, settingsRepo, cfg.FrontendURL, verificationService, mfaService)
	exportHandler := export.NewHandler(exportService)
	userHandler := user.NewHandler(userRepo, storageClient, customDomainService) // customDomainService implements user.DomainCleaner
	settingsHandler := settings.NewHandler(settingsRepo)
	tagHandler := tag.NewHandler(tagRepo)
	winHandler := win.NewHandler(winService, gamificationService)
	jobHandler := job.NewHandler(jobService, gamificationService)
	certHandler := certification.NewHandler(certService, gamificationService)
	skillHandler := skill.NewHandler(skillService, gamificationService)
	companyHandler := company.NewHandler(companyService)
	jobtitleHandler := jobtitle.NewHandler(jobtitleService)
	locationHandler := location.NewHandler(locationService)
	gamificationHandler := gamification.NewHandler(gamificationService)
	linkedAccountHandler := linkedaccount.NewHandler(
		linkedAccountService, cfg.FrontendURL,
		cfg.OAuthLinkedInClientID, cfg.OAuthLinkedInClientSecret,
		cfg.OAuthGitHubClientID, cfg.OAuthGitHubClientSecret,
		cfg.JWTSecret,
	)
	verificationHandler := verification.NewHandler(verificationService)
	passwordResetHandler := passwordreset.NewHandler(passwordResetService, passkeyChallengeFunc)
	customDomainHandler := customdomain.NewHandler(customDomainService, userRepo)
	profileHandler := profile.NewHandler(userRepo, gamificationRepo, winRepo, jobRepo, certRepo, skillRepo, linkedAccountService, customDomainRepo)

	// Router
	router := gin.New()

	// Configure trusted proxies for accurate client IP detection.
	// When TRUSTED_PROXIES is empty (default), trust no proxies so
	// c.ClientIP() returns the direct connection address.
	trustedProxies := cfg.TrustedProxyList()
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		slog.Warn("failed to set trusted proxies", "error", err)
	}

	router.MaxMultipartMemory = 10 << 20 // 10 MB for multipart uploads
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		if c.ContentType() == "application/json" {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		}
		c.Next()
	})
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.SecurityHeaders())

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "healthy"})
	})

	// API v1
	v1 := router.Group("/api/v1")

	// Auth routes (public, with per-route IP rate limiting)
	authGroup := v1.Group("/auth")
	{
		// Login: 5 req/min per IP, burst 10
		authGroup.POST("/login", middleware.IPRateLimit(rate.Limit(5.0/60.0), 10), authHandler.Login)
		// Register: 3 req/min per IP, burst 5
		authGroup.POST("/register", middleware.IPRateLimit(rate.Limit(3.0/60.0), 5), authHandler.Register)
		// Refresh: 10 req/min per IP, burst 20
		authGroup.POST("/refresh", middleware.IPRateLimit(rate.Limit(10.0/60.0), 20), authHandler.Refresh)

		authGroup.GET("/providers", authHandler.Providers)
		authGroup.POST("/verify-email/confirm", verificationHandler.VerifyEmail)
		// Password reset: 3 req/min per IP, burst 5
		authGroup.POST("/forgot-password", middleware.IPRateLimit(rate.Limit(3.0/60.0), 5), passwordResetHandler.ForgotPassword)
		// Reset password: 5 req/min per IP, burst 10
		authGroup.POST("/reset-password", middleware.IPRateLimit(rate.Limit(5.0/60.0), 10), passwordResetHandler.ResetPassword)
		// Reset password passkey challenge: 5 req/min per IP, burst 10
		authGroup.POST("/reset-password/passkey-challenge", middleware.IPRateLimit(rate.Limit(5.0/60.0), 10), passwordResetHandler.PasskeyChallenge)
		// MFA verification during login: 5 req/min per IP, burst 10
		authGroup.POST("/mfa/verify", middleware.IPRateLimit(rate.Limit(5.0/60.0), 10), authHandler.VerifyMFA)
		// MFA passkey challenge during login: 5 req/min per IP, burst 10
		authGroup.POST("/mfa/passkey/challenge", middleware.IPRateLimit(rate.Limit(5.0/60.0), 10), authHandler.MFAPasskeyChallenge)
		authGroup.GET("/:provider", authHandler.OAuthInitiate)
		authGroup.POST("/:provider/callback", authHandler.OAuthCallback)
	}

	// Public email change confirmation (no auth, user clicks link from email)
	v1.POST("/user/me/email/confirm", verificationHandler.ConfirmEmailChange)

	// Public profile routes (no auth required)
	v1.GET("/profiles/:username", profileHandler.GetPublicProfile)
	v1.GET("/profiles/by-domain/:domain", middleware.PublicCORS(), middleware.IPRateLimit(rate.Limit(30.0/60.0), 60), profileHandler.GetPublicProfileByDomain)

	// Public config (feature flags)
	v1.GET("/config", polar.GetConfigHandler(cfg))

	// Linked account OAuth callbacks (public, browser redirects from providers)
	v1.POST("/linked-accounts/link/:provider/callback", linkedAccountHandler.HandleCallback)

	// Protected routes (authenticated, but no email verification required)
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(jwtManager))
	{
		// These routes must work for unverified users
		protected.GET("/user/me", userHandler.GetMe)
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/verify-email/send", verificationHandler.SendVerification)
	}

	// Protected routes requiring verified email
	verified := v1.Group("")
	verified.Use(middleware.AuthRequired(jwtManager))
	verified.Use(middleware.EmailVerifiedRequired())
	{
		// User
		verified.PUT("/user/me", userHandler.UpdateMe)
		verified.PUT("/user/me/password", userHandler.ChangePassword)
		verified.POST("/user/me/avatar", userHandler.UploadAvatar)
		verified.DELETE("/user/me/avatar", userHandler.DeleteAvatar)
		verified.POST("/user/me/delete", userHandler.DeleteAccount)
		verified.PUT("/user/me/newsletter", userHandler.UpdateNewsletter)

		// Email change
		verified.POST("/user/me/email/change", verificationHandler.RequestEmailChange)

		// Profile settings
		verified.GET("/user/me/profile", profileHandler.GetProfileSettings)
		verified.PUT("/user/me/profile", profileHandler.UpdateProfileSettings)

		// Tags
		verified.GET("/tags", tagHandler.List)
		verified.POST("/tags", tagHandler.Create)
		verified.PUT("/tags/:id", tagHandler.Update)
		verified.DELETE("/tags/:id", tagHandler.Delete)

		// Wins
		verified.GET("/wins", winHandler.List)
		verified.POST("/wins", winHandler.Create)
		verified.GET("/wins/:id", winHandler.Get)
		verified.PUT("/wins/:id", winHandler.Update)
		verified.DELETE("/wins/:id", winHandler.Delete)

		// Jobs
		verified.GET("/jobs", jobHandler.List)
		verified.POST("/jobs", jobHandler.Create)
		verified.GET("/jobs/:id", jobHandler.Get)
		verified.PUT("/jobs/:id", jobHandler.Update)
		verified.DELETE("/jobs/:id", jobHandler.Delete)

		// Autocomplete search endpoints (rate-limited: 10 req/s per user, burst 20)
		searchGroup := verified.Group("")
		searchGroup.Use(middleware.SearchRateLimit(10, 20))
		{
			searchGroup.GET("/companies/search", companyHandler.Search)
			searchGroup.GET("/jobtitles/search", jobtitleHandler.Search)
			searchGroup.GET("/locations/search", locationHandler.Search)
			searchGroup.GET("/certifications/search", certHandler.Search)
			searchGroup.GET("/skills/search", skillHandler.Search)
		}

		// Certifications
		verified.GET("/certifications", certHandler.List)
		verified.POST("/certifications", certHandler.Create)
		verified.GET("/certifications/:id", certHandler.Get)
		verified.PUT("/certifications/:id", certHandler.Update)
		verified.PATCH("/certifications/:id/status", certHandler.UpdateStatus)
		verified.DELETE("/certifications/:id", certHandler.Delete)

		// Skills
		verified.GET("/skills", skillHandler.List)
		verified.POST("/skills", skillHandler.Create)
		verified.GET("/skills/:id", skillHandler.Get)
		verified.PUT("/skills/:id", skillHandler.Update)
		verified.DELETE("/skills/:id", skillHandler.Delete)
		verified.POST("/skills/:id/evidence", skillHandler.AddEvidence)
		verified.DELETE("/skills/:id/evidence/:evidenceId", skillHandler.RemoveEvidence)

		// Export
		exportGroup := verified.Group("/export")
		{
			exportGroup.GET("/json", exportHandler.ExportJSON)
			exportGroup.GET("/markdown", exportHandler.ExportMarkdown)
		}

		// Import
		importGroup := verified.Group("/import")
		{
			importGroup.POST("/preview", importerHandler.Preview)
			importGroup.POST("/confirm", importerHandler.Confirm)
			importGroup.GET("/templates/:type", importerHandler.Template)
		}

		// Gamification
		gamificationGroup := verified.Group("/gamification")
		{
			gamificationGroup.GET("/progress", gamificationHandler.GetProgress)
			gamificationGroup.GET("/badges", gamificationHandler.ListBadges)
			gamificationGroup.GET("/heatmap", gamificationHandler.GetHeatmap)
			gamificationGroup.GET("/streak", gamificationHandler.GetStreak)
		}

		// Linked accounts
		linkedGroup := verified.Group("/linked-accounts")
		{
			linkedGroup.GET("", linkedAccountHandler.List)
			linkedGroup.GET("/link/:provider", linkedAccountHandler.InitiateLink)
			linkedGroup.POST("/website", linkedAccountHandler.AddWebsite)
			linkedGroup.POST("/website/verify", linkedAccountHandler.VerifyWebsite)
			linkedGroup.DELETE("/:provider", linkedAccountHandler.Unlink)
		}

		// Custom domains (rate-limited: create 3/min, verify 10/min)
		customDomainGroup := verified.Group("/custom-domain")
		{
			customDomainGroup.POST("", middleware.IPRateLimit(rate.Limit(3.0/60.0), 5), customDomainHandler.Create)
			customDomainGroup.GET("", customDomainHandler.Get)
			customDomainGroup.DELETE("", customDomainHandler.Delete)
			customDomainGroup.POST("/verify", middleware.IPRateLimit(rate.Limit(10.0/60.0), 20), customDomainHandler.Verify)
			customDomainGroup.PUT("/theme", customDomainHandler.UpdateTheme)
		}

		// MFA management
		if mfaHandler != nil {
			mfaGroup := verified.Group("/user/me/mfa")
			{
				mfaGroup.GET("/status", mfaHandler.GetMFAStatus)
				mfaGroup.POST("/totp/setup", mfaHandler.SetupTOTP)
				mfaGroup.POST("/totp/verify", mfaHandler.VerifyTOTP)
				mfaGroup.DELETE("/totp", mfaHandler.DeleteTOTP)
				mfaGroup.POST("/passkeys/register/begin", mfaHandler.BeginPasskeyRegistration)
				mfaGroup.POST("/passkeys/register/complete", mfaHandler.CompletePasskeyRegistration)
				mfaGroup.GET("/passkeys", mfaHandler.ListPasskeys)
				mfaGroup.PUT("/passkeys/:id", mfaHandler.RenamePasskey)
				mfaGroup.DELETE("/passkeys/:id", mfaHandler.DeletePasskey)
				mfaGroup.GET("/backup-codes/count", mfaHandler.GetBackupCodeCount)
				mfaGroup.POST("/backup-codes/regenerate", mfaHandler.RegenerateBackupCodes)
				mfaGroup.POST("/disable", mfaHandler.DisableMFA)
			}
		}
	}

	// Admin routes
	adminGroup := verified.Group("/admin")
	adminGroup.Use(middleware.AdminRequired())
	adminGroup.Use(middleware.MFARequiredForAdmin())
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
		adminGroup.POST("/badges/recalculate", gamificationHandler.AdminRecalculateBadges)
	}

	// Polar supporter integration (only when configured)
	if cfg.PolarEnabled() {
		polarClient := polar.NewClient(cfg.PolarAccessToken, cfg.PolarSandbox)
		polarHandler := polar.NewHandler(polarClient, cfg)
		polarWebhookHandler := polar.NewWebhookHandler(cfg.PolarWebhookSecret, userRepo, gamificationRepo)

		// Webhook endpoint (public, verified by signature)
		v1.POST("/webhooks/polar", polarWebhookHandler.HandleWebhook)

		// Checkout + portal endpoints (authenticated)
		verified.GET("/support/checkout", polarHandler.CreateCheckout)
		verified.GET("/support/portal", polarHandler.GetPortalURL)
	}

	// Start DNS re-verification background task.
	linkedaccount.StartReverification(ctx, linkedAccountRepo)

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
