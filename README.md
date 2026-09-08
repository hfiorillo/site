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
| `view/` | templ templates |
| `handler/` | Go HTTP handlers |
| `internal/markdown/` | Markdown parser |
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

Add an entry to `content/routes/routes.yml`. Route names link directly to Komoot;
existing detail URLs redirect there too.

```yaml
- name: Route Name
  slug: route-slug
  location: Start to End, Country
  date: 2026-05-01
  route_url: https://www.komoot.com/tour/YOUR_TOUR_ID
  distance_km: 126
  elevation_gain: 3519
  packlist: https://lighterpack.com/r/YOUR_LIST
```

## Building for production

```sh
docker build -t site .
```

Or push to main — GitHub Actions builds and deploys to Cloud Run.


## Photo gallery

The Pictures page includes JPEG and PNG files in `public/images/posts/{post-filename}/`
for published posts. Images are grouped by the post's year, not inferred EXIF dates.
Missing historical images are not listed, and byte-identical photos are deduplicated.

Install the WebP encoder once (`brew install webp` on macOS), then run:

```sh
task gallery
```

`task images` also runs this after HEIC conversion. Run it again after publishing a
post or changing its photos. Commit `content/pictures/gallery.json` and the generated
`public/gallery/` files with the post, then rebuild/redeploy. No encoder runs in Cloud
Run. Existing previews are reused on subsequent runs.

The gallery uses one 720px WebP preview per photo, lazy loading and fixed image
dimensions. Original files load when a visitor opens a photo. Preview filenames
include a content/settings hash and receive a one-year immutable cache header.
`task gallery` removes obsolete generated previews after updating the manifest,
so there is only one preview per current photo. Original photos are never modified.

Posts and homepage previews use their original images directly; no article image
variants or separate article manifest are generated.

`task run` watches content, source CSS, YAML/JSON and image changes. Generated
`public/styles.css` is excluded to prevent rebuild loops. Content changes restart
the server and refresh its parsed-post and manifest caches.
