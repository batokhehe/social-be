package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"social-be/internal/config"
	"social-be/internal/domain/attributevolunteer"
	"social-be/internal/domain/auth"
	"social-be/internal/domain/categoryactivity"
	"social-be/internal/domain/levelarea"
	"social-be/internal/domain/levelvolunteer"
	"social-be/internal/domain/masteractivity"
	"social-be/internal/domain/masterarea"
	"social-be/internal/domain/religion"
	"social-be/internal/domain/user"
	"social-be/internal/domain/volunteer"
	"social-be/internal/middleware"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/tracing"

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

	shutdownTracing, err := tracing.Init(context.Background(), "social-be")
	if err != nil {
		logger.Log.WithError(err).Fatal("failed to initialize tracing")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(ctx); err != nil {
			logger.Log.WithError(err).Error("failed to shutdown tracing")
		}
	}()

	_, err = cache.RDB.Ping(cache.Ctx).Result()
	if err != nil {
		logger.Log.WithError(err).Fatal("redis not connected")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	userRepo := user.NewGormRepository(db, logger.Log)
	levelAreaRepo := levelarea.NewGormRepository(db, logger.Log)
	masterAreaRepo := masterarea.NewGormRepository(db, logger.Log)
	levelVolunteerRepo := levelvolunteer.NewGormRepository(db, logger.Log)
	attributeVolunteerRepo := attributevolunteer.NewGormRepository(db, logger.Log)
	categoryActivityRepo := categoryactivity.NewGormRepository(db, logger.Log)
	religionRepo := religion.NewGormRepository(db, logger.Log)
	masterActivityRepo := masteractivity.NewGormRepository(db, logger.Log)
	volunteerRepo := volunteer.NewGormRepository(db, logger.Log)

	userService := &user.Service{Repo: userRepo}
	levelAreaService := &levelarea.Service{Repo: levelAreaRepo}
	masterAreaService := &masterarea.Service{Repo: masterAreaRepo}
	levelVolunteerService := &levelvolunteer.Service{Repo: levelVolunteerRepo}
	attributeVolunteerService := &attributevolunteer.Service{Repo: attributeVolunteerRepo}
	categoryActivityService := &categoryactivity.Service{Repo: categoryActivityRepo}
	religionService := &religion.Service{Repo: religionRepo}
	masterActivityService := &masteractivity.Service{Repo: masterActivityRepo}
	volunteerService := &volunteer.Service{Repo: volunteerRepo}
	authService := auth.NewService(userService, volunteerService)

	userHandler := &user.Handler{Service: userService}
	authHandler := &auth.Handler{Service: authService}
	levelAreaHandler := &levelarea.Handler{Service: levelAreaService}
	masterAreaHandler := &masterarea.Handler{Service: masterAreaService}
	levelVolunteerHandler := &levelvolunteer.Handler{Service: levelVolunteerService}
	attributeVolunteerHandler := &attributevolunteer.Handler{Service: attributeVolunteerService}
	categoryActivityHandler := &categoryactivity.Handler{Service: categoryActivityService}
	religionHandler := &religion.Handler{Service: religionService}
	masterActivityHandler := &masteractivity.Handler{Service: masterActivityService}
	volunteerHandler := &volunteer.Handler{Service: volunteerService}

	// gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(otelgin.Middleware("social-be"))
	r.Use(middleware.TimeoutMiddleware(15 * time.Second))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.FromContext(c.Request.Context()).WithField("panic", recovered).Error("panic recovered")
		response.WriteError(c, apperror.New(http.StatusInternalServerError, apperror.CodeInternal, "internal server error"))
		c.Abort()
	}))
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.ErrorMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(cors.Default())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API documentation routes for Apidog integration
	r.GET("/docs/openapi.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.File("docs/swagger.json")
	})
	r.GET("/docs/openapi.yaml", func(c *gin.Context) {
		c.Header("Content-Type", "application/yaml")
		c.File("docs/swagger.yaml")
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .btn { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 5px; }
        .btn:hover { background: #0056b3; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>API Documentation</h1>
    
    <div class="section">
        <h2>Swagger UI</h2>
        <p>Interactive API documentation with testing capabilities.</p>
        <a href="/swagger/index.html" class="btn">Open Swagger UI</a>
    </div>
    
    <div class="section">
        <h2>Apidog Integration</h2>
        <p>Import the OpenAPI specification into Apidog for enhanced API design and testing.</p>
        
        <h3>Quick Import URLs:</h3>
        <p><strong>JSON:</strong> <code>`+c.Request.Host+`/docs/openapi.json</code></p>
        <p><strong>YAML:</strong> <code>`+c.Request.Host+`/docs/openapi.yaml</code></p>
        
        <h3>How to Import:</h3>
        <ol>
            <li>Open Apidog application</li>
            <li>Click "Import" or "New Project from OpenAPI"</li>
            <li>Paste one of the URLs above</li>
            <li>Click "Import"</li>
        </ol>
        
        <p><em>Note: The specification auto-refreshes on every build.</em></p>
    </div>
    
    <div class="section">
        <h2>Direct Download</h2>
        <p>Download the OpenAPI specification files directly.</p>
        <a href="/docs/openapi.json" class="btn" download>Download JSON</a>
        <a href="/docs/openapi.yaml" class="btn" download>Download YAML</a>
    </div>
</body>
</html>`)
	})

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
	protected.POST("/users/upload-photo", userHandler.UploadProfilePhoto)
	protected.GET("/volunteers", volunteerHandler.GetAll)
	protected.GET("/volunteers/:id", volunteerHandler.GetByID)
	protected.POST("/volunteers", volunteerHandler.Create)
	protected.PUT("/volunteers/:id", volunteerHandler.Update)
	protected.DELETE("/volunteers/:id", volunteerHandler.Delete)

	master := protected.Group("/master")
	master.GET("/level-areas", levelAreaHandler.GetAll)
	master.GET("/level-areas/:id", levelAreaHandler.GetByID)
	master.POST("/level-areas", levelAreaHandler.Create)
	master.PUT("/level-areas/:id", levelAreaHandler.Update)
	master.DELETE("/level-areas/:id", levelAreaHandler.Delete)
	master.GET("/master-areas", masterAreaHandler.GetAll)
	master.GET("/master-areas/:id", masterAreaHandler.GetByID)
	master.POST("/master-areas", masterAreaHandler.Create)
	master.PUT("/master-areas/:id", masterAreaHandler.Update)
	master.DELETE("/master-areas/:id", masterAreaHandler.Delete)
	master.GET("/level-volunteers", levelVolunteerHandler.GetAll)
	master.GET("/level-volunteers/:id", levelVolunteerHandler.GetByID)
	master.POST("/level-volunteers", levelVolunteerHandler.Create)
	master.PUT("/level-volunteers/:id", levelVolunteerHandler.Update)
	master.DELETE("/level-volunteers/:id", levelVolunteerHandler.Delete)
	master.GET("/attribute-volunteers", attributeVolunteerHandler.GetAll)
	master.GET("/attribute-volunteers/:id", attributeVolunteerHandler.GetByID)
	master.POST("/attribute-volunteers", attributeVolunteerHandler.Create)
	master.PUT("/attribute-volunteers/:id", attributeVolunteerHandler.Update)
	master.DELETE("/attribute-volunteers/:id", attributeVolunteerHandler.Delete)
	master.GET("/category-activities", categoryActivityHandler.GetAll)
	master.GET("/category-activities/:id", categoryActivityHandler.GetByID)
	master.POST("/category-activities", categoryActivityHandler.Create)
	master.PUT("/category-activities/:id", categoryActivityHandler.Update)
	master.DELETE("/category-activities/:id", categoryActivityHandler.Delete)
	master.GET("/religions", religionHandler.GetAll)
	master.GET("/religions/:id", religionHandler.GetByID)
	master.POST("/religions", religionHandler.Create)
	master.PUT("/religions/:id", religionHandler.Update)
	master.DELETE("/religions/:id", religionHandler.Delete)
	master.GET("/master-activities", masterActivityHandler.GetAll)
	master.GET("/master-activities/:id", masterActivityHandler.GetByID)
	master.POST("/master-activities", masterActivityHandler.Create)
	master.PUT("/master-activities/:id", masterActivityHandler.Update)
	master.DELETE("/master-activities/:id", masterActivityHandler.Delete)

	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware(1))
	admin.GET("/users", userHandler.GetUsers)

	log.Println("Server running at :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
