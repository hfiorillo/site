package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

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
