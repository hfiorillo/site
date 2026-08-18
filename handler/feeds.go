package handler

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/hfiorillo/site/internal/markdown"
)

type rssFeed struct {
	XMLName xml.Name  `xml:"rss"`
	Version string    `xml:"version,attr"`
	Atom    string    `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	AtomLink    rssAtom   `xml:"atom:link"`
	Language    string    `xml:"language"`
	Items       []rssItem `xml:"item"`
}

type rssAtom struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (p PageHandler) HandleFeed(w http.ResponseWriter, r *http.Request) error {
	siteOnce.Do(loadSiteMeta)
	posts, err := markdown.LoadMarkdownPosts(r.Context())
	if err != nil {
		return err
	}

	items := make([]rssItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, rssItem{
			Title:       post.Title,
			Link:        p.SiteURL + "/blog/" + post.Filename,
			GUID:        p.SiteURL + "/blog/" + post.Filename,
			Description: post.Description,
			PubDate:     post.Date.Format(time.RFC822),
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       siteMeta.Feed.Title,
			Link:        p.SiteURL,
			Description: siteMeta.Feed.Description,
			Language:    "en",
			AtomLink: rssAtom{
				Href: p.SiteURL + "/feed.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	return xml.NewEncoder(w).Encode(feed)
}

func (p PageHandler) HandleSitemap(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts(r.Context())
	if err != nil {
		posts = nil
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	dateFormat := "2006-01-02"

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	addURL := func(loc, lastmod, changefreq, priority string) {
		buf.WriteString("<url>\n")
		buf.WriteString(fmt.Sprintf("<loc>%s</loc>\n", loc))
		buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", lastmod))
		buf.WriteString(fmt.Sprintf("<changefreq>%s</changefreq>\n", changefreq))
		buf.WriteString(fmt.Sprintf("<priority>%s</priority>\n", priority))
		buf.WriteString("</url>\n")
	}

	now := time.Now().Format(dateFormat)
	addURL(p.SiteURL+"/", now, "monthly", "1.0")
	addURL(p.SiteURL+"/blog", now, "weekly", "0.8")
	addURL(p.SiteURL+"/pictures", now, "monthly", "0.6")
	addURL(p.SiteURL+"/work", now, "monthly", "0.6")
	addURL(p.SiteURL+"/aboutme", now, "monthly", "0.6")
	addURL(p.SiteURL+"/about-this-site", now, "monthly", "0.4")

	for _, post := range posts {
		addURL(p.SiteURL+"/blog/"+post.Filename, post.Date.Format(dateFormat), "never", "0.6")
	}

	routesOnce.Do(loadRoutes)
	if routesErr == nil {
		addURL(p.SiteURL+"/routes", now, "monthly", "0.5")
		for _, entry := range routesList {
			addURL(p.SiteURL+"/routes/"+entry.Slug, now, "never", "0.5")
		}
	}

	buf.WriteString("</urlset>\n")
	w.Write(buf.Bytes())
	return nil
}