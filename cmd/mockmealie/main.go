package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

//go:embed images/*.webp
var imageFS embed.FS

type paginatedResponse struct {
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
	Items      []recipe `json:"items"`
}

type recipe struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Slug                string        `json:"slug"`
	Image               any           `json:"image"`
	RecipeServings      float64       `json:"recipeServings"`
	RecipeYield         string        `json:"recipeYield"`
	RecipeYieldQuantity float64       `json:"recipeYieldQuantity"`
	TotalTime           string        `json:"totalTime"`
	PrepTime            string        `json:"prepTime"`
	CookTime            string        `json:"cookTime"`
	PerformTime         string        `json:"performTime"`
	Description         string        `json:"description"`
	DateAdded           string        `json:"dateAdded"`
	DateUpdated         string        `json:"dateUpdated"`
	CreatedAt           string        `json:"createdAt"`
	UpdatedAt           string        `json:"updatedAt"`
	RecipeIngredient    []ingredient  `json:"recipeIngredient"`
	RecipeInstructions  []instruction `json:"recipeInstructions"`
	Notes               []note        `json:"notes"`
	OrgURL              string        `json:"orgURL"`
}

type ingredient struct {
	Quantity     float64         `json:"quantity"`
	Unit         *ingredientUnit `json:"unit"`
	Food         *ingredientFood `json:"food"`
	Note         string          `json:"note"`
	Display      string          `json:"display"`
	Title        string          `json:"title"`
	OriginalText string          `json:"originalText"`
}

type ingredientUnit struct {
	Name            string `json:"name"`
	PluralName      string `json:"pluralName"`
	Abbreviation    string `json:"abbreviation"`
	UseAbbreviation bool   `json:"useAbbreviation"`
	Fraction        bool   `json:"fraction"`
}

type ingredientFood struct {
	Name       string `json:"name"`
	PluralName string `json:"pluralName"`
}

type instruction struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Text    string `json:"text"`
}

type note struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

var recipes = []recipe{
	{
		ID:                  "classic-spaghetti-bolognese",
		Name:                "Classic Spaghetti Bolognese",
		Slug:                "classic-spaghetti-bolognese",
		Image:               "bolognese",
		RecipeYield:         "servings",
		RecipeYieldQuantity: 4,
		TotalTime:           "1 hour 15 minutes",
		PrepTime:            "15 minutes",
		CookTime:            "1 hour",
		PerformTime:         "",
		Description:         "A rich and hearty Italian meat sauce served over perfectly cooked spaghetti. This classic recipe uses a slow-simmered combination of beef, tomatoes, and aromatic vegetables.",
		DateAdded:           "2024-01-15",
		DateUpdated:         "2024-06-01",
		CreatedAt:           "2024-01-15T10:30:00Z",
		UpdatedAt:           "2024-06-01T14:00:00Z",
		OrgURL:              "https://example.com/bolognese",
		RecipeIngredient: []ingredient{
			{Title: "For the sauce", Quantity: 1, Unit: &ingredientUnit{Name: "pound", PluralName: "pounds"}, Food: &ingredientFood{Name: "ground beef", PluralName: "ground beef"}, Display: "1 pound ground beef"},
			{Quantity: 1, Unit: nil, Food: &ingredientFood{Name: "onion", PluralName: "onions"}, Note: "diced", Display: "1 onion, diced"},
			{Quantity: 3, Unit: nil, Food: &ingredientFood{Name: "clove garlic", PluralName: "cloves garlic"}, Note: "minced", Display: "3 cloves garlic, minced"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "can", PluralName: "cans"}, Food: &ingredientFood{Name: "crushed tomatoes"}, Note: "28 oz", Display: "1 can (28 oz) crushed tomatoes"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "tablespoon", PluralName: "tablespoons", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "tomato paste"}, Display: "2 tbsp tomato paste"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", PluralName: "teaspoons", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "dried oregano"}, Display: "1 tsp dried oregano"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", PluralName: "teaspoons", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "dried basil"}, Display: "1 tsp dried basil"},
			{Display: "salt and pepper to taste"},
			{Title: "For the pasta", Quantity: 1, Unit: &ingredientUnit{Name: "pound", PluralName: "pounds"}, Food: &ingredientFood{Name: "spaghetti"}, Display: "1 pound spaghetti"},
			{Display: "salted water for boiling"},
		},
		RecipeInstructions: []instruction{
			{ID: "bolognese-1", Title: "", Text: "Heat olive oil in a large saucepan over medium-high heat. Add ground beef and cook, breaking it up, until browned (about 5 minutes). Drain excess fat."},
			{ID: "bolognese-2", Title: "", Text: "Add diced onion and cook until softened, about 3 minutes. Add garlic and cook for 1 minute more."},
			{ID: "bolognese-3", Title: "", Text: "Stir in crushed tomatoes, tomato paste, oregano, and basil. Season with salt and pepper. Bring to a simmer."},
			{ID: "bolognese-4", Title: "", Text: "Reduce heat to low and let the sauce simmer for 45 minutes, stirring occasionally."},
			{ID: "bolognese-5", Title: "", Text: "Meanwhile, bring a large pot of salted water to a boil. Cook spaghetti according to package directions until al dente. Drain."},
			{ID: "bolognese-6", Title: "", Text: "Serve sauce over spaghetti. Garnish with fresh basil and grated Parmesan if desired."},
		},
		Notes: []note{
			{Title: "Make ahead", Text: "The sauce freezes beautifully for up to 3 months. Thaw overnight in the fridge and reheat gently."},
			{Title: "Variation", Text: "For a richer sauce, use a combination of ground beef and ground pork."},
		},
	},
	{
		ID:                  "chocolate-chip-cookies",
		Name:                "Chocolate Chip Cookies",
		Slug:                "chocolate-chip-cookies",
		Image:               "cookies",
		RecipeYield:         "cookies",
		RecipeYieldQuantity: 24,
		TotalTime:           "35 minutes",
		PrepTime:            "15 minutes",
		CookTime:            "10 minutes",
		PerformTime:         "",
		Description:         "Classic homemade chocolate chip cookies that are crispy on the edges and chewy in the center. A timeless favorite for cookie lovers of all ages.",
		DateAdded:           "2024-02-10",
		DateUpdated:         "2024-02-10",
		CreatedAt:           "2024-02-10T09:00:00Z",
		UpdatedAt:           "2024-02-10T09:00:00Z",
		RecipeIngredient: []ingredient{
			{Quantity: 2.25, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "all-purpose flour"}, Display: "2¼ cups all-purpose flour"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "baking soda"}, Display: "1 tsp baking soda"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "salt"}, Display: "1 tsp salt"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "butter"}, Note: "softened", Display: "1 cup butter, softened"},
			{Quantity: 0.75, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "granulated sugar"}, Display: "¾ cup granulated sugar"},
			{Quantity: 0.75, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "packed brown sugar"}, Display: "¾ cup packed brown sugar"},
			{Quantity: 2, Unit: nil, Food: &ingredientFood{Name: "egg", PluralName: "eggs"}, Display: "2 eggs"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "vanilla extract"}, Display: "1 tsp vanilla extract"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "chocolate chips"}, Display: "2 cups chocolate chips"},
		},
		RecipeInstructions: []instruction{
			{ID: "cookies-1", Title: "", Text: "Preheat oven to 375°F (190°C)."},
			{ID: "cookies-2", Title: "", Summary: "Combine dry", Text: "Combine flour, baking soda, and salt in a bowl. Set aside."},
			{ID: "cookies-3", Title: "", Summary: "Cream butter/sugar", Text: "Beat butter, granulated sugar, brown sugar, and vanilla extract in a large mixing bowl until creamy."},
			{ID: "cookies-4", Title: "", Text: "Add eggs to butter mixture, one at a time, beating well after each addition."},
			{ID: "cookies-5", Title: "", Text: "Gradually beat in the flour mixture. Stir in chocolate chips."},
			{ID: "cookies-6", Title: "", Text: "Drop rounded tablespoons of dough onto ungreased baking sheets."},
			{ID: "cookies-7", Title: "", Text: "Bake for 9 to 11 minutes or until golden brown. Cool on baking sheets for 2 minutes, then move to wire racks."},
		},
		Notes: nil,
	},
	{
		ID:                  "greek-salad",
		Name:                "Greek Salad",
		Slug:                "greek-salad",
		Image:               "greek-salad",
		RecipeYield:         "servings",
		RecipeYieldQuantity: 4,
		TotalTime:           "15 minutes",
		PrepTime:            "15 minutes",
		CookTime:            "",
		PerformTime:         "",
		Description:         "A refreshing and colorful Greek salad with crisp vegetables, tangy feta cheese, and a simple olive oil dressing.",
		DateAdded:           "2024-03-05",
		DateUpdated:         "2024-03-05",
		CreatedAt:           "2024-03-05T12:00:00Z",
		UpdatedAt:           "2024-03-05T12:00:00Z",
		RecipeIngredient: []ingredient{
			{Quantity: 1, Unit: nil, Food: &ingredientFood{Name: "English cucumber"}, Note: "chopped", Display: "1 English cucumber, chopped"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "cherry tomatoes"}, Note: "halved", Display: "2 cups cherry tomatoes, halved"},
			{Quantity: 1, Unit: nil, Food: &ingredientFood{Name: "red onion"}, Note: "thinly sliced", Display: "1 red onion, thinly sliced"},
			{Quantity: 0.5, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "Kalamata olives"}, Display: "½ cup Kalamata olives"},
			{Quantity: 4, Unit: &ingredientUnit{Name: "ounce", PluralName: "ounces", Abbreviation: "oz", UseAbbreviation: true}, Food: &ingredientFood{Name: "feta cheese"}, Note: "crumbled", Display: "4 oz feta cheese, crumbled"},
			{Quantity: 3, Unit: &ingredientUnit{Name: "tablespoon", PluralName: "tablespoons", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "extra-virgin olive oil"}, Display: "3 tbsp extra-virgin olive oil"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "red wine vinegar"}, Display: "1 tbsp red wine vinegar"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "dried oregano"}, Display: "1 tsp dried oregano"},
		},
		RecipeInstructions: []instruction{
			{ID: "salad-1", Title: "", Text: "Combine cucumber, tomatoes, red onion, and olives in a large bowl."},
			{ID: "salad-2", Title: "", Text: "Whisk together olive oil, red wine vinegar, and oregano. Season with salt and pepper."},
			{ID: "salad-3", Title: "", Text: "Pour dressing over vegetables and toss gently. Top with crumbled feta cheese. Serve immediately."},
		},
		Notes: []note{
			{Title: "", Text: "For the best flavor, use high-quality extra-virgin olive oil and let the salad sit for 5 minutes before serving so the vegetables absorb the dressing."},
		},
	},
	{
		ID:                  "chicken-tikka-masala",
		Name:                "Chicken Tikka Masala",
		Slug:                "chicken-tikka-masala",
		Image:               "tikka-masala",
		RecipeYield:         "servings",
		RecipeYieldQuantity: 6,
		TotalTime:           "1 hour",
		PrepTime:            "20 minutes",
		CookTime:            "40 minutes",
		PerformTime:         "",
		Description:         "Tender chunks of marinated chicken in a creamy, spiced tomato sauce. Serve with basmati rice or warm naan bread for a satisfying meal.",
		DateAdded:           "2024-04-12",
		DateUpdated:         "2024-05-20",
		CreatedAt:           "2024-04-12T16:00:00Z",
		UpdatedAt:           "2024-05-20T11:30:00Z",
		RecipeIngredient: []ingredient{
			{Quantity: 2, Unit: &ingredientUnit{Name: "pound", PluralName: "pounds"}, Food: &ingredientFood{Name: "boneless chicken thigh", PluralName: "boneless chicken thighs"}, Note: "cut into bite-sized pieces", Display: "2 pounds boneless chicken thighs, cut into bite-sized pieces"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "plain yogurt"}, Display: "1 cup plain yogurt"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "tablespoon", PluralName: "tablespoons", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "garam masala"}, Display: "2 tbsp garam masala"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "turmeric"}, Display: "1 tbsp turmeric"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "cumin"}, Display: "1 tbsp cumin"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "vegetable oil"}, Display: "2 tbsp vegetable oil"},
			{Quantity: 1, Unit: nil, Food: &ingredientFood{Name: "onion", PluralName: "onions"}, Note: "finely diced", Display: "1 onion, finely diced"},
			{Quantity: 4, Unit: nil, Food: &ingredientFood{Name: "clove garlic", PluralName: "cloves garlic"}, Note: "minced", Display: "4 cloves garlic, minced"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "fresh ginger"}, Note: "grated", Display: "1 tbsp fresh ginger, grated"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "can", PluralName: "cans"}, Food: &ingredientFood{Name: "tomato sauce"}, Note: "14 oz", Display: "1 can (14 oz) tomato sauce"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "heavy cream"}, Display: "1 cup heavy cream"},
		},
		RecipeInstructions: []instruction{
			{ID: "tikka-1", Title: "Marinate the chicken", Text: "In a large bowl, combine yogurt, 1 tbsp garam masala, turmeric, and cumin. Add chicken pieces and toss to coat evenly. Cover and refrigerate for **at least 30 minutes**, or up to overnight for best results."},
			{ID: "tikka-2", Title: "", Text: "Preheat your oven's broiler. Thread the marinated chicken onto skewers or arrange on a lined baking sheet."},
			{ID: "tikka-3", Title: "", Text: "Broil for 12-15 minutes, turning once halfway through, until the chicken is *slightly charred* and cooked through. Set aside."},
			{ID: "tikka-4", Title: "Make the sauce", Text: "Heat vegetable oil in a large skillet over medium heat. Add onion and cook until **soft and golden**, about 5 minutes."},
			{ID: "tikka-5", Title: "", Text: "Add garlic and ginger, cook for 1 minute until fragrant. Stir in the remaining 1 tbsp garam masala."},
			{ID: "tikka-6", Title: "", Text: "Pour in the tomato sauce and bring to a simmer. Cook for 15 minutes, stirring occasionally, until the sauce thickens *slightly*."},
			{ID: "tikka-7", Title: "", Text: "Stir in the heavy cream and add the broiled chicken pieces. Simmer for another 10 minutes to let the flavors meld. Adjust seasoning with salt and pepper."},
			{ID: "tikka-8", Title: "", Text: "Serve hot over **basmati rice** or with warm _naan bread_. Garnish with fresh cilantro."},
		},
		Notes: []note{
			{Title: "Spice level", Text: "For extra heat, add 1-2 teaspoons of cayenne pepper to the marinade or sauce."},
			{Title: "Dairy-free option", Text: "Substitute coconut cream for the heavy cream and use coconut yogurt for the marinade."},
			{Title: "Leftovers", Text: "This dish actually tastes better the next day as the flavors continue to develop. Store in the fridge for up to 3 days."},
		},
	},
	{
		ID:                  "banana-bread",
		Name:                "Banana Bread",
		Slug:                "banana-bread",
		Image:               "banana-bread",
		RecipeYield:         "loaf",
		RecipeYieldQuantity: 1,
		TotalTime:           "1 hour 10 minutes",
		PrepTime:            "10 minutes",
		CookTime:            "1 hour",
		PerformTime:         "",
		Description:         "Moist and flavorful banana bread made with overripe bananas. Simple to make and perfect for breakfast or as a snack.",
		DateAdded:           "2024-05-01",
		DateUpdated:         "2024-05-01",
		CreatedAt:           "2024-05-01T08:00:00Z",
		UpdatedAt:           "2024-05-01T08:00:00Z",
		RecipeIngredient: []ingredient{
			{Quantity: 3, Unit: nil, Food: &ingredientFood{Name: "ripe banana", PluralName: "ripe bananas"}, Note: "mashed", Display: "3 ripe bananas, mashed"},
			{Quantity: 0.333, Unit: &ingredientUnit{Name: "cup", PluralName: "cups", Fraction: true}, Food: &ingredientFood{Name: "melted butter"}, Display: "⅓ cup melted butter"},
			{Quantity: 0.75, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "sugar"}, Display: "¾ cup sugar"},
			{Quantity: 1, Unit: nil, Food: &ingredientFood{Name: "egg", PluralName: "eggs"}, Display: "1 egg"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "vanilla extract"}, Display: "1 tsp vanilla extract"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "teaspoon", Abbreviation: "tsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "baking soda"}, Display: "1 tsp baking soda"},
			{Display: "pinch of salt"},
			{Quantity: 1.5, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "all-purpose flour"}, Display: "1½ cups all-purpose flour"},
		},
		RecipeInstructions: []instruction{
			{ID: "banana-1", Title: "", Text: "Preheat oven to 350°F (175°C). Grease a 9x5 inch loaf pan."},
			{ID: "banana-2", Title: "", Text: "Mash the bananas in a large bowl. Stir in melted butter."},
			{ID: "banana-3", Title: "", Text: "Mix in sugar, beaten egg, and vanilla extract."},
			{ID: "banana-4", Title: "", Text: "Sprinkle in baking soda and salt, then mix in flour until just combined. Do not overmix."},
			{ID: "banana-5", Title: "", Text: "Pour batter into prepared loaf pan. Bake for 55-65 minutes, or until a toothpick inserted in the center comes out clean."},
			{ID: "banana-6", Title: "", Text: "Let cool in the pan for 10 minutes, then transfer to a wire rack to cool completely."},
		},
		Notes: nil,
	},
	{
		ID:                  "french-onion-soup",
		Name:                "French Onion Soup",
		Slug:                "french-onion-soup",
		Image:               nil,
		RecipeServings:      4,
		RecipeYield:         "",
		RecipeYieldQuantity: 0,
		TotalTime:           "1 hour 30 minutes",
		PrepTime:            "20 minutes",
		CookTime:            "1 hour 10 minutes",
		PerformTime:         "",
		Description:         "Rich, deeply caramelized onion soup topped with crusty bread and melted Gruyère cheese. A French bistro classic that's surprisingly easy to make at home.",
		DateAdded:           "2024-06-15",
		DateUpdated:         "2024-06-15",
		CreatedAt:           "2024-06-15T18:00:00Z",
		UpdatedAt:           "2024-06-15T18:00:00Z",
		RecipeIngredient: []ingredient{
			{Quantity: 4, Unit: nil, Food: &ingredientFood{Name: "large yellow onion", PluralName: "large yellow onions"}, Note: "thinly sliced", Display: "4 large yellow onions, thinly sliced"},
			{Quantity: 3, Unit: &ingredientUnit{Name: "tablespoon", PluralName: "tablespoons", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "butter"}, Display: "3 tbsp butter"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "olive oil"}, Display: "1 tbsp olive oil"},
			{Quantity: 0.5, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "dry white wine"}, Display: "½ cup dry white wine"},
			{Quantity: 6, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "beef broth"}, Display: "6 cups beef broth"},
			{Quantity: 1, Unit: &ingredientUnit{Name: "tablespoon", Abbreviation: "tbsp", UseAbbreviation: true}, Food: &ingredientFood{Name: "fresh thyme"}, Note: "or 1 tsp dried", Display: "1 tbsp fresh thyme (or 1 tsp dried)"},
			{Quantity: 4, Unit: nil, Food: &ingredientFood{Name: "slice of crusty bread", PluralName: "slices of crusty bread"}, Display: "4 slices of crusty bread"},
			{Quantity: 2, Unit: &ingredientUnit{Name: "cup", PluralName: "cups"}, Food: &ingredientFood{Name: "shredded Gruyère cheese"}, Display: "2 cups shredded Gruyère cheese"},
		},
		RecipeInstructions: []instruction{
			{ID: "soup-1", Title: "", Text: "Melt butter with olive oil in a large Dutch oven over medium heat. Add sliced onions and stir to coat."},
			{ID: "soup-2", Title: "", Text: "Cook onions, stirring occasionally, for 40-50 minutes until deeply caramelized and golden brown. Reduce heat if they start to burn."},
			{ID: "soup-3", Title: "", Text: "Pour in white wine and scrape up any browned bits from the bottom of the pot. Cook until wine has mostly evaporated, about 2 minutes."},
			{ID: "soup-4", Title: "", Text: "Add beef broth and thyme. Bring to a boil, then reduce heat and simmer for 20 minutes. Season with salt and pepper."},
			{ID: "soup-5", Title: "", Text: "Preheat your oven's broiler. Ladle soup into oven-safe bowls. Top each with a slice of bread and a generous amount of Gruyère."},
			{ID: "soup-6", Title: "", Text: "Broil for 2-3 minutes until the cheese is bubbly and golden brown. Serve immediately — the bowls will be very hot."},
		},
		Notes: nil,
	},
}

var recipesBySlug map[string]recipe

func init() {
	recipesBySlug = make(map[string]recipe, len(recipes))
	for _, r := range recipes {
		recipesBySlug[r.Slug] = r
	}
}

func handleRecipesList(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pp := 50
	if perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			pp = p
		}
	}

	total := len(recipes)
	totalPages := int(math.Ceil(float64(total) / float64(pp)))
	start := (page - 1) * pp
	end := start + pp
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	resp := paginatedResponse{
		Page:       page,
		PerPage:    pp,
		Total:      total,
		TotalPages: totalPages,
		Items:      recipes[start:end],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleRecipeGet(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/recipes/")
	slug = strings.TrimSuffix(slug, "/")

	rec, ok := recipesBySlug[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rec)
}

func handleRecipeImage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/media/recipes/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	recipeID := parts[0]

	data, err := imageFS.ReadFile("images/" + recipeID + ".webp")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Write(data)
}

func main() {
	port := flag.Int("port", 9925, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/recipes/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/recipes")
		path = strings.TrimPrefix(path, "/")
		path = strings.TrimSuffix(path, "/")

		if path == "" {
			handleRecipesList(w, r)
		} else {
			handleRecipeGet(w, r)
		}
	})
	mux.HandleFunc("/api/media/recipes/", handleRecipeImage)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		mux.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("mock mealie server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
