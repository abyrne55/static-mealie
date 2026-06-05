# static-mealie

A CLI tool that generates a ready-to-deploy static HTML website full of recipes from your [Mealie](https://mealie.io) instance. Great for sharing your recipes with the world without directly exposing Mealie itself to the internet. Inspired by [iamkirkbater/mealie-markdown-exporter](https://github.com/iamkirkbater/mealie-markdown-exporter).

## Features

This simple Golang tool fetches all recipes (including images and notes) from your Mealie instance via its API, renders each one into multiple formats, generates a listing page and XML sitemap, and writes everything to a directory ready for any static web server.

- Produces completely static websites composed of…
  - …HTML+CSS for your human friends
  - …XML for search engines
  - …Markdown and JSON-LD for parsers (e.g., Mealie's import-recipe-from-URL tool), AI agents, and everyone else
  - …not a single line of Javascript or CGI/server-side scripts
- Generated sites are responsive/mobile-friendly, lightweight, and support dark mode
- Easy to run as a cron job or one-shot quadlet/systemd service
- Written in pure Golang using only stdlib (no dependencies/mods)
- Containerized atop a [Project Hummingbird](https://hummingbird-project.io/) static base image and rebuilt/pushed to GHCR daily (read: pretty darn secure / unlikely to contain CVEs)

## Getting Started

### Podman/Docker Quick Start (recommended)

```sh
# Log in to demo.mealie.io and store the JWT as a Podman secret
curl -sf -X POST 'https://demo.mealie.io/api/auth/token' \
  -d 'username=changeme@example.com&password=MyPassword' \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4 \
  | podman secret create static-mealie-token -

# Generate the site into a Podman volume
podman volume create static-mealie-output
podman run --rmi \
  -v static-mealie-output:/output \
  --secret static-mealie-token \
  -e SM_MEALIE_URL=https://demo.mealie.io \
  ghcr.io/abyrne55/static-mealie:main

# Serve it
podman run -d --rmi --name recipes-httpd \
  -v static-mealie-output:/usr/local/apache2/htdocs/:ro \
  --publish 8181:8080 \
  quay.io/hummingbird/httpd:latest

# Open it in your browser (on macOS, replace "xdg-open" with just "open")
xdg-open "http://127.0.0.1:8181/" 

# Clean up
read -p "Press Enter to stop demo and clean up" \
  && podman rm -f recipes-httpd \
  && podman volume rm static-mealie-output \
  && podman secret rm static-mealie-token
```

> [!TIP]
> If zero recipes appear, check whether [demo.mealie.io](https://demo.mealie.io) currently has any recipes — the demo site is regularly wiped and may be empty. If it is, [import](https://demo.mealie.io/g/home/r/create/url) a recipe or two ([cacio e pepe](https://cooking.nytimes.com/recipes/1020729-vegan-cacio-e-pepe), perhaps?) and try the commands above again.

### Build from source

Install directly into your `$GOBIN`:

```sh
go install github.com/abyrne55/static-mealie@latest
static-mealie --help
```

Or manually clone and build from source:

```sh
git clone https://github.com/abyrne55/static-mealie.git
cd static-mealie
go build
./static-mealie --help
```

## Demo Site

A live demo is deployed automatically via GitHub Pages on every push to `main`. It uses a built-in mock Mealie server with sample recipes — no real Mealie instance needed.

## Mock Server

The `cmd/mockmealie/` directory contains a standalone mock Mealie API server with embedded sample recipes and images. Useful for development and testing without a live Mealie instance:

```sh
go run ./cmd/mockmealie                # starts on :9925
go run . --mealie-url http://localhost:9925 --mealie-token mock --out-dir output -v
```

The mock server accepts any Bearer token and serves 6 sample recipes covering various features (ingredient sections, fractional quantities, inline markdown, notes, and missing images).

## Reference

### Flags

```text
--mealie-url string      Mealie base URL (env: SM_MEALIE_URL)
--mealie-token string    API token or file:///path (env: SM_MEALIE_TOKEN)
--out-dir string         Output directory (default: "output", env: SM_OUT_DIR)
--out-title string       Site title (default: "Recipes", env: SM_OUT_TITLE)
--out-base-url string    Base URL for sitemap/links; output not standards-compliant
                         unless set to an absolute URL like https://example.com
                         (default: "/", env: SM_OUT_BASE_URL)
--clean-slate            Wipe output directory before building (env: SM_CLEAN_SLATE)
--no-clobber             Skip files that already exist (env: SM_NO_CLOBBER)
-v                       Verbose logging
```

`--clean-slate` and `--no-clobber` are mutually exclusive.

### Token Resolution

The API token is resolved from the first available source:

1. `--mealie-token` flag (supports `file:///path/to/token`)
2. `SM_MEALIE_TOKEN` env var (supports `file:///path/to/token`)
3. `/run/credentials/static-mealie-token` file
4. `/run/secrets/static-mealie-token` file

The credential file paths match the default mount points for systemd/quadlet [credentials](https://systemd.io/CREDENTIALS/) (`LoadCredential=`/`$CREDENTIALS_DIRECTORY`) and Podman/K8s [secrets](https://docs.podman.io/en/latest/markdown/podman-secret-create.1.html).

### Output Structure

```text
output/
  index.html
  sitemap.xml
  {recipe-slug}/
    index.html
    recipe.md
    image.webp      (if recipe has an image)
```

### Running Tests

```sh
go test ./...
```
