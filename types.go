// Package bookcompiler provides functionality for converting markdown files into PDF documents
// with support for chapters, table of contents, and styled text formatting.
package bookie

import "github.com/jung-kurt/gofpdf"

// BookCompiler handles the compilation of markdown files into a structured PDF document.
// It supports chapter organization, table of contents generation, and configurable styling.
type BookCompiler struct {
	// RootDir is the base directory containing chapter subdirectories
	RootDir string

	// OutputPath is the target path for the generated PDF file
	OutputPath string

	// pdf is the underlying PDF generator instance
	pdf *gofpdf.Fpdf

	// imageCache tracks processed images to avoid duplicates
	imageCache map[string]bool

	// chapterFont specifies the font family used for chapter titles
	chapterFont string

	// textFont specifies the font family used for body text
	textFont string

	// toc holds the table of contents entries
	toc []ToCEntry

	// pageNumbers enables/disables page number rendering
	pageNumbers bool

	// tocTitle specifies the heading for the table of contents
	tocTitle string

	// pageWidth is the PDF page width in millimeters (default: A4)
	pageWidth float64

	// pageHeight is the PDF page height in millimeters (default: A4)
	pageHeight float64

	// margin specifies the page margins in millimeters
	margin float64

	// tocLevels maps heading levels to their display styles in the table of contents
	tocLevels map[int]TextStyle

	currentFile    string
	currentChapter interface{}
}

// ToCEntry represents a single table of contents entry with its metadata.
// Each entry corresponds to a section heading in the document.
type ToCEntry struct {
	// Title is the text of the heading
	Title string

	// Level indicates the heading depth (1 = chapter, 2 = section, etc.)
	Level int

	// PageNum is the PDF page number where this heading appears
	PageNum int

	// Link is the internal PDF identifier for creating clickable links
	Link int
}

// Chapter represents a collection of markdown files that form a logical chapter.
// Files within a chapter are processed in alphabetical order.
type Chapter struct {
	Path   string            // Full path to the chapter directory
	Files  []string          // Sorted list of markdown files in the chapter
	Images map[string]string // Map of image references to their full paths
}

// TextStyle defines the visual formatting for a text element.
// This is used for consistent styling across different document sections.
type TextStyle struct {
	// FontFamily specifies the font to use (e.g., "Arial", "Times")
	FontFamily string

	// Style indicates font variations ("B" = bold, "I" = italic, etc.)
	Style string

	// Size is the font size in points
	Size float64

	// Alignment controls text justification ("L" = left, "C" = center, "R" = right)
	Alignment string
}
