package app

import (
	"context"
	"cshdMediaDelivery/internal/config"
	"cshdMediaDelivery/internal/handlers"
	"cshdMediaDelivery/internal/lib/liblogger"
	"cshdMediaDelivery/internal/middleware/midlogger"
	"cshdMediaDelivery/internal/services"
	"cshdMediaDelivery/internal/storage/fs"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) *App {

	// init storage
	//storage := postgresql.MustPosgreSQL(cfg.GetDataSourceName())
	//log.Info("storage are enabled")

	fsStorage := fs.NewFSStorage("./uploads")
	// s3Storage := s3.NewS3Storage(...)
	mediaService := services.NewMediaService(fsStorage)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	// init router
	router := chi.NewRouter()

	app := &App{log: log}
	//TODO
	// init cors
	//app.initCors(router, cfg.AdditionalAddressesConfig)
	// init middlewares
	router.Use(midlogger.New(log))
	router.Use(middleware.URLFormat)
	// add Authentication with JWT token

	// router.Use(func(next http.Handler) http.Handler {
	// 	return auth.AuthenticateMiddleware(next, cfg.Key)
	// })

	//TODO
	// init connection manager for rabbit
	//connectionManager := rabbitmq.New(cfg.AddressRabbitPath, log)
	//connectionManager.Start(context.TODO())
	//
	//// init rabbit consumer
	//rabbitConsumer := consumer.New(log, connectionManager, eventService)
	//rabbitConsumer.Start(context.TODO(), cfg.QueueName)
	//
	//// init rabbit producer
	//rabbitProducer := producer.New(log, connectionManager, gormORM, &outbox_repository.OutboxRepository{})
	//rabbitProducer.Start(context.TODO())
	//TODO

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
	//router.Get("/swagger/*", httpSwagger.WrapHandler)

	// init subjects route
	router.Post("/upload", mediaHandler.Upload)
	router.Get("/{key}", mediaHandler.Get)
	router.Delete("/{key}", mediaHandler.Delete)

	// init events route
	// router.With(auth.RoleBasedAccess(userrole.AdminRole)).Group(func(r chi.Router) {
	// 	r.Post("/api/events", eventHandler.CreateEvent)
	// 	r.Put("/api/events/{id}", eventHandler.UpdateEvent)
	// 	r.Delete("/api/events/{id}", eventHandler.DeleteEvent)
	// })
}

func (a *App) initCors(router *chi.Mux, cfg config.AdditionalAddressesConfig) {
	corsOptions := cors.Options{
		AllowedOrigins: []string{cfg.ReactVision, cfg.JureAssignmentsService, cfg.ApiGateway},
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

	a.log.Info("server starting")
	if err := a.server.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failee to stop sever", liblogger.Err(err))
		return
	}
	a.log.Info("server stopped")
}
