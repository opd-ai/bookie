// Package bookie provides functionality for converting markdown files into PDF documents.
// It supports chapter organization, table of contents generation, and consistent formatting.
package bookie

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

// Package-level errors define common failure conditions during PDF compilation.
var (
	// ErrEmptyChapter indicates a chapter directory contains no content files
	ErrEmptyChapter = errors.New("empty chapter with no content files")

	// ErrNoBody indicates HTML parsing failed to locate the body element
	ErrNoBody = errors.New("HTML document missing body element")

	// ErrNilChapter indicates a nil or invalid chapter was provided
	ErrNilChapter = errors.New("nil chapter provided")
)

// PDF document formatting constants define the layout and styling parameters.
// All measurements are in millimeters unless specified otherwise.
const (
	pdfOrientation = "P"  // Portrait orientation
	pdfUnit        = "mm" // Millimeter measurement unit
	pdfFormat      = "A4" // Standard A4 page size
	pdfMargin      = 20.0 // Page margins

	pageNumFont    = "Arial" // Font for page numbers
	pageNumStyle   = "I"     // Italic style for page numbers
	pageNumSize    = 8.0     // Font size for page numbers
	pageNumYOffset = -15.0   // Vertical offset for page numbers

	chapterTitleFont  = "B"  // Bold style for chapter titles
	chapterTitleSize  = 24.0 // Font size for chapter titles
	chapterLineHeight = 10.0 // Line spacing for chapter titles
	chapterSpacing    = 20.0 // Space after chapter titles
)

// Compile generates a complete PDF document from the organized markdown files.
// It performs two passes:
// 1. Generates table of contents
// 2. Renders actual content with proper page numbers
//
// Returns:
//   - error: Any errors encountered during compilation
//
// The function handles:
// - Compiler state validation
// - Table of contents generation
// - Chapter processing
// - PDF file output
func (bc *BookCompiler) Compile() error {
	if err := bc.validateCompilerState(); err != nil {
		return fmt.Errorf("invalid compiler state: %w", err)
	}

	if err := bc.generateTableOfContents(); err != nil {
		return fmt.Errorf("failed to generate table of contents: %w", err)
	}

	if err := bc.generateContent(); err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	return bc.pdf.OutputFileAndClose(bc.OutputPath)
}

// validateCompilerState ensures all required compiler settings are configured.
//
// Returns:
//   - error: Configuration validation errors
func (bc *BookCompiler) validateCompilerState() error {
	if bc.OutputPath == "" {
		return errors.New("output path not set")
	}
	return nil
}

// generateTableOfContents performs the first pass to collect ToC entries.
// This establishes page numbers for later reference.
//
// Returns:
//   - error: Any errors during ToC generation
func (bc *BookCompiler) generateTableOfContents() error {
	bc.initializePDF()

	if err := bc.collectToCEntries(); err != nil {
		return fmt.Errorf("failed to collect ToC entries: %w", err)
	}

	return nil
}

// ensureChapterBreak adds proper spacing between chapters.
// Always starts a new page and adds vertical spacing.
func (bc *BookCompiler) ensureChapterBreak() {
	bc.pdf.AddPage()
	bc.pdf.Ln(20)
}

// generateContent performs the second pass to create the final PDF content.
// Includes table of contents and all chapters with proper formatting.
//
// Returns:
//   - error: Content generation errors
//
// Ensures chapters start on even pages for proper book layout.
func (bc *BookCompiler) generateContent() error {
	bc.initializePDF()
	bc.generateToC()

	chapters, err := bc.getChapters()
	if err != nil {
		return fmt.Errorf("failed to get chapters: %w", err)
	}

	for i, chapter := range chapters {
		if err := bc.processChapter(chapter); err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapter.Path, err)
		}

		// Ensure chapters start on even pages
		if i < len(chapters)-1 && bc.pdf.PageNo()%2 != 0 {
			bc.pdf.AddPage()
		}
	}

	return nil
}

// initializePDF creates a new PDF document with standard settings.
// Configures page size, margins, and optional page numbering.
func (bc *BookCompiler) initializePDF() {
	bc.pdf = gofpdf.New(pdfOrientation, pdfUnit, pdfFormat, "")
	bc.pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)

	if bc.pageNumbers {
		bc.setupPageNumbers()
	}
}

// setupPageNumbers configures the page numbering footer function.
// Adds centered page numbers at the bottom of each page.
func (bc *BookCompiler) setupPageNumbers() {
	bc.pdf.SetFooterFunc(func() {
		bc.pdf.SetY(pageNumYOffset)
		bc.pdf.SetFont(pageNumFont, pageNumStyle, pageNumSize)
		bc.pdf.CellFormat(0, chapterLineHeight,
			fmt.Sprintf("Page %d", bc.pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
}

// processChapter converts a single chapter's content to PDF format.
//
// Parameters:
//   - chapter: Chapter structure containing content files and metadata
//
// Returns:
//   - error: Chapter processing errors
//
// Handles:
// - Chapter validation
// - Title rendering
// - Content file processing
// - Proper spacing and layout
func (bc *BookCompiler) processChapter(chapter Chapter) error {
	if chapter.Path == "" {
		return ErrNilChapter
	}
	if len(chapter.Files) == 0 {
		return ErrEmptyChapter
	}

	bc.pdf.AddPage()
	bc.pdf.Ln(20)

	if err := bc.renderChapterTitle(chapter.Path); err != nil {
		return fmt.Errorf("failed to render chapter title: %w", err)
	}

	bc.currentChapter = chapter

	for i, file := range chapter.Files {
		bc.currentFile = file
		if err := bc.processMarkdownFile(file); err != nil {
			return fmt.Errorf("failed to process file %s: %w", file, err)
		}

		if i < len(chapter.Files)-1 {
			bc.pdf.Ln(defaultLineHeight * 2)
		}
	}

	bc.pdf.Ln(defaultLineHeight * 2)
	return nil
}

// renderChapterTitle adds a formatted chapter title to the PDF.
//
// Parameters:
//   - chapterPath: Path containing the chapter name to format
//
// Returns:
//   - error: Any rendering errors encountered
//
// Features:
// - Centered title placement
// - Consistent font styling
// - Proper vertical spacing
// - Episode number extraction
func (bc *BookCompiler) renderChapterTitle(chapterPath string) error {
	title := formatChapterTitle(chapterPath)

	bc.pdf.SetFont(bc.chapterFont, chapterTitleFont, chapterTitleSize)

	// Center title horizontally
	titleWidth := bc.pdf.GetStringWidth(title)
	pageWidth, _, _ := bc.pdf.PageSize(0)
	x := (pageWidth - titleWidth) / 2

	bc.pdf.SetX(x)
	bc.pdf.Cell(titleWidth, chapterLineHeight, title)
	bc.pdf.Ln(chapterSpacing)

	return nil
}

// formatChapterTitle creates a consistent chapter title from the path.
//
// Parameters:
//   - path: Full path to chapter directory
//
// Returns:
//   - string: Formatted title string (e.g., "Episode 1")
//
// Handles:
// - Directory name extraction
// - Prefix removal
// - Consistent formatting
func formatChapterTitle(path string) string {
	base := filepath.Base(path)
	base = strings.TrimPrefix(base, "Episode")
	return fmt.Sprintf("Episode %s", strings.TrimSpace(base))
}

// processMarkdownFile converts a single markdown file to PDF content.
//
// Parameters:
//   - filePath: Path to markdown file
//
// Returns:
//   - error: File processing errors
//
// Process:
// 1. Read markdown file
// 2. Convert to HTML
// 3. Parse HTML structure
// 4. Render content
//
// Errors:
// - File reading errors
// - HTML parsing errors
// - Missing body element
// - Rendering errors
func (bc *BookCompiler) processMarkdownFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	htmlContent := convertMarkdownToHTML(content)

	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	body := findBodyNode(doc)
	if body == nil {
		return ErrNoBody
	}

	if err := bc.renderChildren(body); err != nil {
		return fmt.Errorf("failed to render content: %w", err)
	}

	return nil
}

// convertMarkdownToHTML transforms markdown content to HTML format.
//
// Parameters:
//   - content: Raw markdown bytes
//
// Returns:
//   - []byte: HTML content bytes
//
// Features:
// - Common markdown extensions enabled
// - GitHub-flavored markdown support
// - Preserves formatting and structure
//
// Uses blackfriday markdown parser with standard extensions.
func convertMarkdownToHTML(content []byte) []byte {
	return blackfriday.Run(content,
		blackfriday.WithExtensions(blackfriday.CommonExtensions))
}

// findBodyNode locates the body element in an HTML document.
//
// Parameters:
//   - doc: Root HTML node
//
// Returns:
//   - *html.Node: Body node if found, nil otherwise
//
// Implementation:
// - Recursive depth-first search
// - Returns first body element found
// - Handles malformed HTML gracefully
//
// Related: html.Node, processMarkdownFile
func findBodyNode(doc *html.Node) *html.Node {
	var body *html.Node
	var findBody func(*html.Node)

	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}

	findBody(doc)
	return body
}
