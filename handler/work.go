package handler

import (
	"fmt"
	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/view/pages"
	"net/http"
)

func (p PageHandler) HandleWork(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	work, err := markdown.LoadMarkdownPost(r.Context(), paths.WorkMarkdown)
	if err != nil {
		p.Logger.Error("failed to load work", "err", err, "path", r.URL.Path)
		return pages.ErrorPage(fmt.Sprintf("%v", err)).Render(r.Context(), w)
	}

	posts, err := markdown.LoadMarkdownPosts(r.Context())
	if err != nil {
		return err
	}
	var workPosts []*models.BlogPost
	for _, post := range posts {
		if post.Metadata.Section == models.WorkSection {
			workPosts = append(workPosts, post)
		}
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
	return pages.Work(work, workPosts, meta).Render(r.Context(), w)
}
