# static-mealie

A CLI tool that generates a ready-to-deploy static HTML website full of recipes from your [Mealie](https://mealie.io) instance. Great for sharing your recipes with the world without directly exposing Mealie itself to the internet. Inspired by [iamkirkbater/mealie-markdown-exporter](https://github.com/iamkirkbater/mealie-markdown-exporter).

Check out the **[live demo site](https://abyrne55.github.io/static-mealie/)**!

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

## Quick Start

### Via `mockmealie`

This repo includes a [basic tool](#mealie-api-mocking-tool) for mocking up a Mealie API that you can use to test drive `mealie-static`.

```sh
# Clone this repo
git clone https://github.com/abyrne55/static-mealie.git
cd static-mealie

# Mock up a Mealie API
go run ./cmd/mockmealie --port 8181 &

# Run mealie-static and preview the output
go run . --mealie-url http://localhost:8181 --mealie-token my-fake-token
ls output

# Spin up any basic web server
python3 -m http.server 8282 --directory ./output
```

Once your web server is up, point your browser at `http://127.0.0.1:8282`. You should see a page resembling the [demo site](https://abyrne55.github.io/static-mealie/).

### Via Podman and `demo.mealie.io`

The Mealie devs helpfully host an open-to-the-public demo server at [demo.mealie.io](https://demo.mealie.io) that we can use for testing out this repo's container image.

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
  --publish 8282:8080 \
  quay.io/hummingbird/httpd:latest

# Open it in your browser (on macOS, replace "xdg-open" with just "open")
xdg-open "http://127.0.0.1:8282/" 

# Clean up
read -p "Press Enter to stop demo and clean up" \
  && podman rm -f recipes-httpd \
  && podman volume rm static-mealie-output \
  && podman secret rm static-mealie-token
```

> [!TIP]
> If zero recipes appear, check whether [demo.mealie.io](https://demo.mealie.io) currently has any recipes — the demo site is regularly wiped and may be empty. If it is, [import](https://demo.mealie.io/g/home/r/create/url) a recipe or two ([cacio e pepe](https://cooking.nytimes.com/recipes/1020729-vegan-cacio-e-pepe), perhaps?) and try the commands above again.

## Installation

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

Or just use the container image (see this [note](#container-volume-permissions) on permissions):

```sh
podman run --rm ghcr.io/abyrne55/static-mealie:main --help
```

## Usage

For basic use, all you need is a Mealie instance with some recipes in it and an [API token](https://docs.mealie.io/documentation/getting-started/api-usage/) for said instance.

```sh
static-mealie --mealie-url http://my-mealie-instance --mealie-token my-secret-token
# website will be written to ./output/ by default
```

You should try to avoid exposing your API token in your shell history, of course, so consider reading it from a file. See also *[Providing an API Token](#providing-an-api-token)* below

```sh
static-mealie --mealie-url http://my-mealie-instance --mealie-token file:///path/to/token
```

If you plan on publishing your recipes to the internet, try these output control flags.

```sh
static-mealie --mealie-url http://my-mealie-instance \
  --mealie-token file:///path/to/token \
  --out-title "Family Cookbook" \
  --out-base-url https://example.com/cookbook/
```

If you want to use static-mealie as a cron job, you might find `--clean-slate` helpful. This flag causes `static-mealie` to effectively run `rm -rf /your_output_dir/*` before writing, clearing away any recipes that may have been deleted from your Mealie instance and would have otherwise hung around your static website. Just be sure to carefully set `--out-dir`/`$SM_OUT_DIR` to a folder that you don't mind being deleted.

```sh
static-mealie --mealie-url http://my-mealie-instance \
  --mealie-token file:///path/to/token \
  --out-dir /var/www/html/ \
  --clean-slate
```

## Configuration

| CLI Flag         | Env. Var.         | Description                                                                                                         | Default      |
| ---------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------- | ------------ |
| `--mealie-url`   | `SM_MEALIE_URL`   | Mealie base URL                                                                                                     | *(required)* |
| `--mealie-token` | `SM_MEALIE_TOKEN` | API token or `file:///path`                                                                                         | *(required)* |
| `--out-dir`      | `SM_OUT_DIR`      | Output directory                                                                                                    | `./output`   |
| `--out-title`    | `SM_OUT_TITLE`    | Website title (shown on the index page and on the browser tab title for each recipe)                                | `Recipes`    |
| `--out-base-url` | `SM_OUT_BASE_URL` | Base URL for sitemap/links; output not standards-compliant unless set to an absolute URL like `https://example.com` | `/`          |
| `--clean-slate`  | `SM_CLEAN_SLATE`  | Wipe output directory before building                                                                               | `false`      |
| `--no-clobber`   | `SM_NO_CLOBBER`   | Skip files that already exist                                                                                       | `false`      |
| `-v`             |                   | Verbose logging                                                                                                     | `false`      |

CLI flags take precedence over environmental variables. `--clean-slate` and `--no-clobber` are mutually exclusive.

### Providing an API Token

The Mealie API token is resolved from the first available source:

1. `--mealie-token` flag (supports `file:///path/to/token`)
2. `SM_MEALIE_TOKEN` env var (supports `file:///path/to/token`)
3. `/run/credentials/static-mealie-token` file
4. `/run/secrets/static-mealie-token` file

The credential file paths match the default mount points for systemd/quadlet [credentials](https://systemd.io/CREDENTIALS/) (`LoadCredential=`/`$CREDENTIALS_DIRECTORY`) and Podman/K8s [secrets](https://docs.podman.io/en/latest/markdown/podman-secret-create.1.html).

## Demo Site

**[https://abyrne55.github.io/static-mealie/](https://abyrne55.github.io/static-mealie/)**

A live demo is deployed automatically via GitHub Pages on every push to `main`. It uses the mock Mealie server detailed below.

## Technical Details

### Mealie API Mocking Tool

The `cmd/mockmealie/` directory contains a standalone mock Mealie API server with embedded sample recipes and images. Useful for development and testing without a live Mealie instance:

```sh
go run ./cmd/mockmealie                # starts on :9925
go run . --mealie-url http://localhost:9925 --mealie-token mock --out-dir output -v
```

The mock server accepts any Bearer token and serves 6 sample recipes covering various features (ingredient sections, fractional quantities, inline markdown, notes, and missing images).

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

### Container Volume Permissions

For security, `static-mealie` runs as a non-root user inside the container. This complicates file permissions a bit: files produced by `mealie-static` will be owned by UID/GID 65532, so you'll have to use `sudo` if you wanted to edit/delete those output files. The easiest workarounds are to either use Podman named volumes (as shown in the [Podman Quick Start](#via-podman-and-demomealieio)) or just mount a temp dir so you don't have to worry about cleaning it up yourself.

```sh
TMPDIR=$(mktemp -d)
podman run --rm \
  -v $TMPDIR:/output:U,Z \
  -e SM_MEALIE_URL=http://your-mealie-url \
  -e SM_MEALIE_TOKEN=your-mealie-api-token \
  ghcr.io/abyrne55/static-mealie:main
python3 -m http.server 8282 --directory $TMPDIR
```

Alternatively, you can tell podman to map the in-container UID/GID to your actual UID/GID using `--user-ns`.

```sh
mkdir ./output
podman run --rm \
  --userns=keep-id:uid=65532,gid=65532 \
  -v ./output:/output:Z \
  -e SM_MEALIE_URL=http://your-mealie-url \
  -e SM_MEALIE_TOKEN=your-mealie-api-token \
  ghcr.io/abyrne55/static-mealie:main
python3 -m http.server 8282 --directory ./output
```

### Running Tests

```sh
go test ./...
```
