package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func main() {
	var (
		urlFlag     = flag.String("url", "", "Mealie base URL (env: MEALIE_URL)")
		tokenFlag   = flag.String("token", "", "API token or file:///path (env: MEALIE_TOKEN)")
		outFlag     = flag.String("out", "public", "Output directory")
		titleFlag   = flag.String("title", "Recipes", "Site title")
		siteURLFlag = flag.String("site-url", "/", "Base URL for sitemap/links")
		verbose     = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	mealieURL := *urlFlag
	if mealieURL == "" {
		mealieURL = os.Getenv("MEALIE_URL")
	}
	if mealieURL == "" {
		fmt.Fprintln(os.Stderr, "error: --url or MEALIE_URL required")
		flag.Usage()
		os.Exit(1)
	}

	token := *tokenFlag
	if token == "" {
		token = os.Getenv("MEALIE_TOKEN")
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: --token or MEALIE_TOKEN required")
		flag.Usage()
		os.Exit(1)
	}

	token, err := resolveToken(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving token: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(mealieURL, token)

	slog.Info("fetching recipe list")
	summaries, err := client.GetAllRecipes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching recipes: %v\n", err)
		os.Exit(1)
	}
	slog.Info("found recipes", "count", len(summaries))

	var recipes []Recipe
	images := make(map[string][]byte)

	for i, s := range summaries {
		slog.Info("fetching recipe", "slug", s.Slug, "progress", fmt.Sprintf("%d/%d", i+1, len(summaries)))
		r, err := client.GetRecipe(s.Slug)
		if err != nil {
			slog.Error("failed to fetch recipe", "slug", s.Slug, "error", err)
			continue
		}
		recipes = append(recipes, *r)

		img, err := client.GetRecipeImage(r.ID)
		if err != nil {
			slog.Warn("failed to fetch image", "slug", s.Slug, "error", err)
		} else if img != nil {
			images[r.ID] = img
			slog.Debug("fetched image", "slug", s.Slug, "bytes", len(img))
		}
	}

	site := &Site{
		Title:   *titleFlag,
		SiteURL: *siteURLFlag,
		OutDir:  *outFlag,
	}

	slog.Info("building site", "out", *outFlag)
	if err := site.Build(recipes, images); err != nil {
		fmt.Fprintf(os.Stderr, "error building site: %v\n", err)
		os.Exit(1)
	}

	slog.Info("done", "recipes", len(recipes), "images", len(images), "out", *outFlag)
}

func resolveToken(token string) (string, error) {
	if path, ok := strings.CutPrefix(token, "file://"); ok {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read token file %s: %w", path, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return token, nil
}
