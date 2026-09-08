---
title: How this website works (and what it costs to run)
date: 2026-09-08
author: Harry Fiorillo
categories:
- software-engineering
tags:
- work
- golang
- templ
- cloud-run
- devops
section: work
published: false
description: A look under the bonnet of this blog. Go, templ, Markdown, a few useful commands and a GitHub workflow that does the deploying for me.
---

I've spent a fair amount of time fiddling with this website. Probably more time than I've spent writing things for it, which slightly defeats the point.

But that's part of the fun.

I wanted somewhere to put bike rides, trips, photos and the odd software engineering post. Something that felt like the weblogs I grew up reading, with enough freedom to make it look how I wanted. [Kottke](https://kottke.org/) has been a big influence. The writing, the links, the strange colourful things down the side. All good stuff.

So, here's how it actually works. Including how I develop it locally, how it gets onto the internet and the incredibly exciting running cost of £0.01.

## The ingredients

The source lives on [GitHub](https://github.com/hfiorillo/site). Most of the moving parts are fairly small:

- **[Go](https://go.dev/)** runs the web server. It receives a request, loads the relevant content and sends back HTML.
- **[chi](https://github.com/go-chi/chi)** handles routing. It decides whether you're asking for a blog post, the Work page, a route or something else.
- **[templ](https://templ.guide/)** is how I write the page templates. HTML with Go expressions, compiled into Go code.
- **[Goldmark](https://github.com/yuin/goldmark)** turns the Markdown posts into HTML. It also handles the metadata at the top of each post through an extension.
- **[Tailwind CSS](https://tailwindcss.com/)** and **[daisyUI](https://daisyui.com/)** handle styling, alongside some ordinary CSS for the layout and colourful circles.
- **[Alpine.js](https://alpinejs.dev/)** handles the little interactive bits: the theme switch, photo lightbox and reading progress bar.
- **Docker, GitHub Actions and Google Cloud Run** get it built and hosted.

There isn't a database. The posts are files. The photos are files. Most of the things I want to change are files.

Very sophisticated.

## From a Markdown file to a page

A post starts as a Markdown file in `content/posts/`, organised by year. At the top is a small block of YAML called front matter. It tells the site what the post is, when it was written and where it should appear.

For example, the start of a Work post looks something like this:

```yaml
---
title: Something I built
date: 2026-09-08
categories:
- software-engineering
tags:
- work
- golang
section: work
published: false
description: A short explanation of what I built and why.
---
```

The `date` is the publication date. The filename becomes the final part of the post's URL, so renaming it also changes the link.

`section: work` adds it to the Work listing. The `work` tag makes it discoverable through the tag links as well. It still appears in the main blog alongside everything else once it's published.

Setting `published: false` keeps a draft out of the listings, RSS feed and sitemap. Visiting its URL returns a 404 too. When it's ready, I change that to `true`.

When you open a post, the handler loads the file and Goldmark converts it into HTML. That gets passed to a templ component, which wraps it in the shared layout: navigation, colours, metadata and all the bits around the edges.

The blog listing keeps a short-lived cache of the loaded posts for 60 seconds. Individual posts are read from disk when requested. There isn't a database query happening behind every paragraph.

## What templ is doing

I quite like being able to write a component and use normal Go values inside it. A small example:

```templ
package components

templ PostHeading(title string) {
    <h1 class="text-3xl font-bold">{ title }</h1>
}
```

Running `templ generate` turns this into a Go function that renders HTML. In the actual site, the handlers pass in the post and page metadata, and the templates decide how they should look.

The generated `*_templ.go` files aren't committed. They're rebuilt locally and in Docker. The [templ generation docs](https://templ.guide/core-concepts/template-generation/) explain that part in more detail.

Go sends the finished HTML to the browser. Alpine then adds the small interactions on top. The current version uses Alpine; some of my earlier experiments used htmx.

## Working on it locally

I use [Task](https://taskfile.dev/) for the commands I don't want to keep typing, and [Air](https://github.com/air-verse/air) to rebuild the Go app while I'm working.

With Go, Node/npm and Task installed, the main development tools can be installed like this:

```sh
npm ci
go mod download
go install github.com/a-h/templ/cmd/templ@v0.3.1020
go install github.com/air-verse/air@latest
```

The templ version above matches the Go dependency in the repo at the time of writing. Its executable, and Air's, need to be on your `PATH`.

Then:

```sh
task run
```

That starts Air. The build command in `.air.toml` does three things:

```sh
templ generate
npx @tailwindcss/cli -i view/css/app.css -o public/styles.css
go build -o ./.build/main main.go
```

Air then runs the resulting binary. The local site is at `http://localhost:3001` unless I've changed `HTTP_LISTEN_ADDR`.

My usual process is to change one thing, look at it in the browser, decide I've made it worse and change it again. Repeat until the circles look right.

One detail worth knowing: the current Air configuration watches Go and template-related files. It doesn't watch Markdown, CSS source or new photos. Editing an existing post can show up on a refresh because its content is read from disk, but the listing can lag behind because of that 60-second cache. For CSS changes, new posts or new photos, I restart `task run` to get a fresh build.

## Adding a post and its photos

There is a shortcut for a blank post:

```sh
task new-post
```

That creates `new-post.md` in the current year's folder. The script also accepts a title when called directly:

```sh
sh scripts/newpost.sh "A weekend on the bike"
```

It turns the title into a filename and adds the front matter. The current script sets `published: true`, so I change that to `false` while writing. It can overwrite a file with the same name too, so I give each new post its own title.

Photos go into a folder for that trip under `public/images/posts/`. If they're HEIC files from my phone, I run:

```sh
task images
```

This uses macOS's `sips` tool to convert them to JPEG, fitting the longest edge within 1,600 pixels. It then removes the HEIC originals, so I keep my original photos elsewhere. It doesn't currently recompress every existing JPEG or generate different sizes for phones and desktops.

The HTML in the posts can include photo layouts and Komoot embeds. Goldmark is configured to allow that raw HTML, which is useful for a personal blog where I'm writing the content myself.

Static assets are embedded into the Go binary using `go:embed`. This is convenient, but it also means adding a photo to the folder doesn't magically put it inside a running binary. It needs a rebuild locally and a deployment for the live site.

A lesson learnt more than once.

## Getting it onto the internet

Once I'm happy with a change, I check it locally:

```sh
templ generate
npx @tailwindcss/cli -i view/css/app.css -o public/styles.css
go test ./...
go vet ./...
mkdir -p .build
go build -o ./.build/site ./main.go
```

After reviewing the changes, pushing to `main` triggers the workflow in `.github/workflows/google.yml`.

The sequence is:

```text
Push to main
    ↓
GitHub Actions builds the Docker image
    ↓
Image goes to Google Artifact Registry
    ↓
Cloud Run deploys that image
    ↓
The updated site is available
```

The Dockerfile has three stages. A Node stage installs the frontend dependencies and builds the CSS. A Go stage generates the templates and compiles the server. The final stage copies the server, content and public assets into a minimal distroless image, running as a non-root user.

Node and the compilers stay in the build stages. The running container just needs to serve the site.

The workflow pushes two image tags: the Git commit SHA and `latest`. The deployment uses the SHA, so it points at the particular version that was just built.

## What needs setting up in Google Cloud and GitHub?

The workflow does the recurring build and deployment. It doesn't create all the infrastructure from nothing.

To reproduce the setup, I'd start with:

1. A Google Cloud project with billing enabled and the Cloud Run and Artifact Registry APIs enabled.
2. A Docker repository in Artifact Registry called `personal`, matching the image path in this workflow.
3. A deployment service account with permission to push images and deploy Cloud Run services. That means Artifact Registry Writer on the repository, Cloud Run deployment permissions, and Service Account User on the runtime service account. Google's [deployment action documentation](https://github.com/google-github-actions/deploy-cloudrun#authorization) describes the required roles.
4. A Cloud Run service configured for public access, with its container port matching the app's listening address.
5. A GitHub environment named `google`, holding the configuration used by the workflow.

The GitHub environment variables are:

```text
PROJECT_ID     Google Cloud project ID
GAR_LOCATION   Artifact Registry location, e.g. europe-west2
SERVICE        Cloud Run service and image name
REGION         Cloud Run region, e.g. europe-west2
```

The existing workflow also reads a secret called `GCP_CREDENTIALS`. This contains a service-account JSON key, used by the Google authentication action and Docker login step. It belongs in GitHub's encrypted secrets, not in the repository.

For a fresh setup, I'd use [Workload Identity Federation](https://github.com/google-github-actions/auth#workload-identity-federation) instead, so GitHub can authenticate without storing a long-lived key. That requires changing the current authentication and registry-login steps; setting `id-token: write` alone doesn't make it happen.

There's also a small port gotcha. This app reads `HTTP_LISTEN_ADDR`, not Cloud Run's `PORT` variable. One valid setup is a Cloud Run container port of `8080` with `HTTP_LISTEN_ADDR=:8080`. Set `SITE_URL` to the public HTTPS address as well, so canonical links and RSS URLs point to the right place. The workflow doesn't currently specify those service settings, so they need to be configured separately.

Finally, connect the domain using a [supported Cloud Run domain setup](https://cloud.google.com/run/docs/mapping-custom-domains). Cloudflare sits in front of my site and caches static assets. The domain, DNS and HTTPS setup are separate from the GitHub deployment workflow.

## The time the styling disappeared

We recently deployed a new layout and the live site looked broken. The new HTML had arrived, but Cloudflare was still handing out an older cached stylesheet.

Locally? Fine. On the internet? A mess.

The fix was to put a hash of the CSS into its URL. When the stylesheet changes, the URL changes too, so browsers and Cloudflare fetch the new version. If it hasn't changed, they can keep using the cached copy.

We also made Tailwind scan the templates and Markdown explicitly. Docker builds the CSS before generating the Go templates, so relying on files that only happen to exist on my laptop is a bad idea.

This is the sort of thing I like about building the site myself. Occasionally annoying. Usually a useful lesson.

## The running cost

At the moment, my running cost is **£0.01**. One penny.

Here is a chart, because apparently one penny needed a chart.

<figure style="margin:2rem 0; padding:1.25rem; border-top:1px dotted currentColor; border-bottom:1px dotted currentColor;">
  <div role="img" aria-label="Website running cost at the time of writing: one penny in total. Individual cloud services have not been itemised." style="font-family:Arial,sans-serif;">
    <div style="display:flex; justify-content:space-between; gap:1rem; margin-bottom:0.75rem;"><span>Website running cost</span><strong>£0.01</strong></div>
    <div style="height:24px; background:var(--color-primary, #226a43);"></div>
    <div style="display:flex; justify-content:space-between; margin-top:0.5rem; font-size:14px;"><span>£0.00</span><span>£0.01</span></div>
  </div>
  <figcaption style="font-size:14px; margin-top:1rem;">Running cost at the time of writing, not a per-visitor price or a guaranteed monthly bill. Cloud services are grouped because I haven't itemised that penny here.</figcaption>
</figure>

In terms of what makes up the setup: Go, templ, Tailwind, daisyUI and Alpine don't add a software licence fee. GitHub runs the build, Artifact Registry stores the images, Cloud Run runs the server and Cloudflare handles the cache. The penny above is the running-cost total I'm reporting, rather than separate figures for each of those services. Domain registration and renewal are a separate expense.

That doesn't mean unlimited traffic is free. [Cloud Run charges for usage](https://cloud.google.com/run/pricing), including compute, requests and outbound data, and image-heavy posts can move quite a lot of data. Caching helps, but the bill depends on what actually reaches Google and which free allowances are available.

For the amount of traffic this site currently gets, I'm happy with that. If it gets busier, smaller photos, responsive image sizes and better image caching are the next things to work on.

Until then, I'll probably keep moving the circles around.
