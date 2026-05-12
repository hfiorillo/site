package models

import (
	"html/template"
	"time"
)

type BlogPost struct {
	Filename        string
	Title           string
	Slug            string
	Description     string
	Date            time.Time
	Content         template.HTML
	Headers         Headers
	Metadata        Metadata
	ReadTimeMinutes int
}

type Metadata struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags"`
	Published   bool     `yaml:"published"`
	Description string   `yaml:"description"`
	Image       string   `yaml:"image"`
}

type PageMeta struct {
	Title       string
	Description string
	Image       string
	URL         string
}

type TopLevelHeader struct {
	Headers  string
	MidLevel []MidLevelHeader
}

type MidLevelHeader struct {
	Headers     string
	BottomLevel []BottomLevelHeader
}

type BottomLevelHeader struct {
	Headers string
}

type Headers struct {
	TopLevel []TopLevelHeader
}

type Route struct {
	Slug          string
	Name          string
	Location      string
	DistanceKm    float64
	ElevationGain float64
	ElevationMax  float64
	ElevationMin  float64
	Date          time.Time
	CoordsJSON    string
	GPXFile       string
	Packlist      string
}
