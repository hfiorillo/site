package main

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/hfiorillo/site/handler"
	"github.com/hfiorillo/site/utils/logging"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

//go:embed public
var publicFS embed.FS

// define env vars
type config struct {
	Port string `env:"HTTP_LISTEN_ADDR" envDefault:":3001"`
}

func main() {

	logger := logging.NewJsonLogger()

	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found.")
	}

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error(err.Error())
	}

	// load and parse markdown files

	postsHandler := handler.NewPostsHandler(logger)
	generalHandler := handler.NewGeneralHandler(logger)

	router := chi.NewMux()
	router.Handle("/*", public())
	router.Get("/", handler.Make(handler.HandleHomeIndex))
	router.Get("/", handler.Make(postsHandler.ListBlogPosts))
	router.Get("/blog", handler.Make(postsHandler.ListBlogPosts))
	router.Get("/images", handler.Make(postsHandler.ListBlogPosts))
	router.Get("/aboutme", handler.Make(generalHandler.HandleAboutMe))

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("application running", "link: http://localhost"+cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", err)
		}
	}()

	// Create a channel to receive OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stop

	slog.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed: %v", err)
		os.Exit(1)
	}

}

func public() http.Handler {
	return http.FileServerFS(publicFS)
}
