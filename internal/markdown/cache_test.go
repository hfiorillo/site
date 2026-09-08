package markdown

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hfiorillo/site/paths"
)

func TestMissingPostDoesNotBlockLaterLookups(t *testing.T) {
	t.Chdir("../..")
	buildPathMap()
	done := make(chan error, 1)
	go func() {
		if _, err := LoadMarkdownPost(context.Background(), paths.PostsMarkdownPrefix+"definitely-not-a-real-post"); err == nil {
			done <- os.ErrExist
			return
		}
		_, err := LoadMarkdownPost(context.Background(), paths.PostsMarkdownPrefix+"west-coast-ireland")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("post lookup deadlocked")
	}
}

func TestParsedCacheReuseAndInvalidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "post.md")
	content := []byte("---\ntitle: Before\ndate: 2026-03-01\npublished: true\n---\nHello")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	first, err := loadParsedPost(context.Background(), path, "post")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cached, err := loadParsedPost(context.Background(), path, "post")
			if err != nil || cached != first {
				t.Error("cache miss for unchanged content")
			}
		}()
	}
	wg.Wait()
	if err := os.WriteFile(path, []byte("---\ntitle: After editing\ndate: 2026-03-01\npublished: true\n---\nUpdated"), 0600); err != nil {
		t.Fatal(err)
	}
	updated, err := loadParsedPost(context.Background(), path, "post")
	if err != nil {
		t.Fatal(err)
	}
	if updated == first || updated.Title != "After editing" {
		t.Fatal("stale cached content")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loadParsedPost(ctx, path, "post"); err != context.Canceled {
		t.Fatal("cancelled load did not stop")
	}
}

func TestListingAndDetailShareParsedPost(t *testing.T) {
	t.Chdir("../..")
	buildPathMap()
	posts, err := LoadMarkdownPosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) == 0 {
		t.Fatal("no posts")
	}
	post, err := LoadMarkdownPost(context.Background(), paths.PostsMarkdownPrefix+posts[0].Filename)
	if err != nil {
		t.Fatal(err)
	}
	if post != posts[0] {
		t.Fatal("detail reparsed a cached listing post")
	}
}
