package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/view/pages"
)

type PageHandler struct {
	Logger *slog.Logger
}

func NewPageHandler(logger *slog.Logger) *PageHandler {
	return &PageHandler{
		Logger: logger,
	}
}

func (p PageHandler) HandleIndexPage(w http.ResponseWriter, r *http.Request) error {
	return pages.Index().Render(r.Context(), w)
}

func (p PageHandler) HandleAboutMePage(w http.ResponseWriter, r *http.Request) error {
	return pages.AboutMe().Render(r.Context(), w)
}

func (p PageHandler) HandleBlogPage(w http.ResponseWriter, r *http.Request) error {
	const postsPath = "./content/posts"
	posts, err := markdown.LoadMarkdownPosts(postsPath)

	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	return pages.Blog(posts).Render(r.Context(), w)
}
