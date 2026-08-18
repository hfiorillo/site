package markdown

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hfiorillo/site/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

const (
	contentDir        = "./content"
	postsDir          = contentDir + "/posts"
	avgWordsPerMinute = 200
	markdownExtension = ".md"
	dateFormat        = "2006-01-02"
)

var (
	parserInst    goldmark.Markdown
	imgAttrRegex  = regexp.MustCompile(`<img\s`)
	postsCache    []*models.BlogPost
	postsMu       sync.RWMutex
	postsCacheAt  time.Time
	cacheTTL      = 60 * time.Second
	postsPathMu   sync.RWMutex
	postsPathMap  map[string]string
)

func init() {
	parserInst = goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	buildPathMap()
}

func buildPathMap() {
	m := make(map[string]string)
	filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), markdownExtension) {
			name := strings.TrimSuffix(d.Name(), markdownExtension)
			m[name] = path
		}
		return nil
	})
	postsPathMu.Lock()
	postsPathMap = m
	postsPathMu.Unlock()
}

func findPostPath(targetBase string) (string, error) {
	postsPathMu.RLock()
	path, ok := postsPathMap[targetBase]
	postsPathMu.RUnlock()
	if ok {
		return path, nil
	}

	postsPathMu.Lock()
	buildPathMap()
	path = postsPathMap[targetBase]
	postsPathMu.Unlock()
	if path == "" {
		return "", fmt.Errorf("post '%s' not found", targetBase)
	}
	return path, nil
}

func calculateReadTime(text string) int {
	wordCount := len(strings.Fields(text))
	if wordCount == 0 {
		return 0
	}
	return int(math.Ceil(float64(wordCount) / float64(avgWordsPerMinute)))
}

func LoadMarkdownPost(ctx context.Context, fileName string) (*models.BlogPost, error) {
	targetBase := filepath.Base(fileName)

	path, err := findPostPath(targetBase)
	if err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	post, err := ParseMarkdown(content, targetBase)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", fileName, err)
	}

	return post, nil
}

func LoadMarkdownPosts(ctx context.Context) ([]*models.BlogPost, error) {
	postsMu.RLock()
	if postsCache != nil && time.Since(postsCacheAt) < cacheTTL {
		defer postsMu.RUnlock()
		return postsCache, nil
	}
	postsMu.RUnlock()

	postsMu.Lock()
	defer postsMu.Unlock()

	if postsCache != nil && time.Since(postsCacheAt) < cacheTTL {
		return postsCache, nil
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	posts, err := loadPostsUncached(ctx)
	if err != nil {
		return nil, err
	}

	postsCache = posts
	postsCacheAt = time.Now()
	return posts, nil
}

func loadPostsUncached(ctx context.Context) ([]*models.BlogPost, error) {
	var posts []*models.BlogPost

	err := filepath.WalkDir(postsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), markdownExtension) {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Warning: could not read file %s: %v\n", path, err)
			return nil
		}

		filenameWithoutExt := strings.TrimSuffix(d.Name(), markdownExtension)
		post, err := ParseMarkdown(content, filenameWithoutExt)
		if err != nil {
			fmt.Printf("Warning: could not parse markdown for file %s: %v\n", d.Name(), err)
			return nil
		}

		if !post.Metadata.Published {
			return nil
		}

		posts = append(posts, post)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading posts: %w", err)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

func ParseMarkdown(fileContent []byte, filename string) (*models.BlogPost, error) {
	if !bytes.HasPrefix(fileContent, []byte("---")) {
		return nil, errors.New("invalid markdown format: could not find metadata separator '---' at the beginning of the file")
	}

	var contentHTML bytes.Buffer
	ctx := parser.NewContext()
	if err := parserInst.Convert(fileContent, &contentHTML, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("converting markdown for '%s': %w", filename, err)
	}

	metaData := meta.Get(ctx)
	if metaData == nil {
		return nil, errors.New("no metadata found in markdown file")
	}

	metadata := mapMetaToMetadata(metaData)

	postDate, err := time.Parse(dateFormat, metadata.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date '%s' for '%s': %w", metadata.Date, filename, err)
	}

	html := addImgAttrs(contentHTML.String())

	headers, err := ParseHeaders(fileContent)
	if err != nil {
		return nil, fmt.Errorf("parsing headers for '%s': %w", filename, err)
	}

	return &models.BlogPost{
		Filename:        filename,
		Title:           metadata.Title,
		Date:            postDate,
		Description:     metadata.Description,
		Content:         template.HTML(html),
		Metadata:        metadata,
		ReadTimeMinutes: calculateReadTime(string(fileContent)),
		Headers:         headers,
	}, nil
}

func mapMetaToMetadata(metaData map[string]interface{}) models.Metadata {
	var m models.Metadata
	if v, ok := metaData["title"]; ok {
		m.Title = toString(v)
	}
	if v, ok := metaData["date"]; ok {
		m.Date = toString(v)
	}
	if v, ok := metaData["description"]; ok {
		m.Description = toString(v)
	}
	if v, ok := metaData["published"]; ok {
		m.Published = toBool(v)
	}
	if v, ok := metaData["categories"]; ok {
		m.Categories = toStringSlice(v)
	}
	if v, ok := metaData["tags"]; ok {
		m.Tags = toStringSlice(v)
	}
	if v, ok := metaData["image"]; ok {
		m.Image = toString(v)
	}
	if v, ok := metaData["preview-image"]; ok {
		m.PreviewImage = toString(v)
	}
	return m
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func toStringSlice(v interface{}) []string {
	switch s := v.(type) {
	case []interface{}:
		result := make([]string, len(s))
		for i, item := range s {
			result[i] = toString(item)
		}
		return result
	case []string:
		return s
	default:
		return nil
	}
}

func addImgAttrs(html string) string {
	return imgAttrRegex.ReplaceAllString(html, `<img loading="lazy" decoding="async" `)
}

func Slugify(text string) string {
	var buf strings.Builder
	lower := strings.ToLower(text)
	prevHyphen := false

	for _, r := range lower {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' {
			buf.WriteRune(r)
			prevHyphen = false
		} else if r == ' ' || r == '-' {
			if !prevHyphen && buf.Len() > 0 {
				buf.WriteRune('-')
				prevHyphen = true
			}
		}
	}

	result := strings.Trim(buf.String(), "-")
	if result == "" {
		return "heading"
	}
	return result
}
