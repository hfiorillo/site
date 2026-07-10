package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/view/pages"
)

func (p PageHandler) HandleBlogPage(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	posts, err := markdown.LoadMarkdownPosts(r.Context())
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

	title := siteMeta.Blog.Title + " | " + siteMeta.Title
	if filter != "" {
		title = toTitle(filter) + " | " + siteMeta.Title
	}

	meta := models.PageMeta{
		Title:          title,
		Description:    siteMeta.Blog.Description,
		URL:            p.SiteURL + "/blog",
		Canonical:      p.SiteURL + "/blog",
		Image:          p.SiteURL + siteImage(),
		StructuredData: personJSON(p.SiteURL),
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
	post, err := markdown.LoadMarkdownPost(r.Context(), fmt.Sprintf("/posts/%s", filename))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("This page does not exist.").Render(r.Context(), w)
	}

	if !post.Metadata.Published {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("This page does not exist.").Render(r.Context(), w)
	}

	allPosts, err := markdown.LoadMarkdownPosts(r.Context())
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

	siteOnce.Do(loadSiteMeta)
	image := p.SiteURL + siteImage()
	if post.Metadata.Image != "" {
		image = p.SiteURL + post.Metadata.Image
	}

	blogURL := p.SiteURL + "/blog/" + post.Filename
	sd := map[string]any{
		"@context":       "https://schema.org",
		"@type":          "BlogPosting",
		"headline":       post.Title,
		"datePublished":  post.Date.Format("2006-01-02"),
		"dateModified":   post.Date.Format("2006-01-02"),
		"author":         map[string]string{"@type": "Person", "name": "Harry Fiorillo"},
		"description":    post.Description,
		"mainEntityOfPage": map[string]string{"@type": "WebPage", "@id": blogURL},
	}
	sdJSON, _ := json.Marshal(sd)

	meta := models.PageMeta{
		Title:          post.Title + " | Harry Fiorillo-Hughes",
		Description:    post.Description,
		URL:            blogURL,
		Canonical:      blogURL,
		Image:          image,
		StructuredData: string(sdJSON),
	}
	return pages.BlogPage(post, prev, next, meta).Render(r.Context(), w)
}
