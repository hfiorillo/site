# site

Personal blog at [blog.fiorillo.xyz](https://blog.fiorillo.xyz). Go + templ + Tailwind v4 + daisyUI.

## Quick start

```sh
task run       # hot reload (air + templ watch + tailwind watch)
```

Or manually:

```sh
templ generate && npx @tailwindcss/cli -i view/css/app.css -o public/styles.css && go run ./main.go
```

## Structure

| Path | What |
|------|------|
| `content/site.yml` | Site metadata (title, description, OG image) |
| `content/posts/{year}/` | Blog posts (markdown with YAML front matter) |
| `content/routes/routes.yml` | Route metadata |
| `content/aboutme/` | About page content |
| `public/images/` | Blog images, avatar |
| `public/routes/` | GPX files |
| `view/` | templ templates |
| `handler/` | Go HTTP handlers |
| `internal/markdown/` | Markdown parser |
| `internal/gpx/` | GPX parser |
| `scripts/newpost.sh` | Creates a new blog post template |

## Adding a blog post

```sh
task new-post
```

Or manually create `content/posts/{year}/{slug}.md`:

```markdown
---
title: My Post
date: 2026-05-01
tags:
- tag1
published: true
description: Short description.
---

Content here...
```

Images go in `public/images/posts/{slug}/` and are referenced as `/public/images/posts/{slug}/photo.jpg`.

## Adding a route

1. Drop the GPX file in `public/routes/`
2. Add an entry to `content/routes/routes.yml`

```yaml
- name: Route Name
  slug: route-slug
  location: Start to End, Country
  date: 2026-05-01
  gpx: /public/routes/file.gpx
```

## Building for production

```sh
docker build -t site .
```

Or push to main — GitHub Actions builds and deploys to Cloud Run.
