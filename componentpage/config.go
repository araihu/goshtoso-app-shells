// Package componentpage renders the shared structure of a component reference
// entry while leaving controls, previews, examples, and copy consumer-owned.
package componentpage

import (
	"strings"
	"unicode"

	"github.com/a-h/templ"
)

// Example describes one framed component preview and its usage code.
type Example struct {
	ID          string
	Title       string
	Description string
	// PreviewLabel names the rendered state shown above the specimen. Empty
	// values become "Default" for the primary example, the section title for
	// named secondary examples, or "Preview" for an unnamed secondary example.
	PreviewLabel string
	RootAttrs    templ.Attributes
	AbovePreview templ.Component
	Preview      templ.Component
	Code         string
	CodeLabel    string
	Language     string
	MaxHeight    string
}

// Config describes one component reference page.
type Config struct {
	Title       string
	Description string
	RootAttrs   templ.Attributes
	Primary     Example
	Sections    []Example
	After       templ.Component
}

func normalizedExample(pageTitle string, example Example, primary bool) Example {
	if example.ID == "" {
		example.ID = slug(example.Title)
		if primary {
			example.ID = slug(pageTitle)
		}
	}
	if example.Language == "" {
		example.Language = "go"
	}
	if example.CodeLabel == "" {
		if primary {
			example.CodeLabel = "Usage Example"
		} else {
			example.CodeLabel = example.Title
		}
	}
	if example.PreviewLabel == "" {
		if primary {
			example.PreviewLabel = "Default"
		} else if example.Title != "" {
			example.PreviewLabel = example.Title
		} else {
			example.PreviewLabel = "Preview"
		}
	}
	if example.MaxHeight == "" {
		if primary {
			example.MaxHeight = "400px"
		} else {
			example.MaxHeight = "300px"
		}
	}
	return example
}

func slug(value string) string {
	var result strings.Builder
	dash := false
	for _, character := range strings.ToLower(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			result.WriteRune(character)
			dash = false
			continue
		}
		if !dash && result.Len() > 0 {
			result.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(result.String(), "-")
}
