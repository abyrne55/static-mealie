package main

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"strings"
)

const perPage = 50

type Client struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		APIToken:   apiToken,
		HTTPClient: &http.Client{},
	}
}

type PaginatedResponse struct {
	Page       int      `json:"page"`
	PerPage    int      `json:"perPage"`
	Total      int      `json:"total"`
	TotalPages int      `json:"totalPages"`
	Items      []Recipe `json:"items"`
}

type Recipe struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Slug               string              `json:"slug"`
	Image              any                 `json:"image"`
	RecipeYield            string              `json:"recipeYield"`
	RecipeYieldQuantity    float64             `json:"recipeYieldQuantity"`
	TotalTime          string              `json:"totalTime"`
	PrepTime           string              `json:"prepTime"`
	CookTime           string              `json:"cookTime"`
	PerformTime        string              `json:"performTime"`
	Description        string              `json:"description"`
	DateAdded          string              `json:"dateAdded"`
	DateUpdated        string              `json:"dateUpdated"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
	RecipeIngredient   []Ingredient        `json:"recipeIngredient"`
	RecipeInstructions []RecipeInstruction  `json:"recipeInstructions"`
	Notes              []Note              `json:"notes"`
	OrgURL             string              `json:"orgURL"`
}

type Ingredient struct {
	Quantity     float64         `json:"quantity"`
	Unit         *IngredientUnit `json:"unit"`
	Food         *IngredientFood `json:"food"`
	Note         string          `json:"note"`
	Display      string          `json:"display"`
	Title        string          `json:"title"`
	OriginalText string          `json:"originalText"`
}

type IngredientUnit struct {
	Name            string `json:"name"`
	PluralName      string `json:"pluralName"`
	Abbreviation    string `json:"abbreviation"`
	UseAbbreviation bool   `json:"useAbbreviation"`
	Fraction        bool   `json:"fraction"`
}

type IngredientFood struct {
	Name       string `json:"name"`
	PluralName string `json:"pluralName"`
}

type RecipeInstruction struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type Note struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type RenderedIngredient struct {
	IsHeading bool
	Heading   string
	AmountStr string
	FoodStr   string
	NoteStr   string
	Display   string
}

func (ing *Ingredient) Render() RenderedIngredient {
	if ing.Title != "" {
		return RenderedIngredient{IsHeading: true, Heading: ing.Title}
	}

	if ing.Unit == nil && ing.Food == nil {
		display := ing.Display
		if display == "" {
			display = ing.Note
		}
		return RenderedIngredient{Display: display}
	}

	ri := RenderedIngredient{}

	var amountParts []string
	if ing.Quantity > 0 {
		amountParts = append(amountParts, formatQuantity(ing.Quantity))
	}
	if ing.Unit != nil && ing.Unit.Name != "" {
		name := ing.Unit.Name
		if ing.Unit.UseAbbreviation && ing.Unit.Abbreviation != "" {
			name = ing.Unit.Abbreviation
		} else if ing.Quantity > 1 && ing.Unit.PluralName != "" {
			name = ing.Unit.PluralName
		}
		amountParts = append(amountParts, name)
	}
	ri.AmountStr = strings.Join(amountParts, " ")

	if ing.Food != nil && ing.Food.Name != "" {
		name := ing.Food.Name
		if ing.Unit == nil && ing.Quantity > 1 && ing.Food.PluralName != "" {
			name = ing.Food.PluralName
		}
		ri.FoodStr = name
	}

	ri.NoteStr = ing.Note

	return ri
}

var fractions = []struct {
	value float64
	char  string
}{
	{0.125, "⅛"},
	{0.25, "¼"},
	{1.0 / 3.0, "⅓"},
	{0.375, "⅜"},
	{0.5, "½"},
	{0.625, "⅝"},
	{2.0 / 3.0, "⅔"},
	{0.75, "¾"},
	{0.875, "⅞"},
}

func formatQuantity(q float64) string {
	if q <= 0 {
		return ""
	}

	whole := int(math.Floor(q))
	frac := q - float64(whole)

	const tol = 0.01

	if frac < tol {
		return fmt.Sprintf("%d", whole)
	}

	for _, f := range fractions {
		if math.Abs(frac-f.value) < tol {
			if whole > 0 {
				return fmt.Sprintf("%d%s", whole, f.char)
			}
			return f.char
		}
	}

	if q == math.Floor(q) {
		return fmt.Sprintf("%d", int(q))
	}
	s := fmt.Sprintf("%.2f", q)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

var (
	reBold   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reItalic = regexp.MustCompile(`(?:\*(.+?)\*|_(.+?)_)`)
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

func renderInlineMarkdown(s string) template.HTML {
	s = html.EscapeString(s)
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	s = reItalic.ReplaceAllStringFunc(s, func(m string) string {
		sub := reItalic.FindStringSubmatch(m)
		content := sub[1]
		if content == "" {
			content = sub[2]
		}
		return "<em>" + content + "</em>"
	})
	s = reLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	return template.HTML(s)
}

func (c *Client) GetAllRecipes() ([]Recipe, error) {
	var allRecipes []Recipe
	page := 1

	for {
		slog.Debug("fetching recipes", "page", page)
		resp, err := c.getRecipesPage(page)
		if err != nil {
			return nil, fmt.Errorf("fetch recipes page %d: %w", page, err)
		}

		allRecipes = append(allRecipes, resp.Items...)

		slog.Debug("fetched page", "page", page, "totalPages", resp.TotalPages, "count", len(resp.Items))

		if page >= resp.TotalPages || len(resp.Items) == 0 {
			break
		}
		page++
	}

	slog.Debug("fetched all recipes", "total", len(allRecipes))
	return allRecipes, nil
}

func (c *Client) GetRecipe(slug string) (*Recipe, error) {
	url := fmt.Sprintf("%s/api/recipes/%s", c.BaseURL, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var recipe Recipe
	if err := json.NewDecoder(resp.Body).Decode(&recipe); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &recipe, nil
}

func (c *Client) GetRecipeImage(recipeID string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/media/recipes/%s/images/original.webp", c.BaseURL, recipeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching image for recipe %s", resp.StatusCode, recipeID)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read image data: %w", err)
	}

	return data, nil
}

func (c *Client) getRecipesPage(page int) (*PaginatedResponse, error) {
	url := fmt.Sprintf("%s/api/recipes?page=%d&perPage=%d", c.BaseURL, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var paginatedResp PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginatedResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &paginatedResp, nil
}
