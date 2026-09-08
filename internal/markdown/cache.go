package markdown

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hfiorillo/site/models"
)

type parsedPost struct {
	post     *models.BlogPost
	modified time.Time
	size     int64
}

var parsedMu sync.Mutex
var parsedPosts = map[string]parsedPost{}

// Cached posts are immutable. Listings and detail pages share the same parsed value.
// File metadata invalidates changed content without reading and parsing it each visit.
func loadParsedPost(ctx context.Context, path, filename string) (*models.BlogPost, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsedMu.Lock()
	defer parsedMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("checking %s: %w", path, err)
	}
	if cached, ok := parsedPosts[path]; ok && cached.modified.Equal(info.ModTime()) && cached.size == info.Size() {
		return cached.post, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	post, err := ParseMarkdown(data, filename)
	if err != nil {
		return nil, err
	}
	parsedPosts[path] = parsedPost{post: post, modified: info.ModTime(), size: info.Size()}
	return post, nil
}
