package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hfiorillo/site/internal/gpx"
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

func (p PageHandler) HandleBlogPage(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	filter := r.URL.Query().Get("category")
	filterType := "category"
	if filter == "" {
		filter = r.URL.Query().Get("tag")
		filterType = "tag"
	}

	if filter != "" {
		var filtered []*models.BlogPost
		for _, post := range posts {
			var items []string
			if filterType == "category" {
				items = post.Metadata.Categories
			} else {
				items = post.Metadata.Tags
			}
			for _, item := range items {
				if strings.EqualFold(item, filter) {
					filtered = append(filtered, post)
					break
				}
			}
		}
		posts = filtered
	}

	title := "Blog | Harry Fiorillo-Hughes"
	if filter != "" {
		title = strings.Title(filter) + " | Harry Fiorillo-Hughes"
	}

	meta := models.PageMeta{
		Title:       title,
		Description: "Recent posts about DevOps, engineering, and adventures.",
		URL:         p.SiteURL + "/blog",
		Image:       p.SiteURL + "/public/images/avatar.jpg",
	}

	var recent, old []*models.BlogPost
	if filter != "" {
		recent = posts
	} else {
		cutoff := time.Now().AddDate(-5, 0, 0)
		for _, post := range posts {
			if post.Date.After(cutoff) {
				recent = append(recent, post)
			} else {
				old = append(old, post)
			}
		}
	}

	return pages.Blog(posts, filter, meta, recent, old).Render(r.Context(), w)
}

func (p PageHandler) HandleBlogPostPage(w http.ResponseWriter, r *http.Request) error {
	filename := chi.URLParam(r, "filename")
	post, err := markdown.LoadMarkdownPost(fmt.Sprintf("/posts/%s", filename))
	if err != nil {
		return pages.ErrorPage("Tis page does not exist.").Render(r.Context(), w)
	}

	if !post.Metadata.Published {
		return pages.ErrorPage("Tis page does not exist.").Render(r.Context(), w)
	}

	allPosts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		allPosts = nil
	}

	var prev, next *models.BlogPost
	for i, p := range allPosts {
		if p.Filename == filename {
			if i > 0 {
				next = allPosts[i-1]
			}
			if i < len(allPosts)-1 {
				prev = allPosts[i+1]
			}
			break
		}
	}

	image := p.SiteURL + "/public/images/avatar.jpg"
	if post.Metadata.Image != "" {
		image = p.SiteURL + post.Metadata.Image
	}

	meta := models.PageMeta{
		Title:       post.Title + " | Harry Fiorillo-Hughes",
		Description: post.Description,
		URL:         p.SiteURL + "/blog/" + post.Filename,
		Image:       image,
	}
	return pages.BlogPage(post, prev, next, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleFeed(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">` + "\n")
	buf.WriteString("<channel>\n")
	buf.WriteString("<title>Harry Fiorillo-Hughes | Blog</title>\n")
	buf.WriteString(fmt.Sprintf("<link>%s</link>\n", p.SiteURL))
	buf.WriteString("<description>Personal blog about DevOps, engineering, and adventures</description>\n")
	buf.WriteString(fmt.Sprintf("<atom:link href=\"%s/feed.xml\" rel=\"self\" type=\"application/rss+xml\"/>\n", p.SiteURL))
	buf.WriteString("<language>en</language>\n")

	for _, post := range posts {
		buf.WriteString("<item>\n")
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", xmlEscape(post.Title)))
		buf.WriteString(fmt.Sprintf("<link>%s/blog/%s</link>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<guid>%s/blog/%s</guid>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<description>%s</description>\n", xmlEscape(post.Description)))
		buf.WriteString(fmt.Sprintf("<pubDate>%s</pubDate>\n", post.Date.Format(time.RFC822)))
		buf.WriteString("</item>\n")
	}

	buf.WriteString("</channel>\n")
	buf.WriteString("</rss>\n")

	w.Write(buf.Bytes())
	return nil
}

func (p PageHandler) HandleSitemap(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		posts = nil
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	dateFormat := "2006-01-02"

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>monthly</changefreq>\n")
	buf.WriteString("<priority>1.0</priority>\n")
	buf.WriteString("</url>\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/blog</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>weekly</changefreq>\n")
	buf.WriteString("<priority>0.8</priority>\n")
	buf.WriteString("</url>\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/aboutme</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>monthly</changefreq>\n")
	buf.WriteString("<priority>0.6</priority>\n")
	buf.WriteString("</url>\n")

	for _, post := range posts {
		buf.WriteString("<url>\n")
		buf.WriteString(fmt.Sprintf("<loc>%s/blog/%s</loc>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", post.Date.Format(dateFormat)))
		buf.WriteString("<changefreq>never</changefreq>\n")
		buf.WriteString("<priority>0.6</priority>\n")
		buf.WriteString("</url>\n")
	}

	buf.WriteString("</urlset>\n")

	w.Write(buf.Bytes())
	return nil
}

func (p PageHandler) HandleRoutes(w http.ResponseWriter, r *http.Request) error {
	rd, err := gpx.Parse("./public/routes/Badger_divide_reverse.gpx")
	if err != nil {
		return err
	}
	coordsJSON, err := gpx.CoordsToJSON(rd.Coords)
	if err != nil {
		return err
	}

	route := &models.Route{
		Name:          "Badger Divide (Reverse)",
		Location:      "Glasgow to Inverness, Scotland",
		DistanceKm:    math.Round(rd.DistanceKm),
		ElevationGain: math.Round(rd.ElevationGain),
		ElevationMax:  math.Round(rd.ElevationMax),
		ElevationMin:  math.Round(rd.ElevationMin),
		Date:          time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		CoordsJSON:    coordsJSON,
		GPXFile:       "/public/routes/Badger_divide_reverse.gpx",
	}

	meta := models.PageMeta{
		Title:       "Routes | Harry Fiorillo-Hughes",
		Description: "Bikepacking and cycling routes.",
		URL:         p.SiteURL + "/routes",
		Image:       p.SiteURL + "/public/images/avatar.jpg",
	}
	return pages.Routes([]*models.Route{route}, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleRoute(w http.ResponseWriter, r *http.Request) error {
	rd, err := gpx.Parse("./public/routes/Badger_divide_reverse.gpx")
	if err != nil {
		return err
	}
	coordsJSON, err := gpx.CoordsToJSON(rd.Coords)
	if err != nil {
		return err
	}

	route := &models.Route{
		Name:          "Badger Divide (Reverse)",
		Location:      "Glasgow to Inverness, Scotland",
		DistanceKm:    math.Round(rd.DistanceKm),
		ElevationGain: math.Round(rd.ElevationGain),
		ElevationMax:  math.Round(rd.ElevationMax),
		ElevationMin:  math.Round(rd.ElevationMin),
		Date:          time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		CoordsJSON:    coordsJSON,
		GPXFile:       "/public/routes/Badger_divide_reverse.gpx",
	}

	meta := models.PageMeta{
		Title:       route.Name + " | Routes | Harry Fiorillo-Hughes",
		Description: route.Location,
		URL:         p.SiteURL + "/routes/badger-divide",
		Image:       p.SiteURL + "/public/images/avatar.jpg",
	}
	return pages.RoutePage(route, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleRouteCoords(w http.ResponseWriter, r *http.Request) error {
	rd, err := gpx.Parse("./public/routes/Badger_divide_reverse.gpx")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(rd.Coords)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
