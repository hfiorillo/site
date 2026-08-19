package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
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
	aboutme, err := markdown.LoadMarkdownPost(r.Context(), paths.AboutMeMarkdown)
	if err != nil {
		p.Logger.Error("failed to load aboutme", "err", err, "path", r.URL.Path)
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	siteOnce.Do(loadSiteMeta)
	meta := models.PageMeta{
		Title:          siteMeta.Title,
		Description:    siteMeta.Description,
		URL:            p.SiteURL + paths.AboutMe,
		Canonical:      p.SiteURL + paths.AboutMe,
		Image:          p.SiteURL + siteImage(),
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.AboutMe(aboutme, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleAboutThisSite(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	about, err := markdown.LoadMarkdownPost(r.Context(), paths.AboutThisSiteMarkdown)
	if err != nil {
		p.Logger.Error("failed to load about-this-site", "err", err, "path", r.URL.Path)
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if about.Metadata.Image != "" {
		image = p.SiteURL + about.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          about.Title + " | Harry Fiorillo-Hughes",
		Description:    about.Description,
		URL:            p.SiteURL + paths.AboutThisSite,
		Canonical:      p.SiteURL + paths.AboutThisSite,
		Image:          image,
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.AboutThisSite(about, meta).Render(r.Context(), w)
}

func (p PageHandler) HandlePictures(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	pictures, err := markdown.LoadMarkdownPost(r.Context(), paths.PicturesMarkdown)
	if err != nil {
		p.Logger.Error("failed to load pictures", "err", err, "path", r.URL.Path)
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if pictures.Metadata.Image != "" {
		image = p.SiteURL + pictures.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          pictures.Title + " | Harry Fiorillo-Hughes",
		Description:    pictures.Description,
		URL:            p.SiteURL + paths.Pictures,
		Canonical:      p.SiteURL + paths.Pictures,
		Image:          image,
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.Pictures(pictures, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleWork(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	work, err := markdown.LoadMarkdownPost(r.Context(), paths.WorkMarkdown)
	if err != nil {
		p.Logger.Error("failed to load work", "err", err, "path", r.URL.Path)
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	image := p.SiteURL + siteImage()
	if work.Metadata.Image != "" {
		image = p.SiteURL + work.Metadata.Image
	}

	meta := models.PageMeta{
		Title:          work.Title + " | Harry Fiorillo-Hughes",
		Description:    work.Description,
		URL:            p.SiteURL + paths.Work,
		Canonical:      p.SiteURL + paths.Work,
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
