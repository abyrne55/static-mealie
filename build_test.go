package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testRecipes() []Recipe {
	return []Recipe{
		{
			ID:          "id-1",
			Name:        "Test Pasta",
			Slug:        "test-pasta",
			Description: "A simple pasta dish",
			RecipeYield: "4",
			DateAdded:   "2026-05-15",
			RecipeIngredient: []Ingredient{
				{Quantity: 1, Unit: &IngredientUnit{Name: "pound"}, Food: &IngredientFood{Name: "pasta"}},
				{Quantity: 2, Unit: &IngredientUnit{Name: "tablespoon", PluralName: "tablespoons"}, Food: &IngredientFood{Name: "olive oil"}},
				{Title: "For the sauce", Quantity: 0.5, Unit: &IngredientUnit{Name: "cup"}, Food: &IngredientFood{Name: "cream"}},
			},
			RecipeInstructions: []RecipeInstruction{
				{Text: "Boil pasta in salted water."},
				{Text: "Mix sauce ingredients."},
			},
			Notes: []Note{
				{Title: "Tip", Text: "Use **fresh** pasta for best results."},
			},
			OrgURL: "https://example.com/pasta",
		},
		{
			ID:        "id-2",
			Name:      "Simple Salad",
			Slug:      "simple-salad",
			DateAdded: "2026-05-20",
			RecipeIngredient: []Ingredient{
				{Display: "mixed greens"},
				{Quantity: 1, Food: &IngredientFood{Name: "tomato"}},
			},
			RecipeInstructions: []RecipeInstruction{
				{Text: "Toss everything together."},
			},
		},
	}
}

func TestBuildSiteStructure(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{
		Title:   "Test Recipes",
		SiteURL: "https://example.com/",
		OutDir:  outDir,
	}

	recipes := testRecipes()
	images := map[string][]byte{
		"id-1": []byte("fake-webp-data"),
	}

	if err := site.Build(recipes, images); err != nil {
		t.Fatalf("Build: %v", err)
	}

	expected := []string{
		"index.html",
		"sitemap.xml",
		"test-pasta/index.html",
		"test-pasta/recipe.md",
		"test-pasta/image.webp",
		"simple-salad/index.html",
		"simple-salad/recipe.md",
	}

	for _, path := range expected {
		full := filepath.Join(outDir, path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("missing file: %s", path)
		}
	}

	noImage := filepath.Join(outDir, "simple-salad/image.webp")
	if _, err := os.Stat(noImage); !os.IsNotExist(err) {
		t.Errorf("simple-salad should not have image.webp")
	}
}

func TestBuildIndexContent(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "My Recipes", SiteURL: "/", OutDir: outDir}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	html := string(data)

	checks := []string{
		"<h1>My Recipes</h1>",
		`<title>My Recipes</title>`,
		"/test-pasta/",
		"/simple-salad/",
		"Test Pasta",
		"Simple Salad",
		"image.webp",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("index.html missing %q", check)
		}
	}

	salIdx := strings.Index(html, "Simple Salad")
	pastaIdx := strings.Index(html, "Test Pasta")
	if salIdx > pastaIdx {
		t.Error("recipes not sorted by date (newest first)")
	}
}

func TestBuildRecipeContent(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "https://example.com/", OutDir: outDir}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "test-pasta/index.html"))
	if err != nil {
		t.Fatalf("read recipe html: %v", err)
	}
	html := string(data)

	checks := []string{
		"<h1>Test Pasta</h1>",
		"application/ld+json",
		`"@type": "Recipe"`,
		"recipe.md",
		"image.webp",
		"Ingredients",
		"Instructions",
		"For the sauce",
		"1 pound",
		"Notes",
		"Tip",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("recipe html missing %q", check)
		}
	}
}

func TestBuildMarkdown(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "/", OutDir: outDir}
	if err := site.Build(testRecipes(), nil); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "test-pasta/recipe.md"))
	if err != nil {
		t.Fatalf("read recipe.md: %v", err)
	}
	md := string(data)

	checks := []string{
		"# Test Pasta",
		"- [ ]",
		"## Ingredients",
		"## Instructions",
		"1. Boil pasta",
		"## Notes",
		"### Tip",
	}
	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("recipe.md missing %q", check)
		}
	}
}

func TestBuildSitemap(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "https://example.com", OutDir: outDir}
	if err := site.Build(testRecipes(), nil); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("read sitemap.xml: %v", err)
	}
	xml := string(data)

	checks := []string{
		`<?xml version="1.0"`,
		"https://example.com/",
		"https://example.com/test-pasta/",
		"https://example.com/simple-salad/",
		"<lastmod>2026-05-15</lastmod>",
		"<lastmod>2026-05-20</lastmod>",
	}
	for _, check := range checks {
		if !strings.Contains(xml, check) {
			t.Errorf("sitemap.xml missing %q", check)
		}
	}
}

func TestBuildCleanSlate(t *testing.T) {
	outDir := t.TempDir()

	staleFile := filepath.Join(outDir, "stale.html")
	if err := os.WriteFile(staleFile, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}
	staleDir := filepath.Join(outDir, "old-recipe")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleDir, "index.html"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale dir file: %v", err)
	}

	site := &Site{Title: "Recipes", SiteURL: "/", OutDir: outDir, CleanSlate: true}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Error("stale.html should have been removed")
	}
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("old-recipe/ should have been removed")
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Error("fresh index.html should exist")
	}
	if _, err := os.Stat(filepath.Join(outDir, "test-pasta/index.html")); err != nil {
		t.Error("fresh recipe should exist")
	}
}

func TestBuildNoClobber(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "/", OutDir: outDir}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("original-img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	sentinel := []byte("hand-edited")
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), sentinel, 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	site.NoClobber = true
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("new-img")}); err != nil {
		t.Fatalf("Build with NoClobber: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(outDir, "index.html"))
	if string(data) != "hand-edited" {
		t.Error("index.html should have been preserved by no-clobber")
	}

	imgData, _ := os.ReadFile(filepath.Join(outDir, "test-pasta/image.webp"))
	if string(imgData) != "original-img" {
		t.Error("image.webp should have been preserved by no-clobber")
	}
}

func TestBuildNoClobberWritesNewFiles(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "/", OutDir: outDir, NoClobber: true}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	expected := []string{
		"index.html",
		"sitemap.xml",
		"test-pasta/index.html",
		"test-pasta/recipe.md",
		"test-pasta/image.webp",
		"simple-salad/index.html",
		"simple-salad/recipe.md",
	}
	for _, path := range expected {
		if _, err := os.Stat(filepath.Join(outDir, path)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", path)
		}
	}
}

func TestBasePath(t *testing.T) {
	tests := []struct {
		siteURL string
		want    string
	}{
		{"/", "/"},
		{"https://example.com", "/"},
		{"https://example.com/", "/"},
		{"https://example.com/sub", "/sub/"},
		{"https://example.com/sub/", "/sub/"},
		{"https://user.github.io/repo", "/repo/"},
		{"https://user.github.io/repo/", "/repo/"},
	}
	for _, tt := range tests {
		got := basePath(tt.siteURL)
		if got != tt.want {
			t.Errorf("basePath(%q) = %q, want %q", tt.siteURL, got, tt.want)
		}
	}
}

func TestBuildSubpathLinks(t *testing.T) {
	outDir := t.TempDir()

	site := &Site{Title: "Recipes", SiteURL: "https://user.github.io/my-recipes/", OutDir: outDir}
	if err := site.Build(testRecipes(), map[string][]byte{"id-1": []byte("img")}); err != nil {
		t.Fatalf("Build: %v", err)
	}

	indexData, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	indexHTML := string(indexData)

	indexChecks := []string{
		`href="/my-recipes/test-pasta/"`,
		`href="/my-recipes/simple-salad/"`,
		`src="/my-recipes/test-pasta/image.webp"`,
	}
	for _, check := range indexChecks {
		if !strings.Contains(indexHTML, check) {
			t.Errorf("index.html missing %q", check)
		}
	}
	for _, bad := range []string{`href="/test-pasta/"`, `href="/simple-salad/"`, `src="/test-pasta/image.webp"`} {
		if strings.Contains(indexHTML, bad) {
			t.Errorf("index.html should not contain root-relative %q", bad)
		}
	}

	recipeData, err := os.ReadFile(filepath.Join(outDir, "test-pasta/index.html"))
	if err != nil {
		t.Fatalf("read recipe html: %v", err)
	}
	recipeHTML := string(recipeData)

	recipeChecks := []string{
		`href="/my-recipes/"`,
		`href="/my-recipes/test-pasta/recipe.md"`,
		`src="/my-recipes/test-pasta/image.webp"`,
	}
	for _, check := range recipeChecks {
		if !strings.Contains(recipeHTML, check) {
			t.Errorf("recipe html missing %q", check)
		}
	}
}

func TestBuildJSONLD(t *testing.T) {
	rv := recipeToView(testRecipes()[0], true)
	jsonLD, err := buildJSONLD(rv, "https://example.com/")
	if err != nil {
		t.Fatalf("buildJSONLD: %v", err)
	}

	s := string(jsonLD)
	checks := []string{
		`"@context": "https://schema.org"`,
		`"@type": "Recipe"`,
		`"name": "Test Pasta"`,
		`"recipeYield": "4"`,
		`"@type": "HowToStep"`,
		`"image": "https://example.com/test-pasta/image.webp"`,
	}
	for _, check := range checks {
		if !strings.Contains(s, check) {
			t.Errorf("JSON-LD missing %q", check)
		}
	}
}
