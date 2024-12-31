package bookie

import (
	"fmt"

	"golang.org/x/net/html"
)

// getPageHeight returns the current PDF page height in millimeters.
// Used for pagination and layout calculations.
//
// Returns:
//   - float64: Page height in millimeters
//
// Related: pdf.PageSize
func (bc *BookCompiler) getPageHeight() float64 {
	_, height, _ := bc.pdf.PageSize(0)
	return height
}

// renderHeading handles heading elements (h1-h6) with appropriate styling.
// It manages page breaks and spacing for different heading levels.
//
// Parameters:
//   - n: Heading element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Heading levels affect font size, spacing, and page breaks:
// - h1: New page, 24pt
// - h2: 20pt with extra spacing
// - h3: 16pt with moderate spacing
// - h4-h6: 14pt with minimal spacing
func (bc *BookCompiler) renderHeading(n *html.Node) error {
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

// renderBlockElement processes block-level HTML elements.
// Handles paragraphs, blockquotes, and code blocks with appropriate
// styling and spacing.
//
// Parameters:
//   - n: Block element node to render
//
// Returns:
//   - error: Any rendering errors encountered
//
// Manages page breaks and applies element-specific formatting.
func (bc *BookCompiler) renderBlockElement(n *html.Node) error {
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

// setHeadingStyle applies consistent formatting for headings.
//
// Parameters:
//   - size: Font size in points
//   - spacing: Vertical spacing in millimeters
func (bc *BookCompiler) setHeadingStyle(size, spacing float64) {
	bc.pdf.SetFont(bc.chapterFont, fontStyleBold, size)
	bc.pdf.Ln(spacing)
}

// restoreTextState restores previously saved text formatting.
//
// Parameters:
//   - state: TextState containing saved formatting options
func (bc *BookCompiler) restoreTextState(state TextState) {
	bc.pdf.SetFont(state.FontFamily, state.Style, state.Size)
}

// renderHorizontalRule draws a horizontal line across the page width.
// Adds vertical spacing after the line.
//
// Returns:
//   - error: Any drawing errors encountered
func (bc *BookCompiler) renderHorizontalRule() error {
	x := bc.pdf.GetX()
	y := bc.pdf.GetY()
	bc.pdf.Line(x, y, x+pageWidth, y)
	bc.pdf.Ln(8)
	return nil
}

// handleImage processes and renders a JPEG image with optional caption.
// Handles image scaling, page breaks, and positioning.
//
// Parameters:
//   - src: Image file path
//   - alt: Optional caption text
//
// Returns:
//   - error: Image processing or rendering errors
//
// Supports only JPEG images and automatically scales them to fit the page width.
func (bc *BookCompiler) handleImage(src, alt string) error {
	if !isJPEGImage(src) {
		return fmt.Errorf("unsupported image format: %s", src)
	}

	bc.pdf.Ln(defaultLineHeight)
	x := bc.pdf.GetX()
	y := bc.pdf.GetY()

	imgInfo := bc.pdf.RegisterImage(src, "")
	if imgInfo == nil {
		return fmt.Errorf("failed to load image: %s", src)
	}

	imgHeight := (imgInfo.Height() * 100) / imgInfo.Width()
	if y+imgHeight > bc.getPageHeight()-30 {
		bc.pdf.AddPage()
		y = bc.pdf.GetY()
	}

	bc.pdf.Image(src, x, y, 100, 0, false, "", 0, "")
	bc.pdf.SetY(y + imgHeight + 5)

	if alt != "" {
		bc.pdf.SetFont(bc.textFont, fontStyleItalic, 10)
		bc.pdf.Write(defaultLineHeight, alt)
		bc.pdf.Ln(defaultLineHeight)
	}

	bc.pdf.Ln(defaultLineHeight)
	return nil
}

// renderListElement handles ordered and unordered lists.
// Supports nested lists with proper indentation.
//
// Parameters:
//   - n: List or list item node
//
// Returns:
//   - error: Any rendering errors encountered
//
// Features:
// - Automatic numbering for ordered lists
// - Bullet points for unordered lists
// - Nested list indentation
// - Proper spacing between items
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
			indent += indentWidth
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
