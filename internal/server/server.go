package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"bynd-cms-backend/internal/auth"
	"bynd-cms-backend/internal/community"
	"bynd-cms-backend/internal/config"
	"bynd-cms-backend/internal/content"
	"bynd-cms-backend/internal/httpx"
	"bynd-cms-backend/internal/wellness"
)

func New(cfg config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), cors(cfg.AllowedOrigins), securityHeaders())

	authService := auth.NewService(cfg.JWTSecret)
	authHandler := auth.NewHandler(authService, cfg.AdminUsername, cfg.AdminPassword, cfg.AdminPasswordHash)
	communityHandler := community.NewHandler(community.NewClient(cfg.BYNDAPIBaseURL, cfg.InternalToken, cfg.RequestTimeout))
	wellnessHandler := wellness.NewHandler(wellness.NewClient(cfg.BYNDAPIBaseURL, cfg.InternalToken, cfg.RequestTimeout))
	contentHandler := content.NewHandler(content.NewClient(cfg.BYNDAPIBaseURL, cfg.InternalToken, cfg.RequestTimeout))

	engine.GET("/healthz", func(c *gin.Context) { httpx.Success(c, gin.H{"status": "ok"}) })
	api := engine.Group("/api")
	api.POST("/auth/login", authHandler.Login)

	secured := api.Group("")
	secured.Use(auth.Required(authService))
	secured.GET("/user/info", authHandler.UserInfo)
	secured.GET("/community/groups", communityHandler.ListGroups)
	secured.GET("/community/groups/:groupId", communityHandler.GetGroup)
	secured.GET("/community/challenges", communityHandler.ListChallenges)
	secured.GET("/community/challenges/:challengeId", communityHandler.GetChallenge)
	secured.POST("/community/challenges", communityHandler.CreateChallenge)
	secured.PATCH("/community/challenges/:challengeId", communityHandler.UpdateChallenge)
	secured.PUT("/community/challenges/:challengeId/days/:dayNumber", communityHandler.SaveChallengeDay)
	secured.POST("/community/challenges/:challengeId/publish", communityHandler.PublishChallenge)
	secured.POST("/community/challenges/:challengeId/archive", communityHandler.ArchiveChallenge)
	secured.GET("/wellness/sleep", wellnessHandler.ListSleep)
	secured.GET("/wellness/sleep/:id", wellnessHandler.GetSleep)
	secured.GET("/wellness/journals", wellnessHandler.ListJournals)
	secured.GET("/wellness/journals/:id", wellnessHandler.GetJournal)
	secured.GET("/wellness/heart-rate", wellnessHandler.ListHeartRate)
	secured.GET("/wellness/heart-rate/:id", wellnessHandler.GetHeartRate)
	secured.GET("/content/readings/categories", contentHandler.ListCategories)
	secured.POST("/content/readings/categories", contentHandler.CreateCategory)
	secured.PUT("/content/readings/categories/:categoryId", contentHandler.UpdateCategory)
	secured.DELETE("/content/readings/categories/:categoryId", contentHandler.DeleteCategory)
	secured.GET("/content/readings/publications", contentHandler.ListPublications)
	secured.POST("/content/readings/publications", contentHandler.CreatePublication)
	secured.GET("/content/readings/publications/:publicationId", contentHandler.GetPublication)
	secured.PUT("/content/readings/publications/:publicationId", contentHandler.UpdatePublication)
	secured.DELETE("/content/readings/publications/:publicationId", contentHandler.ArchivePublication)
	secured.GET("/content/readings/assets", contentHandler.ListAssets)
	secured.POST("/content/readings/assets", contentHandler.UploadAsset)
	secured.DELETE("/content/readings/assets/:assetId", contentHandler.DeleteAsset)
	secured.GET("/content/guidance/body-foundations", contentHandler.GetGuide)
	secured.PUT("/content/guidance/body-foundations", contentHandler.SaveGuide)
	secured.PUT("/content/guidance/body-foundations/days/:day", contentHandler.SaveGuideDay)

	return engine
}

func cors(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

func HTTPServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: cfg.HTTPAddr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 2 * time.Minute,
		WriteTimeout: 2 * time.Minute, IdleTimeout: 60 * time.Second,
	}
}

func IsAllowedOrigin(origin string, origins []string) bool {
	for _, allowed := range origins {
		if strings.EqualFold(strings.TrimSpace(origin), strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}
