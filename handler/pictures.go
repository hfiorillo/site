package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/view/pages"
	"net/http"
	"os"
	"sync"
)

var (
	galleryOnce  sync.Once
	galleryYears []models.GalleryYear
	galleryErr   error
)

func loadGallery() {
	data, err := os.ReadFile(paths.GalleryManifest)
	if err != nil {
		galleryErr = fmt.Errorf("reading gallery: %w", err)
		return
	}
	if err := json.Unmarshal(data, &galleryYears); err != nil {
		galleryErr = fmt.Errorf("parsing gallery: %w", err)
		return
	}
	// Do not expose an unpublished post through an older generated manifest.
	posts, err := markdown.LoadMarkdownPosts(context.Background())
	if err != nil {
		galleryErr = err
		return
	}
	published := make(map[string]bool, len(posts))
	for _, post := range posts {
		published[paths.Blog+"/"+post.Filename] = true
	}
	visible := []models.GalleryYear{}
	for _, group := range galleryYears {
		photos := []models.GalleryPhoto{}
		for _, photo := range group.Photos {
			if published[photo.PostURL] {
				photos = append(photos, photo)
			}
		}
		if len(photos) > 0 {
			group.Photos = photos
			visible = append(visible, group)
		}
	}
	galleryYears = visible

}

func (p PageHandler) HandlePictures(w http.ResponseWriter, r *http.Request) error {
	galleryOnce.Do(loadGallery)
	if galleryErr != nil {
		return galleryErr
	}
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
	return pages.Pictures(pictures, galleryYears, meta).Render(r.Context(), w)
}
