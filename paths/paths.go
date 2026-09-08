// Package paths centralises the website route and asset path constants so
// paths are defined once and reused across routing, handlers, feeds and templates.
package paths

const (
	// Routes registered on the chi router.
	Root          = "/"
	Blog          = "/blog"
	BlogPost      = Blog + "/{filename}"
	AboutMe       = "/aboutme"
	AboutThisSite = "/about-this-site"
	Pictures      = "/pictures"
	Work          = "/work"
	Routes        = "/routes"
	RouteDetail   = Routes + "/{slug}"
	Feed          = "/feed.xml"
	Sitemap       = "/sitemap.xml"

	// Markdown content base names passed to markdown.LoadMarkdownPost.
	AboutMeMarkdown       = AboutMe + "/aboutme"
	AboutThisSiteMarkdown = AboutThisSite + "/about-this-site"
	PicturesMarkdown      = Pictures + "/pictures"
	WorkMarkdown          = Work + "/work"
	PostsMarkdownPrefix   = "/posts/"

	// URL query parameters.
	BlogQueryTag = "tag"
	BlogQueryCat = "category"

	// Static asset paths (served from the embedded /public tree).
	Styles  = "/public/styles.css"
	Favicon = "/public/favicon.svg"
	Avatar  = "/public/images/harryfiorilloxyz-removebg-preview.png"

	// Content file paths (read from disk at runtime).
	SiteYAML   = "./content/site.yml"
	RoutesYAML = "./content/routes/routes.yml"
)

// StylesURL is set from the embedded stylesheet before the HTTP server starts.
// Its content hash keeps browsers and CDNs from reusing CSS from an older build.
var StylesURL = Styles

const (
	// GalleryManifest contains the precomputed, published-post gallery.
	GalleryManifest = "./content/pictures/gallery.json"
	// GalleryAssets is reserved for content-hashed, immutable WebP previews.
	GalleryAssets = "/public/gallery/"
	// PostImages and PostImagesDirectory identify local post photo folders.
	PostImages          = "/public/images/posts/"
	PostImagesDirectory = "./public/images/posts"
)
