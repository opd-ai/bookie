// Package bookcompiler provides functionality for converting markdown files into PDF documents
package bookie

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

// renderFormattingElement handles inline text formatting elements.
// Supports emphasis (em/i), strong emphasis (strong/b), and underlining (u).
//
// Parameters:
//   - n: Formatting element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Supported styles:
// - em/i: Italic text
// - strong/b: Bold text
// - u: Underlined text
//
// Note: Formatting is automatically restored to normal after rendering.
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

// renderLink processes hyperlink elements with optional styling.
// Links are rendered in blue to distinguish them from normal text.
//
// Parameters:
//   - n: Anchor element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Features:
// - Blue color for link text
// - Preserves href attribute
// - Restores text color after rendering
// - Handles empty links gracefully
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

// renderImage processes img elements and their associated resources.
// Supports multiple image location strategies and alt text handling.
//
// Parameters:
//   - n: Image element node to render
//
// Returns:
//   - error: Image processing or path resolution errors
//
// Features:
// - Chapter-aware image path resolution
// - Multiple fallback paths for image location
// - Alt text support for accessibility
// - Error handling for missing images
//
// Related: handleImage, Chapter
func (bc *BookCompiler) renderImage(n *html.Node) error {
	src := getAttr(n, "src")
	if src == "" {
		return nil
	}

	imagePath := ""
	// Try chapter-specific image mapping first
	if chapter, ok := bc.currentChapter.(Chapter); ok && chapter.Images != nil {
		if fullPath, exists := chapter.Images[src]; exists {
			imagePath = fullPath
		}
	}

	// Fall back to path resolution if not found in chapter
	if imagePath == "" {
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

// renderBlockquote handles quoted text blocks with distinct styling.
// Applies indentation and italic formatting to quoted content.
//
// Parameters:
//   - n: Blockquote element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Features:
// - Left margin indentation (20mm)
// - Italic text styling
// - Proper spacing before and after
// - Maintains original text alignment
func (bc *BookCompiler) renderBlockquote(n *html.Node) error {
	bc.pdf.SetX(bc.pdf.GetX() + 20)
	bc.pdf.SetFont(bc.textFont, fontStyleItalic, defaultFontSize)
	err := bc.renderChildren(n)
	bc.pdf.SetX(bc.pdf.GetX() - 20)
	bc.pdf.Ln(8)
	return err
}

// renderCode handles preformatted and code block elements.
// Uses monospace font and preserves whitespace formatting.
//
// Parameters:
//   - n: Pre or code element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Features:
// - Courier font for code formatting
// - Preserved whitespace and indentation
// - Consistent spacing around blocks
// - Automatic font restoration
func (bc *BookCompiler) renderCode(n *html.Node) error {
	bc.pdf.SetFont("Courier", fontStyleNormal, 10)
	err := bc.renderChildren(n)
	bc.pdf.SetFont(bc.textFont, fontStyleNormal, defaultFontSize)
	bc.pdf.Ln(8)
	return err
}

// needsSpacing determines if vertical spacing should be added before an element.
// Helps maintain consistent document spacing and readability.
//
// Parameters:
//   - n: HTML node to evaluate
//
// Returns:
//   - bool: true if spacing should be added, false otherwise
//
// Elements requiring spacing:
// - Headings (h1-h3)
// - Paragraphs (p)
// - Lists (ul, ol)
// - Tables
// - Blockquotes
func (bc *BookCompiler) needsSpacing(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	spacingElements := map[string]bool{
		"h1": true, "h2": true, "h3": true,
		"p": true, "ul": true, "ol": true,
		"table": true, "blockquote": true,
	}
	return spacingElements[n.Data]
}

// renderSiblings processes all sibling nodes in sequence.
// Ensures proper rendering order and error propagation.
//
// Parameters:
//   - n: Starting node whose siblings should be rendered
//
// Returns:
//   - error: First error encountered during sibling rendering
//
// Features:
// - Sequential processing of all siblings
// - Proper error context preservation
// - Maintains document flow
//
// Related: renderHTML, renderNode
func (bc *BookCompiler) renderSiblings(n *html.Node) error {
	for c := n.NextSibling; c != nil; c = c.NextSibling {
		if err := bc.renderHTML(c); err != nil {
			return fmt.Errorf("failed to render sibling: %w", err)
		}
	}
	return nil
}
