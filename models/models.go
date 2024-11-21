package models

import (
	"html/template"
	"time"
)

type BlogPost struct {
	Filename    string
	Title       string
	Slug        string
	Description string
	Date        time.Time
	Content     template.HTML
	Headers     Headers
	Metadata    Metadata
	Order       int
	// HtmlContent []byte
	// Parent      string
	// TLDR        template.HTML
}

// Metadata represents the YAML front matter
type Metadata struct {
	Title      string   `yaml:"title"`
	Date       string   `yaml:"date"`
	Categories []string `yaml:"categories"`
	Tags       []string `yaml:"tags"`
	Classes    string   `yaml:"classes"`
	TOC        bool     `yaml:"toc"`
	Header     struct {
		OverlayImage  string `yaml:"overlay_image"`
		OverlayFilter string `yaml:"overlay_filter"`
	} `yaml:"header"`
	Published   bool   `yaml:"published"`
	Description string `yaml:"description"`
}

// Headers within headers
type Headers struct {
	TopLevel []struct { // #
		Headers  string
		MidLevel []struct { // ##
			Headers     string
			BottomLevel []struct { // ###
				Headers string
			}
		}
	}
}

type SidebarData struct {
	Categories []Category
}

type Category struct {
	Name  string
	Pages []BlogPost
	Order int
}
