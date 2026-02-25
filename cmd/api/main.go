package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/galihaleanda/todo-app/internal/config"
	"github.com/galihaleanda/todo-app/internal/handler"
	"github.com/galihaleanda/todo-app/internal/repository"
	"github.com/galihaleanda/todo-app/internal/service"
	pkgjwt "github.com/galihaleanda/todo-app/pkg/jwt"
	"github.com/galihaleanda/todo-app/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Bootstrap logger
	log := logger.New(cfg.App.LogLevel, cfg.App.Env)
	log.WithField("env", cfg.App.Env).Info("starting todo-app")

	// 3. Connect to PostgreSQL
	db, err := connectDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Close()
	log.Info("connected to database")

	// 4. Wire dependencies (manual DI — no framework needed at this scale)
	jwtManager := pkgjwt.New(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager, log)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, log)
	projectSvc := service.NewProjectService(projectRepo, log)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// Router
	router := handler.NewRouter(authHandler, taskHandler, projectHandler, analyticsHandler, jwtManager, log)
	engine := router.Setup()

	// 5. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("listening on :%s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("server error")
		}
	}()

	// 6. Graceful shutdown on SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("server forced shutdown")
	}

	log.Info("server stopped cleanly")
}

// connectDB establishes and configures the PostgreSQL connection pool.
func connectDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}
