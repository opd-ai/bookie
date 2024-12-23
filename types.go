// Package bookie provides functionality for converting markdown files into PDF documents.
// It offers support for chapters, table of contents generation, and configurable text styling.
// The package is built on top of gofpdf for PDF generation and provides a clean API for
// converting structured markdown content into professionally formatted PDF documents.
package bookie

import "github.com/jung-kurt/gofpdf"

// Default page settings in millimeters (A4)
const (
	DefaultPageWidth  = 210.0
	DefaultPageHeight = 297.0
	DefaultMargin     = 20.0
)

// Font style constants
const (
	StyleBold   = "B"
	StyleItalic = "I"
	StyleNormal = ""
)

// Text alignment constants
const (
	AlignLeft   = "L"
	AlignCenter = "C"
	AlignRight  = "R"
)

// BookCompiler handles the conversion of markdown files into structured PDF documents.
// It provides functionality for organizing content into chapters, generating a table
// of contents, and applying consistent styling throughout the document.
//
// The compiler processes files in a specified directory structure, where each chapter
// is represented by a subdirectory containing markdown files and associated images.
//
// Example usage:
//
//	compiler := bookie.NewBookCompiler("./content", "output.pdf")
//	compiler.SetChapterFont("Arial")
//	err := compiler.Compile()
//
// Related types: Chapter, ToCEntry, TextStyle
type BookCompiler struct {
	// RootDir is the base directory containing chapter subdirectories.
	// Must be a valid, readable directory path.
	RootDir string

	// OutputPath specifies where the generated PDF will be saved.
	// Must be a writable path.
	OutputPath string

	// pdf is the underlying PDF generator instance.
	// Initialized during compilation.
	pdf *gofpdf.Fpdf

	// imageCache tracks processed images to prevent duplicate processing.
	// Keys are image file paths, values indicate processing status.
	imageCache map[string]bool

	// chapterFont specifies the font family used for chapter titles.
	// Must be a valid font name supported by gofpdf.
	chapterFont string

	// textFont specifies the font family used for body text.
	// Must be a valid font name supported by gofpdf.
	textFont string

	// toc holds the table of contents entries in document order.
	toc []ToCEntry

	// pageNumbers controls whether page numbers are rendered.
	pageNumbers bool

	// tocTitle specifies the heading text for the table of contents.
	tocTitle string

	// pageWidth is the PDF page width in millimeters.
	// Defaults to A4 width (210mm).
	pageWidth float64

	// pageHeight is the PDF page height in millimeters.
	// Defaults to A4 height (297mm).
	pageHeight float64

	// margin specifies the page margins in millimeters.
	// Applied to all sides of the page.
	margin float64

	// tocLevels maps heading levels to their display styles.
	// Keys are heading levels (1-6), values are TextStyle configurations.
	tocLevels map[int]TextStyle

	// currentFile tracks the markdown file being processed.
	currentFile string

	// currentChapter tracks the chapter being processed.
	currentChapter interface{}
}

// ToCEntry represents a single entry in the table of contents.
// Each entry corresponds to a heading in the document and provides
// information for generating both the ToC listing and PDF bookmarks.
//
// ToC entries are collected during the first pass of compilation
// and rendered during the second pass to ensure accurate page numbers.
type ToCEntry struct {
	// Title is the text of the heading as it appears in the document
	Title string

	// Level indicates the heading depth (1 = chapter, 2 = section, etc.)
	// Valid values are 1-6, matching HTML heading levels
	Level int

	// PageNum is the PDF page number where this heading appears
	PageNum int

	// Link is the internal PDF identifier for creating clickable navigation
	Link int
}

// Chapter represents a collection of markdown files forming a logical unit.
// Files within a chapter are processed in alphabetical order to maintain
// consistent document structure.
//
// Chapters are typically mapped to directories named "EpisodeNN" where
// NN is a number indicating the chapter order.
//
// Related functions: NewBookCompiler, extractEpisodeNumber
type Chapter struct {
	// Path is the full filesystem path to the chapter directory
	Path string

	// Files contains the sorted list of markdown files in this chapter
	Files []string

	// Images maps image references to their full filesystem paths.
	// Keys are image filenames as referenced in markdown,
	// values are absolute paths to the image files.
	Images map[string]string
}

// TextStyle defines visual formatting attributes for text elements.
// It encapsulates font settings and alignment to ensure consistent
// styling across similar elements in the document.
//
// Styles can be applied to headings, body text, and table of contents entries.
//
// Example usage:
//
//	style := TextStyle{
//	    FontFamily: "Arial",
//	    Style:      "B",
//	    Size:       14,
//	    Alignment:  "L",
//	}
type TextStyle struct {
	// FontFamily specifies the font to use (e.g., "Arial", "Times")
	// Must be a font name supported by gofpdf
	FontFamily string

	// Style indicates font variations:
	// - "" (empty) for normal
	// - "B" for bold
	// - "I" for italic
	// - "BI" for bold italic
	Style string

	// Size is the font size in points
	// Must be greater than 0
	Size float64

	// Alignment controls text justification:
	// - "L" for left
	// - "C" for center
	// - "R" for right
	// - "J" for justified
	Alignment string
}
