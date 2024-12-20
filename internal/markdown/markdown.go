package markdown

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hfiorillo/site/models"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"gopkg.in/yaml.v2"
)

const contentDir string = "./content"

// Returns markdown parser with extensions
func NewGoldMarkParser() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
	)
}

// Loads a given markdown post from the posts directory
func LoadMarkdownPost(fileName string) (*models.BlogPost, error) {

	path := contentDir + fileName + ".md"
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	post, err := ParseMarkdownFile(content)
	if err != nil {
		return nil, err
	}

	post.Filename = fileName

	return post, nil
}

// Loads all the markdown posts in the posts directory
func LoadMarkdownPosts() ([]*models.BlogPost, error) {
	var posts []*models.BlogPost

	postsDir := contentDir + "/posts"

	files, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			path := postsDir + "/" + file.Name()
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			post, err := ParseMarkdownFile(content)
			if err != nil {
				return nil, err
			}

			post.Filename = strings.Trim(file.Name(), ".md")

			posts = append(posts, post)
		}
	}

	// TODO: maintain an order based on dates
	// Sort posts by date
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })

	return posts, nil
}

// TODO: refactor
// Accepts markdown file - parses and returns a BlogPost
func ParseMarkdownFile(file []byte) (*models.BlogPost, error) {
	sections := bytes.SplitN(file, []byte("---"), 2)
	if len(sections) < 2 {
		return nil, errors.New("invalid markdown format")
	}

	// Parse metadata
	var metadata models.Metadata
	if err := yaml.Unmarshal(sections[0], &metadata); err != nil {
		return nil, err
	}

	// Convert markdown to HTML using goldmark
	var buf bytes.Buffer
	md := NewGoldMarkParser()
	if err := md.Convert(sections[1], &buf); err != nil {
		return nil, err
	}

	// TODO: fix date
	date, err := time.Parse("2006-01-02", metadata.Date)
	if err != nil {
		return nil, err
	}

	// Piece together blog post
	blogPost := &models.BlogPost{
		Title:       metadata.Title,
		Date:        date,
		Description: metadata.Description,
		Content:     template.HTML(buf.String()), // For use in templates
		Metadata:    metadata,
		Headers:     parseHeaders(sections[1]),
	}

	return blogPost, nil
}
