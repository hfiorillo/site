package handler

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type SiteMeta struct {
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Image       string       `yaml:"image"`
	Blog        SectionMeta  `yaml:"blog"`
	Feed        SectionMeta  `yaml:"feed"`
	Routes      SectionMeta  `yaml:"routes"`
}

type SectionMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

var (
	siteOnce sync.Once
	siteMeta SiteMeta
	siteErr  error
)

func loadSiteMeta() {
	raw, err := os.ReadFile("./content/site.yml")
	if err != nil {
		siteErr = fmt.Errorf("reading site.yml: %w", err)
		return
	}
	if err := yaml.Unmarshal(raw, &siteMeta); err != nil {
		siteErr = fmt.Errorf("parsing site.yml: %w", err)
		return
	}
}

func siteImage() string {
	siteOnce.Do(loadSiteMeta)
	if siteMeta.Image != "" {
		return siteMeta.Image
	}
	return "/public/images/avatar.jpg"
}
