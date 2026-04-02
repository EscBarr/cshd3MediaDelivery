package app

import (
	"context"
	"cshdMediaDelivery/internal/config"
	"cshdMediaDelivery/internal/handlers"
	"cshdMediaDelivery/internal/lib/liblogger"
	"cshdMediaDelivery/internal/middleware/midlogger"
	"cshdMediaDelivery/internal/services"
	"cshdMediaDelivery/internal/storage"
	s3 "cshdMediaDelivery/internal/storage/S3storage"
	"cshdMediaDelivery/internal/storage/fs"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type App struct {
	server    *http.Server
	log       *slog.Logger
	scheduler *services.Scheduler // Добавляем планировщик
}

func New(log *slog.Logger, cfg *config.Config) *App {

	// init storage
	var mediaStorage storage.Storage

	if cfg.MinioConfig.Endpoint != "" {
		var s3Storage storage.Storage
		var err error

		retries := 10
		delay := 3 * time.Second

		for i := 0; i < retries; i++ {
			s3Storage, err = s3.New(
				cfg.MinioConfig.Endpoint,
				cfg.MinioConfig.AccessKeyID,
				cfg.MinioConfig.SecretAccessKey,
				cfg.MinioConfig.Bucket,
			)

			if err == nil {
				break
			}

			log.Warn("failed to connect to S3",
				"attempt", i+1,
				"error", err,
			)

			time.Sleep(delay)
		}

		if err != nil {
			log.Error("failed to init s3 storage", err)
			panic(err)
		}

		mediaStorage = s3Storage

	} else {
		mediaStorage = fs.NewFSStorage("./uploads")
	}

	mediaService := services.NewMediaService(mediaStorage)
	mediaHandler := handlers.NewMediaHandler(mediaService)

	// Создаем планировщик
	scheduler := services.NewScheduler(
		mediaService,
		cfg.DatabaseConfig,
	)

	// init router
	router := chi.NewRouter()

	app := &App{
		log:       log,
		scheduler: scheduler,
	}

	// init cors
	app.initCors(router, cfg.AdditionalAddressesConfig)
	// init middlewares
	router.Use(midlogger.New(log))
	router.Use(middleware.URLFormat)

	// init routes
	app.initRoutes(router, mediaHandler)

	// init server
	app.server = &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: router,
	}

	return app
}

func (a *App) initRoutes(router *chi.Mux,
	mediaHandler *handlers.MediaHandler,
) {
	router.Post("/upload", mediaHandler.Upload)
	router.Get("/{key}", mediaHandler.Get)
	router.Delete("/{key}", mediaHandler.Delete)
}

func (a *App) initCors(router *chi.Mux, cfg config.AdditionalAddressesConfig) {
	corsOptions := cors.Options{
		AllowedOrigins: []string{cfg.Vue},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Link",
			"Content-Length",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Credentials",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
	router.Use(cors.Handler(corsOptions))
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"

	// Запускаем планировщик
	a.scheduler.Start()
	a.log.Info("scheduler started")

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	serverErr := make(chan error, 1)
	go func() {
		a.log.Info("server starting", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("%s: %w", op, err)
		}
	}()

	// Ожидаем сигнал остановки или ошибку сервера
	select {
	case err := <-serverErr:
		a.log.Error("server error", liblogger.Err(err))
		return err
	case <-stop:
		a.log.Info("shutdown signal received")
		return a.Stop()
	}
}

func (a *App) Stop() error {
	const op = "app.Stop"

	a.log.Info("stopping application")

	// Останавливаем планировщик
	a.scheduler.Stop()
	a.log.Info("scheduler stopped")

	// Останавливаем HTTP сервер
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failed to stop server", liblogger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("server stopped")
	return nil
}
