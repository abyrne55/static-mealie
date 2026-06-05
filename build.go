package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

type Site struct {
	Title      string
	SiteURL    string
	OutDir     string
	CleanSlate bool
	NoClobber  bool
}

func (s *Site) skipIfExists(path string) bool {
	if !s.NoClobber {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		slog.Debug("no-clobber: skipping existing file", "path", path)
		return true
	}
	return false
}

type RecipeView struct {
	Name          string
	Slug          string
	Description   string
	RecipeYield   string
	PrepTime      string
	CookTime      string
	TotalTime     string
	PerformTime   string
	DateISO       string
	DateFormatted string
	HasImage      bool
	Ingredients   []RenderedIngredient
	Instructions  []RecipeInstruction
	Notes         []Note
	OrgURL        string
}

type IndexData struct {
	SiteTitle string
	Recipes   []RecipeView
}

type RecipeData struct {
	SiteTitle string
	Recipe    RecipeView
	JSONLD    template.HTML
}

type SitemapData struct {
	SiteURL string
	Recipes []RecipeView
}

type SchemaRecipe struct {
	Context            string       `json:"@context"`
	Type               string       `json:"@type"`
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	Image              string       `json:"image,omitempty"`
	RecipeYield        string       `json:"recipeYield,omitempty"`
	RecipeIngredient   []string     `json:"recipeIngredient,omitempty"`
	RecipeInstructions []SchemaStep `json:"recipeInstructions,omitempty"`
	DatePublished      string       `json:"datePublished,omitempty"`
}

type SchemaStep struct {
	Type string `json:"@type"`
	Text string `json:"text"`
}

func formatYield(yield string, qty, servings float64) string {
	if qty > 0 && yield != "" {
		q := int(qty)
		if float64(q) == qty {
			return fmt.Sprintf("%d %s", q, yield)
		}
		return fmt.Sprintf("%g %s", qty, yield)
	}
	if qty > 0 {
		q := int(qty)
		if float64(q) == qty {
			return fmt.Sprintf("%d", q)
		}
		return fmt.Sprintf("%g", qty)
	}
	if yield != "" {
		return yield
	}
	if servings > 0 {
		s := int(servings)
		if float64(s) == servings {
			return fmt.Sprintf("%d servings", s)
		}
		return fmt.Sprintf("%g servings", servings)
	}
	return ""
}

func recipeToView(r Recipe, hasImage bool) RecipeView {
	rv := RecipeView{
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		RecipeYield: formatYield(r.RecipeYield, r.RecipeYieldQuantity, r.RecipeServings),
		PrepTime:    r.PrepTime,
		CookTime:    r.CookTime,
		TotalTime:   r.TotalTime,
		PerformTime: r.PerformTime,
		HasImage:    hasImage,
		Notes:       r.Notes,
		OrgURL:      r.OrgURL,
	}

	rv.Instructions = r.RecipeInstructions

	dateStr := r.DateAdded
	if dateStr == "" {
		dateStr = r.CreatedAt
	}
	if dateStr != "" {
		rv.DateISO = dateStr[:min(len(dateStr), 10)]
		if t, err := time.Parse("2006-01-02", rv.DateISO); err == nil {
			rv.DateFormatted = t.Format("January 2, 2006")
		}
	}

	for i := range r.RecipeIngredient {
		rv.Ingredients = append(rv.Ingredients, r.RecipeIngredient[i].Render())
	}

	return rv
}

func buildJSONLD(rv RecipeView, siteURL string) (template.HTML, error) {
	sr := SchemaRecipe{
		Context:       "https://schema.org",
		Type:          "Recipe",
		Name:          rv.Name,
		Description:   rv.Description,
		RecipeYield:   rv.RecipeYield,
		DatePublished: rv.DateISO,
	}

	if rv.HasImage {
		sr.Image = strings.TrimRight(siteURL, "/") + "/" + rv.Slug + "/image.webp"
	}

	for _, ing := range rv.Ingredients {
		if ing.IsHeading {
			continue
		}
		var s string
		if ing.Display != "" {
			s = ing.Display
		} else {
			parts := []string{}
			if ing.AmountStr != "" {
				parts = append(parts, ing.AmountStr)
			}
			if ing.FoodStr != "" {
				parts = append(parts, ing.FoodStr)
			}
			s = strings.Join(parts, " ")
			if ing.NoteStr != "" {
				s += ", " + ing.NoteStr
			}
		}
		if s != "" {
			sr.RecipeIngredient = append(sr.RecipeIngredient, s)
		}
	}

	for _, inst := range rv.Instructions {
		sr.RecipeInstructions = append(sr.RecipeInstructions, SchemaStep{
			Type: "HowToStep",
			Text: inst.Text,
		})
	}

	data, err := json.MarshalIndent(sr, "    ", "  ")
	if err != nil {
		return "", err
	}

	tag := fmt.Sprintf(`<script type="application/ld+json">
    %s
  </script>`, string(data))
	return template.HTML(tag), nil
}

func (s *Site) Build(recipes []Recipe, images map[string][]byte) error {
	fm := template.FuncMap{
		"md": renderInlineMarkdown,
	}

	sub, err := fs.Sub(templateFS, "templates")
	if err != nil {
		return fmt.Errorf("template fs: %w", err)
	}

	base, err := template.New("").Funcs(fm).ParseFS(sub, "base.html.tmpl")
	if err != nil {
		return fmt.Errorf("parse base template: %w", err)
	}

	indexT, err := template.Must(base.Clone()).ParseFS(sub, "index.html.tmpl")
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}

	recipeT, err := template.Must(base.Clone()).ParseFS(sub, "recipe.html.tmpl")
	if err != nil {
		return fmt.Errorf("parse recipe template: %w", err)
	}

	mdT, err := texttemplate.New("recipe.md.tmpl").Funcs(texttemplate.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).ParseFS(sub, "recipe.md.tmpl")
	if err != nil {
		return fmt.Errorf("parse markdown template: %w", err)
	}

	sitemapT, err := texttemplate.New("sitemap.xml.tmpl").ParseFS(sub, "sitemap.xml.tmpl")
	if err != nil {
		return fmt.Errorf("parse sitemap template: %w", err)
	}

	var views []RecipeView
	for _, r := range recipes {
		_, hasImage := images[r.ID]
		views = append(views, recipeToView(r, hasImage))
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].DateISO > views[j].DateISO
	})

	if err := os.MkdirAll(s.OutDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if s.CleanSlate {
		slog.Debug("clean-slate: wiping output directory", "path", s.OutDir)
		entries, err := os.ReadDir(s.OutDir)
		if err != nil {
			return fmt.Errorf("read output dir for clean-slate: %w", err)
		}
		for _, e := range entries {
			if err := os.RemoveAll(filepath.Join(s.OutDir, e.Name())); err != nil {
				return fmt.Errorf("clean-slate remove %s: %w", e.Name(), err)
			}
		}
	}

	indexPath := filepath.Join(s.OutDir, "index.html")
	if !s.skipIfExists(indexPath) {
		indexFile, err := os.Create(indexPath)
		if err != nil {
			return fmt.Errorf("create index.html: %w", err)
		}
		if err := indexT.ExecuteTemplate(indexFile, "base.html.tmpl", IndexData{SiteTitle: s.Title, Recipes: views}); err != nil {
			_ = indexFile.Close()
			return fmt.Errorf("execute index template: %w", err)
		}
		if err := indexFile.Close(); err != nil {
			return fmt.Errorf("close index.html: %w", err)
		}
	}

	for _, rv := range views {
		recipeDir := filepath.Join(s.OutDir, rv.Slug)
		if err := os.MkdirAll(recipeDir, 0o755); err != nil {
			return fmt.Errorf("create recipe dir %s: %w", rv.Slug, err)
		}

		recipeHTMLPath := filepath.Join(recipeDir, "index.html")
		if !s.skipIfExists(recipeHTMLPath) {
			jsonLD, err := buildJSONLD(rv, s.SiteURL)
			if err != nil {
				return fmt.Errorf("build JSON-LD for %s: %w", rv.Slug, err)
			}

			htmlFile, err := os.Create(recipeHTMLPath)
			if err != nil {
				return fmt.Errorf("create recipe html %s: %w", rv.Slug, err)
			}
			if err := recipeT.ExecuteTemplate(htmlFile, "base.html.tmpl", RecipeData{SiteTitle: s.Title, Recipe: rv, JSONLD: jsonLD}); err != nil {
				_ = htmlFile.Close()
				return fmt.Errorf("execute recipe template %s: %w", rv.Slug, err)
			}
			if err := htmlFile.Close(); err != nil {
				return fmt.Errorf("close recipe html %s: %w", rv.Slug, err)
			}
		}

		recipeMDPath := filepath.Join(recipeDir, "recipe.md")
		if !s.skipIfExists(recipeMDPath) {
			mdFile, err := os.Create(recipeMDPath)
			if err != nil {
				return fmt.Errorf("create recipe md %s: %w", rv.Slug, err)
			}
			if _, err := mdFile.Write([]byte("\xEF\xBB\xBF")); err != nil { // UTF-8 BOM so browsers don't fall back to Latin-1
				_ = mdFile.Close()
				return fmt.Errorf("write BOM for %s: %w", rv.Slug, err)
			}
			if err := mdT.Execute(mdFile, rv); err != nil {
				_ = mdFile.Close()
				return fmt.Errorf("execute md template %s: %w", rv.Slug, err)
			}
			if err := mdFile.Close(); err != nil {
				return fmt.Errorf("close recipe md %s: %w", rv.Slug, err)
			}
		}
	}

	for id, data := range images {
		var slug string
		for _, r := range recipes {
			if r.ID == id {
				slug = r.Slug
				break
			}
		}
		if slug == "" {
			continue
		}
		imgPath := filepath.Join(s.OutDir, slug, "image.webp")
		if !s.skipIfExists(imgPath) {
			if err := os.WriteFile(imgPath, data, 0o644); err != nil {
				return fmt.Errorf("write image %s: %w", slug, err)
			}
		}
	}

	sitemapPath := filepath.Join(s.OutDir, "sitemap.xml")
	if s.skipIfExists(sitemapPath) {
		return nil
	}
	sitemapFile, err := os.Create(sitemapPath)
	if err != nil {
		return fmt.Errorf("create sitemap.xml: %w", err)
	}

	siteURL := strings.TrimRight(s.SiteURL, "/") + "/"
	if err := sitemapT.Execute(sitemapFile, SitemapData{SiteURL: siteURL, Recipes: views}); err != nil {
		_ = sitemapFile.Close()
		return fmt.Errorf("execute sitemap template: %w", err)
	}
	if err := sitemapFile.Close(); err != nil {
		return fmt.Errorf("close sitemap.xml: %w", err)
	}

	return nil
}
