package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/hfiorillo/site/handler"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/utils/logging"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	stylesURL, err := stylesheetURL(publicFS)
	if err != nil {
		logger.Error("loading stylesheet", "error", err)
		os.Exit(1)
	}
	paths.StylesURL = stylesURL

	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found.")
	}

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error(err.Error())
	}

	pageHandler := handler.NewPageHandler(logger, cfg.SiteURL)

	router := chi.NewMux()
	router.Use(chimiddleware.Logger, chimiddleware.Recoverer, chimiddleware.Timeout(30*time.Second))
	router.Handle("/*", public())
	router.Get(paths.Root, handler.Make(pageHandler.HandleIndexPage))
	router.Get(paths.Blog, handler.Make(pageHandler.HandleBlogPage))
	router.Get(paths.BlogPost, handler.Make(pageHandler.HandleBlogPostPage))
	router.Get(paths.AboutMe, handler.Make(pageHandler.HandleAboutMePage))
	router.Get(paths.AboutThisSite, handler.Make(pageHandler.HandleAboutThisSite))
	router.Get(paths.Pictures, handler.Make(pageHandler.HandlePictures))
	router.Get(paths.Work, handler.Make(pageHandler.HandleWork))
	router.Get(paths.Feed, handler.Make(pageHandler.HandleFeed))
	router.Get(paths.Sitemap, handler.Make(pageHandler.HandleSitemap))
	router.Get(paths.Routes, handler.Make(pageHandler.HandleRoutes))
	router.Get(paths.RouteDetail, handler.Make(pageHandler.HandleRoute))
	router.Get(paths.RouteCoords, handler.Make(pageHandler.HandleRouteCoords))

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info(fmt.Sprintf("application running: http://localhost%s", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", logging.ErrAttr(err))
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

// stylesheetURL fingerprints the bytes served by this binary, not files on disk.
func stylesheetURL(assets fs.FS) (string, error) {
	css, err := fs.ReadFile(assets, strings.TrimPrefix(paths.Styles, "/"))
	if err != nil {
		return "", fmt.Errorf("reading embedded stylesheet: %w", err)
	}
	hash := sha256.Sum256(css)
	return fmt.Sprintf("%s?v=%x", paths.Styles, hash[:12]), nil
}
