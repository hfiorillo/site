package pkg

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hfiorillo/site/models"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func LoadMarkdownPosts(dir string) ([]models.BlogPost, error) {
	var posts []models.BlogPost
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
func ParseMarkdownFile(file []byte) (models.BlogPost, error) {
	sections := strings.SplitN(string(file), "---", 2)
	if len(sections) < 2 {
		return models.BlogPost{}, errors.New("invalid markdown format")
	}

	metadata := sections[0]
	mdContent := sections[1]

	// deal with rogue \r's
	metadata = strings.ReplaceAll(metadata, "\r", "")
	mdContent = strings.ReplaceAll(mdContent, "\r", "")

	title, slug, parent, description, order, metaDescriptionStr,
		metaPropertyTitleStr, metaPropertyDescriptionStr,
		metaOgURLStr := parseMetadata(metadata)

	htmlContent := MdToHTML([]byte(mdContent))
	headers := ExtractHeaders([]byte(mdContent))

	return models.BlogPost{
		Title:                   title,
		Slug:                    slug,
		Parent:                  parent,
		Description:             description,
		Content:                 template.HTML(htmlContent),
		HtmlContent:             htmlContent,
		Headers:                 headers,
		Order:                   order,
		MetaDescription:         metaDescriptionStr,
		MetaPropertyTitle:       metaPropertyTitleStr,
		MetaPropertyDescription: metaPropertyDescriptionStr,
		MetaOgURL:               metaOgURLStr,
	}, nil
}

func parseMetadata(metadata string) (
	title string,
	slug string,
	parent string,
	description string,
	order int,
	metaDescription string,
	metaPropertyTitle string,
	metaPropertyDescription string,
	metaOgURL string,
) {
	re := regexp.MustCompile(`(?m)^(\w+):\s*(.+)`)
	matches := re.FindAllStringSubmatch(metadata, -1)

	metaDataMap := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			metaDataMap[match[1]] = match[2]
		}
	}

	title = metaDataMap["Title"]
	slug = metaDataMap["Slug"]
	parent = metaDataMap["Parent"]
	description = metaDataMap["Description"]
	orderStr := metaDataMap["Order"]
	metaDescriptionStr := metaDataMap["MetaDescription"]
	metaPropertyTitleStr := metaDataMap["MetaPropertyTitle"]
	metaPropertyDescriptionStr := metaDataMap["MetaPropertyDescription"]
	metaOgURLStr := metaDataMap["MetaOgURL"]

	order, err := strconv.Atoi(orderStr)
	if err != nil {
		order = 9999 // set this to a high number in case of err
	}

	return title, slug, parent, description, order, metaDescriptionStr,
		metaPropertyTitleStr, metaPropertyDescriptionStr, metaOgURLStr
}

func MdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)

	doc := parser.Parse(md)

	output := markdown.Render(doc, renderer)

	// output2 := goldmark.Convert()

	return output
}

func ExtractHeaders(content []byte) []string {
	var headers []string
	//match only level 2 markdown headers
	re := regexp.MustCompile(`(?m)^##\s+(.*)`)
	matches := re.FindAllSubmatch(content, -1)

	for _, match := range matches {
		// match[1] contains header text without the '##'
		headers = append(headers, string(match[1]))
	}

	return headers
}
