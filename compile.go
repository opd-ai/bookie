// Package bookcompiler provides functionality for converting markdown files into PDF documents
package bookcompiler

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

// Common errors that can occur during compilation
var (
	ErrEmptyChapter = errors.New("empty chapter with no content files")
	ErrNoBody       = errors.New("HTML document missing body element")
	ErrNilChapter   = errors.New("nil chapter provided")
)

// PDF document constants
const (
	pdfOrientation = "P"
	pdfUnit        = "mm"
	pdfFormat      = "A4"
	pdfMargin      = 20.0

	pageNumFont    = "Arial"
	pageNumStyle   = "I"
	pageNumSize    = 8.0
	pageNumYOffset = -15.0

	chapterTitleFont  = "B"
	chapterTitleSize  = 24.0
	chapterLineHeight = 10.0
	chapterSpacing    = 20.0
)

// Compile generates a PDF document from markdown files organized in chapters.
func (bc *BookCompiler) Compile() error {
	if err := bc.validateCompilerState(); err != nil {
		return fmt.Errorf("invalid compiler state: %w", err)
	}

	// First pass: generate table of contents
	if err := bc.generateTableOfContents(); err != nil {
		return fmt.Errorf("failed to generate table of contents: %w", err)
	}

	// Second pass: generate actual content
	if err := bc.generateContent(); err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	return bc.pdf.OutputFileAndClose(bc.OutputPath)
}

// validateCompilerState checks if the compiler is properly configured
func (bc *BookCompiler) validateCompilerState() error {
	if bc.OutputPath == "" {
		return errors.New("output path not set")
	}
	return nil
}

// generateTableOfContents performs the first pass to collect ToC entries
func (bc *BookCompiler) generateTableOfContents() error {
	bc.initializePDF()

	if err := bc.collectToCEntries(); err != nil {
		return fmt.Errorf("failed to collect ToC entries: %w", err)
	}

	return nil
}

// ensureChapterBreak ensures proper spacing between chapters
func (bc *BookCompiler) ensureChapterBreak() {
	// Always start on a new page
	bc.pdf.AddPage()

	// Add some spacing at the top
	bc.pdf.Ln(20)
}

// generateContent performs the second pass to create the actual PDF content
func (bc *BookCompiler) generateContent() error {
	bc.initializePDF()
	bc.generateToC()

	chapters, err := bc.getChapters()
	if err != nil {
		return fmt.Errorf("failed to get chapters: %w", err)
	}

	// Process each chapter
	for i, chapter := range chapters {
		if err := bc.processChapter(chapter); err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapter.Path, err)
		}

		// If this isn't the last chapter, ensure we end on an even page number
		if i < len(chapters)-1 {
			// If we're on an odd page, add another to make it even
			if bc.pdf.PageNo()%2 != 0 {
				bc.pdf.AddPage()
			}
		}
	}

	return nil
}

// initializePDF creates a new PDF document with standard settings
func (bc *BookCompiler) initializePDF() {
	bc.pdf = gofpdf.New(pdfOrientation, pdfUnit, pdfFormat, "")
	bc.pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)

	if bc.pageNumbers {
		bc.setupPageNumbers()
	}
}

// setupPageNumbers configures the page numbering footer
func (bc *BookCompiler) setupPageNumbers() {
	bc.pdf.SetFooterFunc(func() {
		bc.pdf.SetY(pageNumYOffset)
		bc.pdf.SetFont(pageNumFont, pageNumStyle, pageNumSize)
		bc.pdf.CellFormat(0, chapterLineHeight,
			fmt.Sprintf("Page %d", bc.pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
}

// processChapter handles the conversion of a single chapter to PDF format.
func (bc *BookCompiler) processChapter(chapter Chapter) error {
	if chapter.Path == "" {
		return ErrNilChapter
	}
	if len(chapter.Files) == 0 {
		return ErrEmptyChapter
	}

	// Always start a new chapter on a new page
	bc.pdf.AddPage()

	// Add some spacing at the start of each chapter
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

		// Add spacing between files within the same chapter
		if i < len(chapter.Files)-1 {
			bc.pdf.Ln(defaultLineHeight * 2)
		}
	}

	// Ensure we end with some spacing
	bc.pdf.Ln(defaultLineHeight * 2)

	return nil
}

// renderChapterTitle adds a formatted chapter title to the PDF
func (bc *BookCompiler) renderChapterTitle(chapterPath string) error {
	title := formatChapterTitle(chapterPath)

	// Set font for chapter title
	bc.pdf.SetFont(bc.chapterFont, chapterTitleFont, chapterTitleSize)

	// Center the title
	titleWidth := bc.pdf.GetStringWidth(title)
	pageWidth, _, _ := bc.pdf.PageSize(0)
	x := (pageWidth - titleWidth) / 2

	// Position and write the title
	bc.pdf.SetX(x)
	bc.pdf.Cell(titleWidth, chapterLineHeight, title)

	// Add spacing after title
	bc.pdf.Ln(chapterSpacing)

	return nil
}

// formatChapterTitle creates a consistent chapter title from the path
func formatChapterTitle(path string) string {
	base := filepath.Base(path)
	base = strings.TrimPrefix(base, "Episode")
	return fmt.Sprintf("Episode %s", strings.TrimSpace(base))
}

// processMarkdownFile converts a markdown file to PDF content
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

// convertMarkdownToHTML converts markdown content to HTML
func convertMarkdownToHTML(content []byte) []byte {
	return blackfriday.Run(content,
		blackfriday.WithExtensions(blackfriday.CommonExtensions))
}

// findBodyNode locates the body element in an HTML document
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
