# AGENTS.md

## Project Overview

Personal blog/portfolio site deployed at blog.fiorillo.xyz (domain: fiorillo.xyz). Built as a Go web server that compiles into a static Docker image + self-hosted on Google Cloud Run. The site is content-driven: nearly all text, metadata, routes, and post content lives in Markdown/YAML files under `content/`, not in Go code.

## Stack

- **Language / server**: Go (`chi` router), module `github.com/hfiorillo/site`
- **Templating**: `a-h/templ` (`.templ` files compile to Go via `templ generate`)
- **Markdown**: `yuin/goldmark` (with `goldmark-meta` for front matter, `html.WithUnsafe()` for raw HTML / embeds)
- **CSS**: Tailwind CSS v4 + daisyUI v5, compiled via `@tailwindcss/cli`
- **Interactivity**: Alpine.js only (self-hosted — no CDN). No htmx, no other JS frameworks.
- **Fonts**: Geist (self-hosted)
- **Deploy**: multi-stage Dockerfile → GitHub Actions (`google.yml`, on push to main) → Cloud Run

## Directory Structure (big picture)

- `main.go` — entrypoint, chi router, middleware, graceful shutdown, `//go:embed public`
- `handler/` — HTTP handlers, split by domain:
  - `handler.go` — PageHandler struct, `Make()` error wrapper, index/about/handle helpers
  - `blog.go` — blog listing + individual post pages
  - `routes.go` — route listing/detail/coords
  - `feeds.go` — RSS feed + sitemap
  - `site.go` — `content/site.yml` site metadata loader
- `internal/` — non-HTTP libraries:
  - `internal/markdown/` — goldmark setup, front matter parsing, post loading, header/TOC parsing
  - `internal/gpx/` — GPX parsing (Haversine distance, elevation gain)
- `models/models.go` — shared structs (BlogPost, Route, PageMeta, Headers, etc.)
- `paths/paths.go` — single source of truth for website route & static-asset path constants (used by main.go, handlers, feeds, and templates). Content-file and asset paths live here too.
- `view/` — all templ templates:
  - `layout/` — `app.templ` (Base layout / three-column shell), `brand.templ` (left sidebar), `sidenav.templ` (right sidebar)
  - `pages/` — one template file per page type
  - `components/icons/` — SVG icons (GitHub, Instagram, Komoot, Strava)
- `content/` — ALL content & metadata (Markdown + YAML), not code
- `public/` — static assets (images, GPX, CSS output, favicon, robots.txt), embedded at compile time
- `scripts/` — `newpost.sh`, `images.sh`
- `Taskfile.yml` — task runner (e.g. `task new-post`, `task images`)
- `.github/workflows/google.yml`, `Dockerfile` — CI/CD

## The Content Model (most important mental model)

Everything a user would want to change lives in `content/`, NOT in Go or templates:

- `content/site.yml` — global site metadata (title, description, OG image, per-section titles/descriptions). Handlers read this via `loadSiteMeta()`.
- `content/posts/{year}/*.md` — blog posts, one file per post, organized into year subdirectories. Front matter includes `title`, `date`, `categories`, `tags`, `published`, `description`, and optional `preview-image` (shown on the blog listing).
- `content/routes/routes.yml` — route list (name, slug, location, date, gpx path, distance_km, elevation_gain, packlist). Distance/elevation come from YAML; GPX is only parsed for the map.
- `content/aboutme/`, `content/about-this-site/` — standalone markdown pages loaded the same way as posts.
- `content/projects/` — projects content.

Consequence: when asked to change site text, listing behavior, or add content, prefer editing these YAML/Markdown files. Only touch Go when the structure/behavior itself needs changing.

## Key Recurring Patterns & Gotchas

- **Posts are loaded twice, always keep them in sync**: `LoadMarkdownPosts` (blog listing, returns all) vs `LoadMarkdownPost` (single post via an O(1) filename→path map built at init). Both are used by handlers.
- **`published: false`** hides a post from listings, RSS, and sitemap; direct URL shows the 404 page. Set to `true` to publish.
- **Front matter must open AND close with `---`** — goldmark-meta requires a leading delimiter (a common past breakage).
- **`//go:embed public` bakes assets at compile time**: newly added images/static assets require a rebuild+redeploy before they appear. Missing/casing-mismatched files produce 404s.
- **Serving static assets**: images are committed to the repo under `public/images/...` and referenced as `/public/...`. HEIC originals are gitignored; `task images` converts HEIC→JPEG (and deletes the HEIC).
- **Tailwind v4 does NOT scan `.templ` files by default.** Responsive utility classes used in templates are maintained as explicit `@media` rules in `view/css/app.css`. Don't rely on arbitrary new responsive classes from templates.
- **Raw HTML/embeds**: `html.WithUnsafe()` allows `<blank>` in markdown. Used intentionally for embeds (it is a personal blog).
- **Jasan animations/branding**: left sidebar has animated color dots + avatar; right sidebar has nav + socials + theme toggle. `templ generate` output (`*_templ.go`) is gitignored.

## Build & Run

Local dev / verify changes:

```sh
templ generate && npx @tailwindcss/cli -i view/css/app.css -o public/styles.css && go run ./main.go
```

Then visit `http://localhost:3001` (default `HTTP_LISTEN_ADDR=:3001`). Hot reload via `air` if preferred.

Common tasks:
- `task new-post` — scaffold a blank post in the current year's folder
- `task images` — batch-convert HEIC images to JPEG across image dirs
- Full production build: compile CSS → `templ generate` → `go build`

## CI/CD & Deploy

- Push to `main` triggers GitHub Actions → builds the Docker image (CSS stage `@tailwindcss/cli`, Go stage with templ) → pushes to GAR → deploys to Cloud Run.
- .dockerignore keeps the build context minimal; CI uses `docker/build-push-action`, image tagged with SHA + `latest`.
- Always run `go vet ./...` and a clean build before considering a change complete.

## Conventions & Coding Notes

- Page metadata flows through the `PageMeta` struct (title, description, OG image, canonical URL, structured data). Every handler builds one and passes it to the `Base()` layout.
- Site text/headings should come from content files (front matter / site.yml). Avoid hardcoding strings in Go unless it's an inherently generated listing page.
- Handlers are split by domain (blog/routes/feeds/site) — keep it that way; don't grow `handler.go` back into a monolith.
- **Always use `paths` constants** for any website route, static-asset URL, or content-file path instead of hardcoding strings (e.g. `paths.Blog`, `paths.Avatar`, `paths.RoutesYAML`). Never hardcode a new `/blog`, `/routes`, `/public/...`, or `content/...` literal — add a constant in `paths/paths.go` and reference it, including inside `.templ` files.
- Favicon currently the fish SVG (avatar PNG doesn't crop well as a favicon).
