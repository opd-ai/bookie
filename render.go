// Package bookcompiler provides functionality for converting markdown files into PDF documents
package bookcompiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// Constants for styling to improve maintainability and consistency
const (
	defaultLineHeight = 5.0
	defaultFontSize   = 12.0
	indentWidth       = 10.0
	pageWidth         = 190.0
)

// Font styles for easier maintenance
const (
	fontStyleNormal = ""
	fontStyleBold   = "B"
	fontStyleItalic = "I"
)

// renderNode processes a single HTML node for PDF rendering.
func (bc *BookCompiler) renderNode(n *html.Node) error {
	if n == nil {
		return nil
	}

	// Add spacing check before rendering
	if bc.needsSpacing(n) {
		bc.pdf.Ln(defaultLineHeight)
	}

	return bc.renderHTML(n)
}

// renderChildren processes all direct children of an HTML node.
func (bc *BookCompiler) renderChildren(n *html.Node) error {
	if n == nil {
		return nil
	}

	// Improved error handling with context
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := bc.renderNode(c); err != nil {
			return fmt.Errorf("failed to render child node: %w", err)
		}
	}
	return nil
}

// TextState encapsulates PDF text styling state
type TextState struct {
	FontFamily string
	Style      string
	Size       float64
	Alignment  string
}

// renderHTML converts an HTML node and its siblings to PDF format.
func (bc *BookCompiler) renderHTML(n *html.Node) error {
	// Save current text state for restoration
	currentState := TextState{
		FontFamily: bc.textFont,
		Style:      fontStyleNormal,
		Size:       defaultFontSize,
		Alignment:  "L",
	}
	defer bc.restoreTextState(currentState)

	switch n.Type {
	case html.TextNode:
		return bc.renderTextNode(n)
	case html.ElementNode:
		return bc.renderElement(n)
	}

	// Process siblings
	return bc.renderSiblings(n)
}

// renderTextNode handles text content
func (bc *BookCompiler) renderTextNode(n *html.Node) error {
	text := bc.cleanText(n.Data)
	if strings.TrimSpace(text) != "" {
		bc.pdf.Write(defaultLineHeight, text)
	}
	return nil
}

// renderElement handles HTML elements with appropriate styling
func (bc *BookCompiler) renderElement(n *html.Node) error {
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return bc.renderHeading(n)
	case "p", "blockquote", "pre", "code":
		return bc.renderBlockElement(n)
	case "ul", "ol", "li":
		return bc.renderListElement(n)
	case "em", "i", "strong", "b", "u":
		return bc.renderFormattingElement(n)
	case "table":
		return bc.renderTable(n)
	case "a":
		return bc.renderLink(n)
	case "img":
		return bc.renderImage(n)
	case "hr":
		return bc.renderHorizontalRule()
	}
	return nil
}

func (bc *BookCompiler) getPageHeight() float64 {
	_, height, _ := bc.pdf.PageSize(0)
	return height
}

// renderHeading handles different heading levels
func (bc *BookCompiler) renderHeading(n *html.Node) error {
	// Ensure headings don't start too close to bottom of page
	if bc.pdf.GetY() > bc.getPageHeight()-100 {
		bc.pdf.AddPage()
	}

	switch n.Data {
	case "h1":
		bc.pdf.AddPage()
		bc.setHeadingStyle(24, 20)
	case "h2":
		bc.pdf.Ln(20)
		bc.setHeadingStyle(20, 15)
	case "h3":
		bc.pdf.Ln(15)
		bc.setHeadingStyle(16, 10)
	default: // h4, h5, h6
		bc.pdf.Ln(10)
		bc.setHeadingStyle(14, 8)
	}

	if err := bc.renderChildren(n); err != nil {
		return err
	}
	bc.pdf.Ln(defaultLineHeight * 2)
	return nil
}

// renderBlockElement handles block-level elements
func (bc *BookCompiler) renderBlockElement(n *html.Node) error {
	// Check if we need a page break
	if bc.pdf.GetY() > bc.getPageHeight()-50 {
		bc.pdf.AddPage()
	}

	switch n.Data {
	case "blockquote":
		bc.pdf.Ln(defaultLineHeight)
		err := bc.renderBlockquote(n)
		bc.pdf.Ln(defaultLineHeight)
		return err
	case "pre", "code":
		bc.pdf.Ln(defaultLineHeight)
		err := bc.renderCode(n)
		bc.pdf.Ln(defaultLineHeight)
		return err
	default: // p
		bc.pdf.SetFont(bc.textFont, fontStyleNormal, defaultFontSize)
		bc.pdf.Ln(defaultLineHeight / 2)
		if err := bc.renderChildren(n); err != nil {
			return err
		}
		bc.pdf.Ln(defaultLineHeight)
	}
	return nil
}

// Helper methods for consistent styling and behavior
func (bc *BookCompiler) setHeadingStyle(size float64, spacing float64) {
	bc.pdf.SetFont(bc.chapterFont, fontStyleBold, size)
	bc.pdf.Ln(spacing)
}

func (bc *BookCompiler) restoreTextState(state TextState) {
	bc.pdf.SetFont(state.FontFamily, state.Style, state.Size)
}

// renderHorizontalRule draws a horizontal line
func (bc *BookCompiler) renderHorizontalRule() error {
	x := bc.pdf.GetX()
	y := bc.pdf.GetY()
	bc.pdf.Line(x, y, x+pageWidth, y)
	bc.pdf.Ln(8)
	return nil
}

// handleImage processes and renders an image
func (bc *BookCompiler) handleImage(src, alt string) error {
	if !isJPEGImage(src) {
		return fmt.Errorf("unsupported image format: %s", src)
	}

	// Add spacing before image
	bc.pdf.Ln(defaultLineHeight)

	// Get current position
	x := bc.pdf.GetX()
	y := bc.pdf.GetY()

	// Check if image will fit on current page
	imgInfo := bc.pdf.RegisterImage(src, "")
	if imgInfo == nil {
		return fmt.Errorf("failed to load image: %s", src)
	}

	imgHeight := (imgInfo.Height() * 100) / imgInfo.Width() // scaled to 100mm width
	if y+imgHeight > bc.getPageHeight()-30 {
		bc.pdf.AddPage()
		y = bc.pdf.GetY()
	}

	bc.pdf.Image(src, x, y, 100, 0, false, "", 0, "")
	bc.pdf.SetY(y + imgHeight + 5)

	// Add caption if present
	if alt != "" {
		bc.pdf.SetFont(bc.textFont, fontStyleItalic, 10)
		bc.pdf.Write(defaultLineHeight, alt)
		bc.pdf.Ln(defaultLineHeight)
	}

	bc.pdf.Ln(defaultLineHeight)
	return nil
}

// renderSiblings processes all siblings of the current node
func (bc *BookCompiler) renderSiblings(n *html.Node) error {
	for c := n.NextSibling; c != nil; c = c.NextSibling {
		if err := bc.renderHTML(c); err != nil {
			return fmt.Errorf("failed to render sibling: %w", err)
		}
	}
	return nil
}

// renderListElement handles ordered and unordered lists
func (bc *BookCompiler) renderListElement(n *html.Node) error {
	switch n.Data {
	case "ul", "ol":
		bc.pdf.Ln(5)
		if err := bc.renderChildren(n); err != nil {
			return err
		}
		bc.pdf.Ln(5)
	case "li":
		indent := indentWidth
		if parent := findParent(n, "li"); parent != nil {
			indent += indentWidth // Nested list indentation
		}

		bc.pdf.SetX(bc.pdf.GetX() + indent)
		if parent := findParent(n, "ol"); parent != nil {
			number := countPreviousSiblings(n) + 1
			bc.pdf.Write(defaultLineHeight, fmt.Sprintf("%d. ", number))
		} else {
			bc.pdf.Write(defaultLineHeight, "• ")
		}
		if err := bc.renderChildren(n); err != nil {
			return err
		}
		bc.pdf.Ln(5)
		bc.pdf.SetX(bc.pdf.GetX() - indent)
	}
	return nil
}

// renderFormattingElement handles text formatting elements
func (bc *BookCompiler) renderFormattingElement(n *html.Node) error {
	switch n.Data {
	case "em", "i":
		bc.pdf.SetFont(bc.textFont, fontStyleItalic, defaultFontSize)
		err := bc.renderChildren(n)
		bc.pdf.SetFont(bc.textFont, fontStyleNormal, defaultFontSize)
		return err
	case "strong", "b":
		bc.pdf.SetFont(bc.textFont, fontStyleBold, defaultFontSize)
		err := bc.renderChildren(n)
		bc.pdf.SetFont(bc.textFont, fontStyleNormal, defaultFontSize)
		return err
	case "u":
		x := bc.pdf.GetX()
		y := bc.pdf.GetY()
		if err := bc.renderChildren(n); err != nil {
			return err
		}
		width := bc.pdf.GetStringWidth(getTextContent(n))
		bc.pdf.Line(x, y+3, x+width, y+3)
	}
	return nil
}

// renderLink handles hyperlinks with optional colors
func (bc *BookCompiler) renderLink(n *html.Node) error {
	href := getAttr(n, "href")
	if href != "" {
		bc.pdf.SetTextColor(0, 0, 255) // Blue color for links
		err := bc.renderChildren(n)
		bc.pdf.SetTextColor(0, 0, 0) // Reset to black
		return err
	}
	return bc.renderChildren(n)
}

// renderImage handles image elements with alt text
func (bc *BookCompiler) renderImage(n *html.Node) error {
	src := getAttr(n, "src")
	if src == "" {
		return nil
	}

	// Find the correct image path relative to the markdown file
	imagePath := ""
	if chapter, ok := bc.currentChapter.(Chapter); ok && chapter.Images != nil {
		if fullPath, exists := chapter.Images[src]; exists {
			imagePath = fullPath
		}
	}
	if imagePath == "" {
		// Try different possible paths
		possibilities := []string{
			src,
			filepath.Join(bc.RootDir, src),
			filepath.Join(filepath.Dir(bc.currentFile), src),
		}
		for _, path := range possibilities {
			if _, err := os.Stat(path); err == nil {
				imagePath = path
				break
			}
		}
	}

	if imagePath == "" {
		return fmt.Errorf("image not found: %s", src)
	}

	return bc.handleImage(imagePath, getAttr(n, "alt"))
}

// renderBlockquote handles blockquote elements with indentation
func (bc *BookCompiler) renderBlockquote(n *html.Node) error {
	bc.pdf.SetX(bc.pdf.GetX() + 20)
	bc.pdf.SetFont(bc.textFont, fontStyleItalic, defaultFontSize)
	err := bc.renderChildren(n)
	bc.pdf.SetX(bc.pdf.GetX() - 20)
	bc.pdf.Ln(8)
	return err
}

// renderCode handles pre and code elements
func (bc *BookCompiler) renderCode(n *html.Node) error {
	bc.pdf.SetFont("Courier", fontStyleNormal, 10)
	err := bc.renderChildren(n)
	bc.pdf.SetFont(bc.textFont, fontStyleNormal, defaultFontSize)
	bc.pdf.Ln(8)
	return err
}

// Add helper to determine if spacing is needed
func (bc *BookCompiler) needsSpacing(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	// Elements that need vertical spacing
	spacingElements := map[string]bool{
		"h1": true, "h2": true, "h3": true,
		"p": true, "ul": true, "ol": true,
		"table": true, "blockquote": true,
	}
	return spacingElements[n.Data]
}
