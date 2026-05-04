package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"social-be/internal/config"
	"social-be/internal/domain/auth"
	"social-be/internal/domain/levelarea"
	"social-be/internal/domain/user"
	"social-be/internal/middleware"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "social-be/docs"
)

// @title Social BE API
// @version 1.0
// @description API documentation
// @host
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load()

	if err := config.RequireEnv("JWT_SECRET", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"); err != nil {
		log.Fatal(err)
	}

	cache.Init()
	logger.Init()
	defer func() {
		_ = logger.Log.Sync()
	}()

	_, err := cache.RDB.Ping(cache.Ctx).Result()
	if err != nil {
		logger.Log.Fatal("redis not connected", zap.Error(err))
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	userRepo := &user.GormRepository{DB: db}
	levelAreaRepo := &levelarea.GormRepository{DB: db}

	userService := &user.Service{Repo: userRepo}
	levelAreaService := &levelarea.Service{Repo: levelAreaRepo}

	userHandler := &user.Handler{Service: userService}
	authHandler := &auth.Handler{UserService: userService}
	levelAreaHandler := &levelarea.Handler{Service: levelAreaService}

	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.ErrorMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(cors.Default())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health/live", middleware.LivenessHandler)
	r.GET("/health/ready", middleware.ReadinessHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api")

	api.Use(middleware.RateLimitMiddleware())
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	api.POST("/refresh", authHandler.Refresh)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/users", userHandler.GetUsers)
	protected.GET("/users/:id", userHandler.GetUserByID)

	master := protected.Group("/master")
	master.GET("/level-areas", levelAreaHandler.GetAll)
	master.GET("/level-areas/:id", levelAreaHandler.GetByID)
	master.POST("/level-areas", levelAreaHandler.Create)
	master.PUT("/level-areas/:id", levelAreaHandler.Update)
	master.DELETE("/level-areas/:id", levelAreaHandler.Delete)

	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware(1))
	admin.GET("/users", userHandler.GetUsers)

	log.Println("Server running at :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
