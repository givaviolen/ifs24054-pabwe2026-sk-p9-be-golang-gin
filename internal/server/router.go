package server

import (
	"net/http"
	"time"

	"app/docs"
	"app/internal/config"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/middleware"
	"app/internal/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Deps memungkinkan klien Search Engine & OpenAI diganti (mis. saat testing).
type Deps struct {
	Searcher  ai.Searcher
	Completer ai.Completer
}

func NewRouter(cfg *config.Config, db *gorm.DB, deps *Deps) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	if deps == nil {
		deps = &Deps{}
	}
	if deps.Searcher == nil {
		deps.Searcher = ai.NewSearchClient(cfg.SearchEngineEndpoint)
	}
	if deps.Completer == nil {
		deps.Completer = ai.NewOpenAIClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel)
	}

	authH := auth.NewHandler(auth.NewService(auth.NewRepository(db), cfg.AppKey))
	postH := post.NewHandler(post.NewService(post.NewRepository(db)))
	aiH := ai.NewHandler(ai.NewService(deps.Searcher, deps.Completer))

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	r.GET("/api.json", func(c *gin.Context) { c.Data(http.StatusOK, "application/json", docs.Spec) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api.json")))

	api := r.Group("/api")

	pub := api.Group("/public")
	pub.GET("/categories", postH.Categories)
	pub.GET("/posts", postH.PublicList)
	pub.GET("/posts/:slug", postH.PublicDetail)

	a := api.Group("/auth")
	a.POST("/register", authH.Register)
	a.POST("/login", authH.Login)
	a.POST("/logout", middleware.Auth(cfg.AppKey), authH.Logout)
	a.GET("/profile", middleware.Auth(cfg.AppKey), authH.Profile)
	a.PUT("/profile", middleware.Auth(cfg.AppKey), authH.UpdateProfile)

	p := api.Group("/posts", middleware.Auth(cfg.AppKey))
	p.GET("", postH.List)
	p.POST("", postH.Create)
	p.GET("/stats", postH.Stats)
	p.POST("/ai-generate", middleware.RateLimiter(10, time.Minute), aiH.Generate)
	p.GET("/:id", postH.Detail)
	p.PUT("/:id", postH.Update)
	p.PATCH("/:id/toggle-publish", postH.TogglePublish)
	p.DELETE("/:id", postH.Delete)

	adm := api.Group("/admin", middleware.Auth(cfg.AppKey), middleware.RequireRole("admin"))
	adm.GET("/users", postH.AdminUsers)
	adm.GET("/posts", postH.AdminPosts)
	adm.GET("/users/:id/posts", postH.AdminUserPosts)

	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "Endpoint not found")
	})
	return r
}
