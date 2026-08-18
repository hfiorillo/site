package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/view/pages"
)

func personJSON(siteURL string) string {
	b, _ := json.Marshal(map[string]string{
		"@context": "https://schema.org",
		"@type":    "Person",
		"name":     "Harry Fiorillo",
		"url":      siteURL,
	})
	return string(b)
}

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
	return p.HandleBlogPage(w, r)
}

func (p PageHandler) HandleAboutMePage(w http.ResponseWriter, r *http.Request) error {
	aboutme, err := markdown.LoadMarkdownPost(r.Context(), "/aboutme/aboutme")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	siteOnce.Do(loadSiteMeta)
	meta := models.PageMeta{
		Title:          siteMeta.Title,
		Description:    siteMeta.Description,
		URL:            p.SiteURL + "/aboutme",
		Canonical:      p.SiteURL + "/aboutme",
		Image:          p.SiteURL + siteImage(),
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.AboutMe(aboutme, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleAboutThisSite(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	about, err := markdown.LoadMarkdownPost(r.Context(), "/about-this-site/about-this-site")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if about.Metadata.Image != "" {
		image = p.SiteURL + about.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          about.Title + " | Harry Fiorillo-Hughes",
		Description:    about.Description,
		URL:            p.SiteURL + "/about-this-site",
		Canonical:      p.SiteURL + "/about-this-site",
		Image:          image,
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.AboutThisSite(about, meta).Render(r.Context(), w)
}

func (p PageHandler) HandlePictures(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	pictures, err := markdown.LoadMarkdownPost(r.Context(), "/pictures/pictures")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if pictures.Metadata.Image != "" {
		image = p.SiteURL + pictures.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          pictures.Title + " | Harry Fiorillo-Hughes",
		Description:    pictures.Description,
		URL:            p.SiteURL + "/pictures",
		Canonical:      p.SiteURL + "/pictures",
		Image:          image,
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.Pictures(pictures, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleWork(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	work, err := markdown.LoadMarkdownPost(r.Context(), "/work/work")
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if work.Metadata.Image != "" {
		image = p.SiteURL + work.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          work.Title + " | Harry Fiorillo-Hughes",
		Description:    work.Description,
		URL:            p.SiteURL + "/work",
		Canonical:      p.SiteURL + "/work",
		Image:          image,
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.Work(work, meta).Render(r.Context(), w)
}

func toTitle(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
