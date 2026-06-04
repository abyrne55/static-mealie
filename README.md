# static-mealie

A CLI tool that exports recipes from a [Mealie](https://mealie.io) instance as a complete static HTML website — no Hugo, no JavaScript, no external dependencies.

## Features

- Generates a static site with [Simple.css](https://simplecss.org/) styling and dark mode support
- Structured ingredient data with Unicode vulgar fractions (½, ¼, ⅓, etc.)
- JSON-LD schema.org/Recipe embedded in each recipe page
- Markdown versions of each recipe with checkbox ingredients
- XML sitemap for search engines
- Recipe images downloaded and served alongside HTML
- Zero external Go dependencies (stdlib only)

## Installation

Build from source:

```sh
go build -o static-mealie .
```

## Usage

```sh
static-mealie --url https://your-mealie-instance --token your-token
```

### Flags

```
--url string       Mealie base URL (env: MEALIE_URL)
--token string     API token or file:///path (env: MEALIE_TOKEN)
--out string       Output directory (default: "public")
--title string     Site title (default: "Recipes")
--site-url string  Base URL for sitemap/links (default: "/")
-v                 Verbose logging
```

### Environment Variables

Flags take precedence over environment variables.

- `MEALIE_URL` — Mealie base URL
- `MEALIE_TOKEN` — API token (supports `file:///path/to/token`)

### API Token from File

To avoid exposing your API token in shell history:

```sh
static-mealie --url https://your-mealie-instance --token file:///path/to/token
```

## Container

Images are published to `ghcr.io/abyrne55/static-mealie` for `linux/amd64` and `linux/arm64`.

```sh
# Store your API token as a Podman secret
podman secret create mealie-api-key /path/to/token

# Generate the site into ./output
podman run --rm \
  -v ./output:/output:U,Z \
  --secret mealie-api-key \
  ghcr.io/abyrne55/static-mealie:main \
  --url https://your-mealie-instance \
  --token file:///run/secrets/mealie-api-key \
  --out /output
```

## Output Structure

```
public/
  index.html
  sitemap.xml
  {recipe-slug}/
    index.html
    recipe.md
    image.webp      (if recipe has an image)
```

## Running Tests

```sh
go test ./...
```
