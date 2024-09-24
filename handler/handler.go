package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	// return pages.Index().Render(r.Context(), w)
	aboutme, err := markdown.LoadMarkdownPost("/aboutme/aboutme")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}
	return pages.AboutMe(aboutme).Render(r.Context(), w)

}

func (p PageHandler) HandleAboutMePage(w http.ResponseWriter, r *http.Request) error {
	aboutme, err := markdown.LoadMarkdownPost("/aboutme/aboutme")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}
	return pages.AboutMe(aboutme).Render(r.Context(), w)
}

func (p PageHandler) HandleBlogPage(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	return pages.Blog(posts).Render(r.Context(), w)
}

func (p PageHandler) HandleBlogPostPage(w http.ResponseWriter, r *http.Request) error {

	post, err := markdown.LoadMarkdownPost(fmt.Sprintf("/posts/%s", chi.URLParam(r, "filename")))
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	return pages.BlogPage(post).Render(r.Context(), w)
}
