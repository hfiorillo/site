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
	"strings"
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
	router.Get(paths.RouteCoords, handler.Make(page.HandleRouteCoords))
	for _, tc := range []struct {
		name, path   string
		status       int
		want, absent string
	}{
		{"listing", paths.Routes, 200, "126 km", "Jan 0001"},
		{"Ireland details", paths.Routes + "/west-coast-of-ireland", 200, "779 km", "Badger_divide_reverse.gpx"},
		{"Bilbao without GPX", paths.Routes + "/bilbao-to-san-sebastian", 200, "126 km", "Download GPX"},
		{"Bilbao coordinates unavailable", paths.RouteCoordsPrefix + "bilbao-to-san-sebastian" + paths.RouteCoordsSuffix, 404, "404", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
			if response.Code != tc.status {
				t.Fatalf("status %d", response.Code)
			}
			body := response.Body.String()
			if !strings.Contains(body, tc.want) {
				t.Errorf("missing %q", tc.want)
			}
			if tc.absent != "" && strings.Contains(body, tc.absent) {
				t.Errorf("unexpected %q", tc.absent)
			}
		})
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", paths.RouteCoordsPrefix+"west-coast-of-ireland"+paths.RouteCoordsSuffix, nil))
	var coords []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &coords); err != nil {
		t.Fatal(err)
	}
	if len(coords) < 100 {
		t.Fatal("missing Ireland track")
	}
	first, last := coords[0], coords[len(coords)-1]
	if first.Lat < 51 || first.Lat > 52 || last.Lat < 54 || last.Lat > 56 || first.Lon > -8 || last.Lon > -7 {
		t.Fatal("track is not Cork to Derry")
	}
}
