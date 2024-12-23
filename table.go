package bookie

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

// Table rendering constants
const (
	tableWidth      = 170.0 // A4 width minus margins
	tableLineHeight = 6.0
	tableFontSize   = 10.0

	headerFillR = 240
	headerFillG = 240
	headerFillB = 240
)

// Table-related errors
var (
	ErrInvalidTable = errors.New("invalid table structure")
	ErrEmptyTable   = errors.New("table has no content")
)

// TableCell represents a single table cell with its content and metadata
type TableCell struct {
	content  string
	isHeader bool
	width    float64
	height   float64
}

// TableRow represents a row of cells in the table
type TableRow struct {
	cells     []TableCell
	maxHeight float64
}

// renderTable converts an HTML table node into a formatted PDF table.
func (bc *BookCompiler) renderTable(n *html.Node) error {
	if n == nil || n.Type != html.ElementNode || n.Data != "table" {
		return ErrInvalidTable
	}

	// Extract table structure
	headers, rows, err := bc.parseTableStructure(n)
	if err != nil {
		return err
	}

	// Calculate layout
	colCount := bc.determineColumnCount(headers, rows)
	if colCount == 0 {
		return ErrEmptyTable
	}

	colWidth := tableWidth / float64(colCount)

	// Render table
	if err := bc.renderTableContent(headers, rows, colWidth); err != nil {
		return err
	}

	return nil
}

// parseTableStructure extracts headers and rows from the table HTML
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

// parseTableRow extracts cells from a table row
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

// determineColumnCount calculates the number of columns in the table
func (bc *BookCompiler) determineColumnCount(headers []string, rows [][]string) int {
	if len(headers) > 0 {
		return len(headers)
	}
	if len(rows) > 0 {
		return len(rows[0])
	}
	return 0
}

// renderTableContent handles the actual PDF rendering of the table
func (bc *BookCompiler) renderTableContent(headers []string, rows [][]string, colWidth float64) error {
	// Set initial styling
	bc.pdf.SetFont(bc.textFont, "B", tableFontSize)

	// Render headers
	if len(headers) > 0 {
		if err := bc.renderTableHeaders(headers, colWidth); err != nil {
			return err
		}
	}

	// Render data rows
	return bc.renderTableRows(rows, colWidth)
}

// renderTableHeaders renders the header row with background
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

// renderTableRows renders the data rows with appropriate heights
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

// calculateRowHeight determines the maximum height needed for a row
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

// renderTableRow renders a single row with the specified height
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

// SplitText splits text into lines that fit within a specified width.
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
