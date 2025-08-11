package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"modress/internal/controllers"
	"modress/internal/middleware"
	"modress/internal/repositories"
	"modress/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Get environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Connect to database with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize validator
	validate := validator.New()
	wsController := controllers.NewWebSocketController()
	go wsController.Run()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db) // Assumes this exists
	storeRepo := repositories.NewStoreRepository(db)     // Assumes this exists
	

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtSecret)
	storeService := services.NewStoreService(storeRepo)
	productService := services.NewProductService(productRepo, storeRepo)

	authController := controllers.NewAuthController(authService)
	productController := controllers.NewProductController(productService, storeService)
	storeController := controllers.NewStoreController(storeService)
	// No seu main.go, antes do router.Run()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery()) // Manually add default middleware (replaces gin.Default())
	router.RedirectTrailingSlash = false

	router.Use(middleware.CORSMiddleware())

	router.Use(func(c *gin.Context) {
		c.Set("validator", validate)
		c.Next()
	})

	router.Static("/images", "./uploads/images")

	// Group all routes under /api/v1
	api := router.Group("/api/v1")

	// WebSocket route
	api.GET("/ws", middleware.AuthMiddleware(jwtSecret), wsController.HandleConnections)

	// Health Check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Public routes (auth)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}

	temp := api.Group("/temp")
	{
		temp.GET("/users", func(c *gin.Context) {
			users, err := userRepo.ListAll(c.Request.Context())
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, users)
		})
	}

	// Protected routes (users)
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtSecret))
	{
		users.GET("/me", authController.GetProfile)
		users.PUT("/me", authController.UpdateProfile)
		users.DELETE("/me", authController.DeleteProfile)
	}

	// Product routes
	products := api.Group("/products")
	{
		// Public product routes
		products.GET("/:id", productController.GetProduct)
		products.GET("/store/:storeId", productController.GetProductsByStore)
		products.GET("/category/:category", productController.GetProductsByCategory)
		products.GET("/", productController.ListProducts)
		products.GET("/search", productController.SearchProducts)

		// Protected product routes (require authentication)
		products.Use(middleware.AuthMiddleware(jwtSecret))
		{
			products.POST("/", productController.CreateProduct)
			products.PUT("/:id", productController.UpdateProduct)
			products.DELETE("/:id", productController.DeleteProduct)
			products.PUT("/:id/quantity", productController.UpdateQuantity)
		}
	}

	// Store routes
	stores := api.Group("/stores")
	{
		// Public store routes
		stores.GET("/:id", storeController.GetStore)
		stores.GET("/slug/:slug", storeController.GetStoreBySlug)
		stores.GET("/", storeController.ListStores)
		stores.GET("/approved", storeController.ListApprovedStores)

		// Protected store routes (require authentication)
		stores.Use(middleware.AuthMiddleware(jwtSecret))
		{
			stores.POST("/", storeController.CreateStore)
			stores.GET("/my", storeController.GetMyStore)
			stores.PUT("/:id", storeController.UpdateStore)
			stores.DELETE("/:id", storeController.DeleteStore)

			// Admin-only route
			stores.PUT("/:id/approve", storeController.ApproveStore)
		}
	}

	// Start the server
	log.Printf("Server running on port %s...", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
