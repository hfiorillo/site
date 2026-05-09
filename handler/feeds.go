package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/hfiorillo/site/internal/markdown"
)

func (p PageHandler) HandleFeed(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">` + "\n")
	buf.WriteString("<channel>\n")
	buf.WriteString("<title>Harry Fiorillo-Hughes | Blog</title>\n")
	buf.WriteString(fmt.Sprintf("<link>%s</link>\n", p.SiteURL))
	buf.WriteString("<description>Personal blog about DevOps, engineering, and adventures</description>\n")
	buf.WriteString(fmt.Sprintf("<atom:link href=\"%s/feed.xml\" rel=\"self\" type=\"application/rss+xml\"/>\n", p.SiteURL))
	buf.WriteString("<language>en</language>\n")

	for _, post := range posts {
		buf.WriteString("<item>\n")
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", xmlEscape(post.Title)))
		buf.WriteString(fmt.Sprintf("<link>%s/blog/%s</link>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<guid>%s/blog/%s</guid>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<description>%s</description>\n", xmlEscape(post.Description)))
		buf.WriteString(fmt.Sprintf("<pubDate>%s</pubDate>\n", post.Date.Format(time.RFC822)))
		buf.WriteString("</item>\n")
	}

	buf.WriteString("</channel>\n")
	buf.WriteString("</rss>\n")

	w.Write(buf.Bytes())
	return nil
}

func (p PageHandler) HandleSitemap(w http.ResponseWriter, r *http.Request) error {
	posts, err := markdown.LoadMarkdownPosts()
	if err != nil {
		posts = nil
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	dateFormat := "2006-01-02"

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>monthly</changefreq>\n")
	buf.WriteString("<priority>1.0</priority>\n")
	buf.WriteString("</url>\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/blog</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>weekly</changefreq>\n")
	buf.WriteString("<priority>0.8</priority>\n")
	buf.WriteString("</url>\n")

	buf.WriteString("<url>\n")
	buf.WriteString(fmt.Sprintf("<loc>%s/aboutme</loc>\n", p.SiteURL))
	buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
	buf.WriteString("<changefreq>monthly</changefreq>\n")
	buf.WriteString("<priority>0.6</priority>\n")
	buf.WriteString("</url>\n")

	for _, post := range posts {
		buf.WriteString("<url>\n")
		buf.WriteString(fmt.Sprintf("<loc>%s/blog/%s</loc>\n", p.SiteURL, post.Filename))
		buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", post.Date.Format(dateFormat)))
		buf.WriteString("<changefreq>never</changefreq>\n")
		buf.WriteString("<priority>0.6</priority>\n")
		buf.WriteString("</url>\n")
	}

	routesOnce.Do(loadRoutes)
	if routesErr == nil {
		buf.WriteString("<url>\n")
		buf.WriteString(fmt.Sprintf("<loc>%s/routes</loc>\n", p.SiteURL))
		buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
		buf.WriteString("<changefreq>monthly</changefreq>\n")
		buf.WriteString("<priority>0.5</priority>\n")
		buf.WriteString("</url>\n")
		for _, entry := range routesList {
			buf.WriteString("<url>\n")
			buf.WriteString(fmt.Sprintf("<loc>%s/routes/%s</loc>\n", p.SiteURL, entry.Slug))
			buf.WriteString(fmt.Sprintf("<lastmod>%s</lastmod>\n", time.Now().Format(dateFormat)))
			buf.WriteString("<changefreq>never</changefreq>\n")
			buf.WriteString("<priority>0.5</priority>\n")
			buf.WriteString("</url>\n")
		}
	}

	buf.WriteString("</urlset>\n")

	w.Write(buf.Bytes())
	return nil
}
