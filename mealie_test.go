package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllRecipesPagination(t *testing.T) {
	page1 := PaginatedResponse{
		Page: 1, PerPage: 2, Total: 3, TotalPages: 2,
		Items: []Recipe{
			{ID: "1", Name: "Recipe One", Slug: "recipe-one"},
			{ID: "2", Name: "Recipe Two", Slug: "recipe-two"},
		},
	}
	page2 := PaginatedResponse{
		Page: 2, PerPage: 2, Total: 3, TotalPages: 2,
		Items: []Recipe{
			{ID: "3", Name: "Recipe Three", Slug: "recipe-three"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header")
		}
		q := r.URL.Query()
		switch q.Get("page") {
		case "1", "":
			_ = json.NewEncoder(w).Encode(page1)
		case "2":
			_ = json.NewEncoder(w).Encode(page2)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	recipes, err := c.GetAllRecipes()
	if err != nil {
		t.Fatalf("GetAllRecipes: %v", err)
	}
	if len(recipes) != 3 {
		t.Fatalf("got %d recipes, want 3", len(recipes))
	}
	if recipes[2].Name != "Recipe Three" {
		t.Errorf("third recipe name = %q, want %q", recipes[2].Name, "Recipe Three")
	}
}

func TestGetRecipe(t *testing.T) {
	recipe := Recipe{
		ID: "abc", Name: "Test Recipe", Slug: "test-recipe",
		Description: "A test",
		RecipeIngredient: []Ingredient{
			{Quantity: 2, Unit: &IngredientUnit{Name: "cups"}, Food: &IngredientFood{Name: "flour"}, Note: "sifted"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/recipes/test-recipe" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(recipe)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	got, err := c.GetRecipe("test-recipe")
	if err != nil {
		t.Fatalf("GetRecipe: %v", err)
	}
	if got.Name != "Test Recipe" {
		t.Errorf("name = %q, want %q", got.Name, "Test Recipe")
	}
	if len(got.RecipeIngredient) != 1 {
		t.Fatalf("ingredients = %d, want 1", len(got.RecipeIngredient))
	}
	if got.RecipeIngredient[0].Unit.Name != "cups" {
		t.Errorf("unit = %q, want %q", got.RecipeIngredient[0].Unit.Name, "cups")
	}
}

func TestGetRecipeImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/media/recipes/abc/images/original.webp" {
			_, _ = w.Write([]byte("fake-image-data"))
			return
		}
		if r.URL.Path == "/api/media/recipes/none/images/original.webp" {
			http.NotFound(w, r)
			return
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")

	data, err := c.GetRecipeImage("abc")
	if err != nil {
		t.Fatalf("GetRecipeImage: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Errorf("image data = %q", string(data))
	}

	data, err = c.GetRecipeImage("none")
	if err != nil {
		t.Fatalf("GetRecipeImage 404: %v", err)
	}
	if data != nil {
		t.Errorf("expected nil for 404, got %d bytes", len(data))
	}
}

func TestIngredientRender(t *testing.T) {
	tests := []struct {
		name string
		ing  Ingredient
		want RenderedIngredient
	}{
		{
			name: "full ingredient",
			ing:  Ingredient{Quantity: 2, Unit: &IngredientUnit{Name: "cup", PluralName: "cups"}, Food: &IngredientFood{Name: "flour"}, Note: "sifted"},
			want: RenderedIngredient{AmountStr: "2 cups", FoodStr: "flour", NoteStr: "sifted"},
		},
		{
			name: "section heading",
			ing:  Ingredient{Title: "For the sauce", Quantity: 1, Unit: &IngredientUnit{Name: "pound"}, Food: &IngredientFood{Name: "ground beef"}},
			want: RenderedIngredient{IsHeading: true, Heading: "For the sauce", AmountStr: "1 pound", FoodStr: "ground beef"},
		},
		{
			name: "no unit no food — display fallback",
			ing:  Ingredient{Display: "a pinch of salt"},
			want: RenderedIngredient{Display: "a pinch of salt"},
		},
		{
			name: "no unit no food — note fallback",
			ing:  Ingredient{Note: "salt to taste"},
			want: RenderedIngredient{Display: "salt to taste"},
		},
		{
			name: "food but no unit",
			ing:  Ingredient{Quantity: 3, Food: &IngredientFood{Name: "egg", PluralName: "eggs"}},
			want: RenderedIngredient{AmountStr: "3", FoodStr: "eggs"},
		},
		{
			name: "zero quantity",
			ing:  Ingredient{Unit: &IngredientUnit{Name: "cup"}, Food: &IngredientFood{Name: "water"}},
			want: RenderedIngredient{AmountStr: "cup", FoodStr: "water"},
		},
		{
			name: "abbreviation",
			ing:  Ingredient{Quantity: 1, Unit: &IngredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &IngredientFood{Name: "oil"}},
			want: RenderedIngredient{AmountStr: "1 tbsp", FoodStr: "oil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ing.Render()
			if got.IsHeading != tt.want.IsHeading {
				t.Errorf("IsHeading = %v, want %v", got.IsHeading, tt.want.IsHeading)
			}
			if got.Heading != tt.want.Heading {
				t.Errorf("Heading = %q, want %q", got.Heading, tt.want.Heading)
			}
			if got.AmountStr != tt.want.AmountStr {
				t.Errorf("AmountStr = %q, want %q", got.AmountStr, tt.want.AmountStr)
			}
			if got.FoodStr != tt.want.FoodStr {
				t.Errorf("FoodStr = %q, want %q", got.FoodStr, tt.want.FoodStr)
			}
			if got.NoteStr != tt.want.NoteStr {
				t.Errorf("NoteStr = %q, want %q", got.NoteStr, tt.want.NoteStr)
			}
			if got.Display != tt.want.Display {
				t.Errorf("Display = %q, want %q", got.Display, tt.want.Display)
			}
		})
	}
}

func TestFormatQuantity(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, ""},
		{1, "1"},
		{2, "2"},
		{0.5, "½"},
		{0.25, "¼"},
		{0.75, "¾"},
		{0.333, "⅓"},
		{0.666, "⅔"},
		{0.125, "⅛"},
		{0.375, "⅜"},
		{0.625, "⅝"},
		{0.875, "⅞"},
		{1.5, "1½"},
		{2.25, "2¼"},
		{3.333, "3⅓"},
		{0.15, "0.15"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatQuantity(tt.input)
			if got != tt.want {
				t.Errorf("formatQuantity(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderInlineMarkdown(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"**bold**", "<strong>bold</strong>"},
		{"*italic*", "<em>italic</em>"},
		{"_italic_", "<em>italic</em>"},
		{"[link](http://example.com)", `<a href="http://example.com">link</a>`},
		{"plain text", "plain text"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"**bold** and *italic*", "<strong>bold</strong> and <em>italic</em>"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := string(renderInlineMarkdown(tt.input))
			if got != tt.want {
				t.Errorf("renderInlineMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
