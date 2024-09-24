package models

import (
	"html/template"
)

type BlogPost struct {
	Filename    string
	Title       string
	Slug        string
	Parent      string
	Content     template.HTML
	TLDR        template.HTML
	HtmlContent []byte
	Description string
	Order       int
	Headers     []string // for page h2's
	Metadata    Metadata
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

type SidebarData struct {
	Categories []Category
}

type Category struct {
	Name  string
	Pages []BlogPost
	Order int
}
