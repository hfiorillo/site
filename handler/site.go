package handler

import (
	"log/slog"
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
)

func loadSiteMeta() {
	raw, err := os.ReadFile("./content/site.yml")
	if err != nil {
		slog.Error("reading site.yml", "error", err)
		return
	}
	if err := yaml.Unmarshal(raw, &siteMeta); err != nil {
		slog.Error("parsing site.yml", "error", err)
		return
	}
}

func siteImage() string {
	siteOnce.Do(loadSiteMeta)
	if siteMeta.Image != "" {
		return siteMeta.Image
	}
	return "/public/images/harryfiorilloxyz-removebg-preview.png"
}
