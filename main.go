package main

import (
	"context"
	"embed"
	"fmt"
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

type config struct {
	Port    string `env:"HTTP_LISTEN_ADDR" envDefault:":3001"`
	SiteURL string `env:"SITE_URL" envDefault:"https://blog.fiorillo.xyz"`
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

	pageHandler := handler.NewPageHandler(logger, cfg.SiteURL)

	router := chi.NewMux()
	router.Handle("/*", public())
	router.Get("/", handler.Make(pageHandler.HandleIndexPage))
	router.Get("/blog", handler.Make(pageHandler.HandleBlogPage))
	router.Get("/blog/{filename}", handler.Make(pageHandler.HandleBlogPostPage))
	router.Get("/aboutme", handler.Make(pageHandler.HandleAboutMePage))
	router.Get("/feed.xml", handler.Make(pageHandler.HandleFeed))
	router.Get("/sitemap.xml", handler.Make(pageHandler.HandleSitemap))
	router.Get("/routes", handler.Make(pageHandler.HandleRoutes))
	router.Get("/routes/badger-divide", handler.Make(pageHandler.HandleRoute))
	router.Get("/api/routes/badger-divide/coords", handler.Make(pageHandler.HandleRouteCoords))

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info(fmt.Sprintf("application running: http://localhost%s", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.ErrAttr(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}
}

func public() http.Handler {
	return http.FileServerFS(publicFS)
}
