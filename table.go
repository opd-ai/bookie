// Package bookie provides functionality for PDF document generation with support
// for complex table rendering from HTML/markdown source content.
package bookie

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

// Table layout constants define the default dimensions and styling for PDF tables.
// All measurements are in millimeters unless otherwise specified.
const (
	tableWidth      = 170.0 // Total table width (A4 width minus margins)
	tableLineHeight = 6.0   // Height of a single line in table cells
	tableFontSize   = 10.0  // Font size for table content in points

	// Header cell background color (RGB values)
	headerFillR = 240 // Red component
	headerFillG = 240 // Green component
	headerFillB = 240 // Blue component
)

// Table-related errors define common failure conditions during table processing.
var (
	// ErrInvalidTable indicates malformed or unsupported table structure
	ErrInvalidTable = errors.New("invalid table structure")

	// ErrEmptyTable indicates a table without any content
	ErrEmptyTable = errors.New("table has no content")
)

// TableCell represents a single cell within a table, containing both
// content and layout information for PDF rendering.
type TableCell struct {
	content  string  // Text content of the cell
	isHeader bool    // Whether this cell is part of the header row
	width    float64 // Cell width in millimeters
	height   float64 // Cell height in millimeters
}

// TableRow represents a horizontal row of cells within the table,
// tracking both the cells and the maximum height needed for the row.
type TableRow struct {
	cells     []TableCell // Ordered list of cells in this row
	maxHeight float64     // Maximum height needed for any cell in the row
}

// renderTable converts an HTML table node into a formatted PDF table.
// It handles table parsing, layout calculation, and PDF rendering.
//
// Parameters:
//   - n: HTML node representing a <table> element. Must be non-nil and
//     have a valid table structure.
//
// Returns:
//   - error: ErrInvalidTable if the input is nil or not a table node
//   - error: ErrEmptyTable if the table has no content to render
//   - error: Any errors encountered during PDF generation
//
// The table is rendered at the current PDF cursor position with
// the configured styling and dimensions.
func (bc *BookCompiler) renderTable(n *html.Node) error {
	if n == nil || n.Type != html.ElementNode || n.Data != "table" {
		return ErrInvalidTable
	}

	headers, rows, err := bc.parseTableStructure(n)
	if err != nil {
		return err
	}

	colCount := bc.determineColumnCount(headers, rows)
	if colCount == 0 {
		return ErrEmptyTable
	}

	colWidth := tableWidth / float64(colCount)
	return bc.renderTableContent(headers, rows, colWidth)
}

// SplitText splits text into lines that fit within a specified width.
// This is a public utility function used for text wrapping in table cells
// and other contexts where text needs to fit within constraints.
//
// Parameters:
//   - text: The input text to split. May contain multiple words/spaces.
//   - width: Maximum width in millimeters for each line.
//
// Returns:
//   - []string: Array of text lines that fit within the specified width.
//     Returns nil for empty input text.
//
// The function attempts to split on word boundaries when possible,
// only splitting words when they exceed the width constraint.
func (bc *BookCompiler) SplitText(text string, width float64) []string {
	if text == "" {
		return nil
	}

	var lines []string
	words := strings.Split(text, " ")
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if bc.pdf.GetStringWidth(testLine) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				lines = append(lines, word)
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// Internal helper functions below - documented for maintainability

// parseTableStructure extracts headers and data rows from an HTML table node.
// Returns the headers as strings and rows as string arrays.
func (bc *BookCompiler) parseTableStructure(n *html.Node) ([]string, [][]string, error) {
	var headers []string
	var rows [][]string

	for tr := n.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.Data != "tr" {
			continue
		}

		row, isHeader := bc.parseTableRow(tr)
		if isHeader {
			headers = append(headers, row...)
		} else if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	return headers, rows, nil
}

// parseTableRow extracts cell content from a table row node.
// Returns the cell contents and whether this is a header row.
func (bc *BookCompiler) parseTableRow(tr *html.Node) ([]string, bool) {
	var cells []string
	isHeader := false

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type != html.ElementNode || (td.Data != "td" && td.Data != "th") {
			continue
		}

		cellText := getTextContent(td)
		cells = append(cells, cellText)
		isHeader = isHeader || td.Data == "th"
	}

	return cells, isHeader
}

// determineColumnCount calculates the number of columns needed for the table.
// Uses header count if available, otherwise uses the first data row.
func (bc *BookCompiler) determineColumnCount(headers []string, rows [][]string) int {
	if len(headers) > 0 {
		return len(headers)
	}
	if len(rows) > 0 {
		return len(rows[0])
	}
	return 0
}

// renderTableContent handles the PDF generation for the table content.
// Applies appropriate styling and renders headers and data rows.
func (bc *BookCompiler) renderTableContent(headers []string, rows [][]string, colWidth float64) error {
	bc.pdf.SetFont(bc.textFont, "B", tableFontSize)

	if len(headers) > 0 {
		if err := bc.renderTableHeaders(headers, colWidth); err != nil {
			return err
		}
	}

	return bc.renderTableRows(rows, colWidth)
}

// renderTableHeaders renders the table header row with background color.
func (bc *BookCompiler) renderTableHeaders(headers []string, colWidth float64) error {
	bc.pdf.SetFillColor(headerFillR, headerFillG, headerFillB)

	for _, header := range headers {
		x := bc.pdf.GetX()
		y := bc.pdf.GetY()
		bc.pdf.Rect(x, y, colWidth, tableLineHeight, "F")
		bc.pdf.Cell(colWidth, tableLineHeight, header)
	}
	bc.pdf.Ln(tableLineHeight)

	return nil
}

// renderTableRows renders all data rows with appropriate heights.
func (bc *BookCompiler) renderTableRows(rows [][]string, colWidth float64) error {
	bc.pdf.SetFont(bc.textFont, "", tableFontSize)

	for _, row := range rows {
		maxHeight := bc.calculateRowHeight(row, colWidth)
		if err := bc.renderTableRow(row, colWidth, maxHeight); err != nil {
			return err
		}
	}

	return nil
}

// calculateRowHeight determines the maximum height needed for a row.
func (bc *BookCompiler) calculateRowHeight(row []string, colWidth float64) float64 {
	maxHeight := tableLineHeight

	for _, cell := range row {
		lines := bc.SplitText(cell, colWidth)
		height := float64(len(lines)) * tableLineHeight
		if height > maxHeight {
			maxHeight = height
		}
	}

	return maxHeight
}

// renderTableRow renders a single row with specified dimensions.
func (bc *BookCompiler) renderTableRow(row []string, colWidth, rowHeight float64) error {
	y := bc.pdf.GetY()
	x := bc.pdf.GetX()

	for i, cell := range row {
		cellX := x + float64(i)*colWidth
		bc.pdf.Rect(cellX, y, colWidth, rowHeight, "D")
		bc.pdf.MultiCell(colWidth, tableLineHeight, cell, "0", "L", false)
		bc.pdf.SetXY(cellX+colWidth, y)
	}
	bc.pdf.Ln(rowHeight)

	return nil
}
