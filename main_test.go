package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
	"github.com/hfiorillo/site/handler"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/view/layout"
)

func TestStylesheetURL(t *testing.T) {
	assetPath := strings.TrimPrefix(paths.Styles, "/")
	original := fstest.MapFS{assetPath: &fstest.MapFile{Data: []byte("body{color:black}")}}
	baseline, err := stylesheetURL(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name             string
		assets           fs.FS
		changed, missing bool
	}{
		{name: "same CSS preserves cache key", assets: original},
		{name: "changed CSS invalidates cache", assets: fstest.MapFS{assetPath: &fstest.MapFile{Data: []byte("body{color:green}")}}, changed: true},
		{name: "missing CSS fails startup", assets: fstest.MapFS{}, missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := stylesheetURL(tc.assets)
			if tc.missing {
				if !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("expected missing file error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(got)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Path != paths.Styles || parsed.Query().Get("v") == "" {
				t.Fatalf("invalid stylesheet URL: %s", got)
			}
			if (got != baseline) != tc.changed {
				t.Fatalf("unexpected cache key: %s", got)
			}
		})
	}
}

func TestVersionedStylesheetRenderedAndServed(t *testing.T) {
	versioned, err := stylesheetURL(publicFS)
	if err != nil {
		t.Fatal(err)
	}
	previous := paths.StylesURL
	paths.StylesURL = versioned
	t.Cleanup(func() { paths.StylesURL = previous })
	var html bytes.Buffer
	if err := layout.Base(models.PageMeta{}).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), `href="`+versioned+`"`) {
		t.Fatal("layout omitted fingerprinted stylesheet URL")
	}
	response := httptest.NewRecorder()
	public().ServeHTTP(response, httptest.NewRequest(http.MethodGet, versioned, nil))
	expected, err := publicFS.ReadFile(strings.TrimPrefix(paths.Styles, "/"))
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), expected) {
		t.Fatal("versioned URL did not serve embedded stylesheet")
	}
}

func TestWorkPostListings(t *testing.T) {
	page := handler.NewPageHandler(slog.Default(), "https://example.com")
	engineering := []string{"FirstRestApiAzure", "kubernetes-ingress", "kubernetes-pi", "raspberry-pi", "kubernetes-monitoring"}
	for _, tc := range []struct {
		name, path                  string
		handle                      func(http.ResponseWriter, *http.Request) error
		wantEngineering, wantTravel bool
	}{
		{"work", paths.Work, page.HandleWork, true, false},
		{"home", paths.Root, page.HandleIndexPage, true, true},
		{"blog", paths.Blog, page.HandleBlogPage, true, true},
		{"feed", paths.Feed, page.HandleFeed, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			if err := tc.handle(response, httptest.NewRequest(http.MethodGet, tc.path, nil)); err != nil {
				t.Fatal(err)
			}
			body := response.Body.String()
			for _, slug := range engineering {
				if strings.Contains(body, paths.Blog+"/"+slug) != tc.wantEngineering {
					t.Errorf("unexpected listing for %s", slug)
				}
			}
			if strings.Contains(body, paths.Blog+"/west-coast-ireland") != tc.wantTravel {
				t.Error("unexpected travel post listing")
			}
			if strings.Contains(body, paths.Blog+"/building-blog-pt1") {
				t.Error("unpublished draft exposed")
			}
		})
	}
}

func TestRouteUpdates(t *testing.T) {
	page := handler.NewPageHandler(slog.Default(), "https://example.com")
	router := chi.NewRouter()
	router.Get(paths.Routes, handler.Make(page.HandleRoutes))
	router.Get(paths.RouteDetail, handler.Make(page.HandleRoute))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", paths.Routes, nil))
	body := response.Body.String()
	for _, wanted := range []string{"126 km", "779 km", "348 km", "https://www.komoot.com/tour/2765576526", "https://www.komoot.com/tour/3262934167", "https://www.komoot.com/tour/3268174732"} {
		if !strings.Contains(body, wanted) {
			t.Errorf("missing %q", wanted)
		}
	}
	for _, old := range []string{".gpx", "leaflet", "route-map", "Download"} {
		if strings.Contains(body, old) {
			t.Errorf("retired route feature %q", old)
		}
	}
	for _, slug := range []string{"badger-divide", "west-coast-of-ireland", "bilbao-to-san-sebastian"} {
		t.Run(slug, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", paths.Routes+"/"+slug, nil))
			if response.Code != http.StatusTemporaryRedirect || !strings.HasPrefix(response.Header().Get("Location"), "https://www.komoot.com/tour/") {
				t.Fatal("missing Komoot redirect")
			}
		})
	}
}

func TestPicturesGallery(t *testing.T) {
	data, err := os.ReadFile(paths.GalleryManifest)
	if err != nil {
		t.Fatal(err)
	}
	var years []models.GalleryYear
	if err := json.Unmarshal(data, &years); err != nil {
		t.Fatal(err)
	}
	if len(years) == 0 {
		t.Fatal("empty gallery")
	}
	page := handler.NewPageHandler(slog.Default(), "https://example.com")
	response := httptest.NewRecorder()
	if err := page.HandlePictures(response, httptest.NewRequest("GET", paths.Pictures, nil)); err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	seen := map[string]bool{}
	for i, year := range years {
		if i > 0 && year.Year > years[i-1].Year {
			t.Fatal("years are not newest first")
		}
		for _, photo := range year.Photos {
			if seen[photo.Original] {
				t.Fatal("duplicate photo")
			}
			seen[photo.Original] = true
			if photo.Width <= 0 || photo.Height <= 0 {
				t.Fatal("missing layout dimensions")
			}
			if !strings.Contains(body, photo.Preview) || !strings.Contains(body, photo.PostURL) {
				t.Fatal("photo or source link missing")
			}
			if strings.Contains(body, `src="`+photo.Original+`"`) {
				t.Fatal("gallery eagerly references full-size photo")
			}
			for _, src := range []string{photo.Preview} {
				if _, err := fs.Stat(publicFS, strings.TrimPrefix(src, "/")); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	entries, err := fs.ReadDir(publicFS, strings.TrimSuffix(strings.TrimPrefix(paths.GalleryAssets, "/"), "/"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(seen) {
		t.Fatalf("expected one preview per photo, got %d files for %d photos", len(entries), len(seen))
	}
	photo := years[0].Photos[0]
	cached := httptest.NewRecorder()
	public().ServeHTTP(cached, httptest.NewRequest("GET", photo.Preview, nil))
	if cached.Code != 200 || cached.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatal("preview not cached immutably")
	}
	missing := httptest.NewRecorder()
	public().ServeHTTP(missing, httptest.NewRequest("GET", paths.GalleryAssets+"missing.webp", nil))
	if missing.Code != 404 || missing.Header().Get("Cache-Control") != "" {
		t.Fatal("missing preview cached")
	}
	if strings.Count(body, `loading="lazy"`) < len(seen) {
		t.Fatal("gallery photos must lazy-load")
	}
}

func TestConcurrentKomootRoutes(t *testing.T) {
	page := handler.NewPageHandler(slog.Default(), "https://example.com")
	router := chi.NewRouter()
	router.Get(paths.Routes, handler.Make(page.HandleRoutes))
	router.Get(paths.RouteDetail, handler.Make(page.HandleRoute))
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, path := range []string{paths.Routes, paths.Routes + "/badger-divide"} {
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
				if response.Code != 200 && response.Code != 307 {
					t.Errorf("unexpected route status: %d", response.Code)
				}
			}
		}()
	}
	wg.Wait()
}
