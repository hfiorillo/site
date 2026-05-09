package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/view/pages"
)

type PageHandler struct {
	Logger  *slog.Logger
	SiteURL string
}

func NewPageHandler(logger *slog.Logger, siteURL string) *PageHandler {
	return &PageHandler{
		Logger:  logger,
		SiteURL: siteURL,
	}
}

func Make(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Error("internal server error", "err", err, "path", r.URL.Path)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func (p PageHandler) HandleIndexPage(w http.ResponseWriter, r *http.Request) error {
	aboutme, err := markdown.LoadMarkdownPost("/aboutme/aboutme")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	meta := models.PageMeta{
		Title:       "Harry Fiorillo-Hughes",
		Description: "DevOps Engineer and Platform Engineer based in Manchester, UK.",
		URL:         p.SiteURL + "/",
		Image:       p.SiteURL + "/public/images/avatar.jpg",
	}
	return pages.AboutMe(aboutme, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleAboutMePage(w http.ResponseWriter, r *http.Request) error {
	aboutme, err := markdown.LoadMarkdownPost("/aboutme/aboutme")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	meta := models.PageMeta{
		Title:       "About Me | Harry Fiorillo-Hughes",
		Description: "DevOps Engineer and Platform Engineer based in Manchester, UK.",
		URL:         p.SiteURL + "/aboutme",
		Image:       p.SiteURL + "/public/images/avatar.jpg",
	}
	return pages.AboutMe(aboutme, meta).Render(r.Context(), w)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func toTitle(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
