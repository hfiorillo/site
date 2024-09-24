package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/hfiorillo/site/models"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v2"
)

func LoadMarkdownPosts(dir string) ([]*models.BlogPost, error) {
	var posts []*models.BlogPost
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			path := dir + "/" + file.Name()
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			// fmt.Println(file)
			post, err := ParseMarkdownFile(content)
			if err != nil {
				return nil, fmt.Errorf("failed parsing markdown file: %w", err)
			}

			posts = append(posts, post)
		}
	}

	return posts, nil
}

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
	md := goldmark.New()
	if err := md.Convert(sections[1], &buf); err != nil {
		return nil, err
	}

	// Populate BlogPost struct
	blogPost := &models.BlogPost{
		Title:       metadata.Title,
		Description: metadata.Description,
		HtmlContent: buf.Bytes(),
		Content:     template.HTML(buf.String()), // For use in templates
		Metadata:    metadata,
	}

	// Parse headers (H2) from the markdown content
	blogPost.Headers = extractHeaders(sections[1])

	return blogPost, nil
}

// Function to extract headers (h2) from the markdown content
func extractHeaders(content []byte) []string {
	lines := strings.Split(string(content), "\n")
	var headers []string
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") { // Assuming H2 headers start with "## "
			headers = append(headers, strings.TrimPrefix(line, "## "))
		}
	}
	return headers
}
