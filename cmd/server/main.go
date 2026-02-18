package main

import (
	"log/slog"
	//_ "cshdMediaDelivery/docs" swagger потом
	"cshdMediaDelivery/internal/app"
	"cshdMediaDelivery/internal/config"
	"cshdMediaDelivery/internal/lib/liblogger"
	"os"
	"os/signal"
	"syscall"
)

const LocalFilePath = "config-yaml/dev.yaml"
const DockerFilePath = "config-yaml/docker.yaml"

func main() {
	// Init config
	cfg := config.GetConfig(LocalFilePath)

	// Init logger
	log := liblogger.SetupLogger(cfg.Env)
	log.Info("start cshdMediaDelivery", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init app
	app := app.New(log, cfg)
	// init gorutine for start server
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// start app
	go app.MustRun()
	log.Info("server started")
	<-done
	// stop app
	app.Stop()
	log.Info("server stopped")
}
