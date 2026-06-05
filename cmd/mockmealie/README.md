# Mockmealie

A standalone mock [Mealie](https://mealie.io) API server for development and CI. Serves 6 sample recipes with embedded images — no real Mealie instance needed.

## Usage

```sh
go run ./cmd/mockmealie          # starts on :9925
go run ./cmd/mockmealie --port 8080
```

Then point static-mealie at it:

```sh
static-mealie --mealie-url http://localhost:9925 --mealie-token mock --out-dir output -v
```

## Endpoints

| Endpoint                                           | Description           |
| -------------------------------------------------- | --------------------- |
| `GET /api/recipes?page=N&perPage=N`                | Paginated recipe list |
| `GET /api/recipes/{slug}`                          | Full recipe JSON      |
| `GET /api/media/recipes/{id}/images/original.webp` | Recipe image          |

Accepts any Bearer token. All requests are logged to stderr.

## Sample Recipes

| Recipe                      | Features                                                     |
| --------------------------- | ------------------------------------------------------------ |
| Classic Spaghetti Bolognese | Ingredient sections, notes, all time fields, orgURL          |
| Chocolate Chip Cookies      | Fractional quantities, abbreviations, step summaries         |
| Greek Salad                 | No cook time, title-less note                                |
| Chicken Tikka Masala        | Inline markdown, instruction section headers, multiple notes |
| Banana Bread                | Ingredients without units, recipeYield as "loaf"             |
| French Onion Soup           | No image (404), recipeServings-only (no recipeYield)         |
