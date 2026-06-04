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
static-mealie --mealie-url https://your-mealie-instance --mealie-token your-token
```

### Flags

```
--mealie-url string      Mealie base URL (env: SM_MEALIE_URL)
--mealie-token string    API token or file:///path (env: SM_MEALIE_TOKEN)
--out-dir string         Output directory (default: "public", env: SM_OUT_DIR)
--out-title string       Site title (default: "Recipes", env: SM_OUT_TITLE)
--out-base-url string    Base URL for sitemap/links; output not standards-compliant
                         unless set to an absolute URL like https://example.com
                         (default: "/", env: SM_OUT_BASE_URL)
-v                       Verbose logging
```

### Token Resolution

The API token is resolved from the first available source:

1. `--mealie-token` flag (supports `file:///path/to/token`)
2. `SM_MEALIE_TOKEN` env var (supports `file:///path/to/token`)
3. `/run/credentials/static-mealie-token` file
4. `/run/secrets/static-mealie-token` file

The credential file paths match the default mount points for Podman secrets and Kubernetes secrets/projected volumes.

## Container

Images are published to `ghcr.io/abyrne55/static-mealie` for `linux/amd64` and `linux/arm64`.

```sh
# Store your API token as a Podman secret
podman secret create static-mealie-token /path/to/token

# Generate the site into ./output
podman run --rm \
  -v ./output:/output:U,Z \
  --secret static-mealie-token \
  ghcr.io/abyrne55/static-mealie:main \
  --mealie-url https://your-mealie-instance \
  --out-dir /output
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
