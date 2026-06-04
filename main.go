package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func main() {
	var (
		urlFlag     = flag.String("mealie-url", "", "Mealie base URL (env: SM_MEALIE_URL)")
		tokenFlag   = flag.String("mealie-token", "", "API token or file:///path (env: SM_MEALIE_TOKEN)")
		outFlag     = flag.String("out-dir", "public", "Output directory (env: SM_OUT_DIR)")
		titleFlag   = flag.String("out-title", "Recipes", "Site title (env: SM_OUT_TITLE)")
		siteURLFlag    = flag.String("out-base-url", "/", "Base URL for sitemap/links; output not standards-compliant unless set to an absolute URL like https://example.com (env: SM_OUT_BASE_URL)")
		cleanSlateFlag = flag.Bool("clean-slate", false, "Wipe output directory before building (env: SM_CLEAN_SLATE)")
		noClobberFlag  = flag.Bool("no-clobber", false, "Skip files that already exist (env: SM_NO_CLOBBER)")
		verbose        = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	mealieURL := *urlFlag
	if mealieURL == "" {
		mealieURL = os.Getenv("SM_MEALIE_URL")
	}
	if mealieURL == "" {
		fmt.Fprintln(os.Stderr, "error: --mealie-url or SM_MEALIE_URL required")
		flag.Usage()
		os.Exit(1)
	}

	credentialFiles := []string{
		"/run/credentials/static-mealie-token",
		"/run/secrets/static-mealie-token",
	}

	token := *tokenFlag
	if token == "" {
		token = os.Getenv("SM_MEALIE_TOKEN")
	}
	if token != "" {
		resolved, err := resolveToken(token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error resolving token: %v\n", err)
			os.Exit(1)
		}
		token = resolved
	} else {
		for _, path := range credentialFiles {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			token = strings.TrimSpace(string(data))
			if token != "" {
				break
			}
		}
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: no API token found; tried --mealie-token flag, SM_MEALIE_TOKEN env var, /run/credentials/static-mealie-token, /run/secrets/static-mealie-token")
		flag.Usage()
		os.Exit(1)
	}

	cleanSlate := *cleanSlateFlag
	if !isFlagSet("clean-slate") {
		if v := os.Getenv("SM_CLEAN_SLATE"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid SM_CLEAN_SLATE value %q\n", v)
				os.Exit(1)
			}
			cleanSlate = b
		}
	}

	noClobber := *noClobberFlag
	if !isFlagSet("no-clobber") {
		if v := os.Getenv("SM_NO_CLOBBER"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid SM_NO_CLOBBER value %q\n", v)
				os.Exit(1)
			}
			noClobber = b
		}
	}

	if cleanSlate && noClobber {
		fmt.Fprintln(os.Stderr, "error: --clean-slate and --no-clobber are mutually exclusive")
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

	outDir := *outFlag
	if !isFlagSet("out-dir") {
		if v := os.Getenv("SM_OUT_DIR"); v != "" {
			outDir = v
		}
	}
	title := *titleFlag
	if !isFlagSet("out-title") {
		if v := os.Getenv("SM_OUT_TITLE"); v != "" {
			title = v
		}
	}
	siteURL := *siteURLFlag
	if !isFlagSet("out-base-url") {
		if v := os.Getenv("SM_OUT_BASE_URL"); v != "" {
			siteURL = v
		}
	}

	site := &Site{
		Title:      title,
		SiteURL:    siteURL,
		OutDir:     outDir,
		CleanSlate: cleanSlate,
		NoClobber:  noClobber,
	}

	slog.Info("building site", "out", outDir)
	if err := site.Build(recipes, images); err != nil {
		fmt.Fprintf(os.Stderr, "error building site: %v\n", err)
		os.Exit(1)
	}

	slog.Info("done", "recipes", len(recipes), "images", len(images), "out", outDir)
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
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
