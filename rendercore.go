// Package bookie provides functionality for converting markdown documents into PDF files.
// It supports rich text formatting, image handling, tables, and hierarchical document structure.
package bookie

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Layout constants define dimensions and spacing for PDF elements.
// All measurements are in millimeters unless specified otherwise.
const (
	defaultLineHeight = 5.0   // Vertical spacing between lines
	defaultFontSize   = 12.0  // Base font size in points
	indentWidth       = 10.0  // List and blockquote indentation
	pageWidth         = 190.0 // Available content width (A4 minus margins)
)

// Font style constants define standard text formatting options.
// These match the gofpdf style string requirements.
const (
	fontStyleNormal = ""  // Regular weight
	fontStyleBold   = "B" // Bold weight
	fontStyleItalic = "I" // Italic style
)

// TextState encapsulates the PDF text styling configuration.
// It's used to save and restore text rendering state during node processing.
type TextState struct {
	// FontFamily specifies the font name (e.g., "Arial", "Times")
	FontFamily string

	// Style contains font variations ("", "B", "I", "BI")
	Style string

	// Size specifies the font size in points
	Size float64

	// Alignment controls text justification ("L", "C", "R", "J")
	Alignment string
}

// renderNode processes a single HTML node and converts it to PDF format.
// It handles spacing between elements and delegates to specific renderers
// based on node type.
//
// Parameters:
//   - n: HTML node to process. May be nil, in which case no action is taken.
//
// Returns:
//   - error: Any rendering errors encountered
//
// Related: renderHTML, needsSpacing
func (bc *BookCompiler) renderNode(n *html.Node) error {
	if n == nil {
		return nil
	}

	if bc.needsSpacing(n) {
		bc.pdf.Ln(defaultLineHeight)
	}

	return bc.renderHTML(n)
}

// renderChildren processes all direct child nodes of the given HTML node.
// It maintains document structure and handles error propagation.
//
// Parameters:
//   - n: Parent HTML node. If nil, returns without error.
//
// Returns:
//   - error: First error encountered during child rendering, with context
//
// Related: renderNode
func (bc *BookCompiler) renderChildren(n *html.Node) error {
	if n == nil {
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := bc.renderNode(c); err != nil {
			return fmt.Errorf("failed to render child node: %w", err)
		}
	}
	return nil
}

// renderHTML converts an HTML node tree to PDF format.
// It preserves text styling state and handles different node types.
//
// Parameters:
//   - n: Root HTML node to process
//
// Returns:
//   - error: Any rendering errors encountered
//
// The function saves and restores text styling to ensure consistent
// formatting across the document.
func (bc *BookCompiler) renderHTML(n *html.Node) error {
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

	return bc.renderSiblings(n)
}

// renderTextNode processes text content for PDF output.
// It handles text cleaning and writes content to the PDF.
//
// Parameters:
//   - n: Text node containing content to render
//
// Returns:
//   - error: Any writing errors encountered
//
// Empty or whitespace-only text is skipped.
func (bc *BookCompiler) renderTextNode(n *html.Node) error {
	text := bc.cleanText(n.Data)
	if strings.TrimSpace(text) != "" {
		bc.pdf.Write(defaultLineHeight, text)
	}
	return nil
}

// renderElement dispatches HTML elements to appropriate handlers.
// It supports headings, block elements, lists, formatting, tables,
// links, images, and horizontal rules.
//
// Parameters:
//   - n: Element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Elements without specific handlers are ignored.
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
