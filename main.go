package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hfiorillo/site/handler"
	"github.com/hfiorillo/site/pkg"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

//go:embed public
var publicFS embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// load and parse markdown files
	posts, err := pkg.LoadMarkdownPosts("./content/posts")
	if err != nil {
		log.Fatal(err)
	}

	postsHandler := handler.NewPostsHandler(posts)

	router := chi.NewMux()
	router.Handle("/*", public())
	router.Get("/", handler.Make(handler.HandleHomeIndex))
	router.Get("/", handler.Make(postsHandler.ListBlogPosts))
	router.Get("/blog", handler.Make(postsHandler.ListBlogPosts))

	port := os.Getenv("HTTP_LISTEN_ADDR")

	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("application running", "link: http://localhost"+port)
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
