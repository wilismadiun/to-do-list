package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/src/Applications/security"
	"user-service/src/Applications/usecase"
	"user-service/src/Infrastructures/database"
	infrarepo "user-service/src/Infrastructures/repository"
	"user-service/src/Interfaces/http/handlers"
	"user-service/src/Interfaces/http/routes"
	"user-service/src/commons/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.NewPostgres(database.Config{Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser, Password: cfg.DBPassword, Name: cfg.DBName, SSLMode: cfg.DBSSLMode})
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database pool: %v", err)
	}
	defer sqlDB.Close()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health/live", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "alive"}) })
	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	userRepo := infrarepo.NewGormUserRepository(db)
	registerUC := usecase.NewRegisterUseCase(userRepo, security.NewBcryptPasswordHasher())
	authHandler := handlers.NewAuthHandler(registerUC)
	api := router.Group("/api/v1")
	routes.RegisterAuthRoutes(api, authHandler)

	server := &http.Server{Addr: ":" + cfg.AppPort, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("user-service listening on :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-stop:
		log.Printf("received signal %s; shutting down", sig)
	case err := <-serverErr:
		log.Printf("http server error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
