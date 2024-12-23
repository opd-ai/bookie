// Package bookie provides utilities for converting markdown documents into PDF files.
// It supports chapter organization, table of contents generation, and rich text formatting.
// The package uses blackfriday for markdown parsing and gofpdf for PDF generation.
package bookie

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

// Common constants for file extensions and patterns
const (
	jpgExtension  = ".jpg"
	jpegExtension = ".jpeg"
)

// getString extracts all text content from a markdown node by walking its tree.
// It concatenates content from Text nodes while preserving document order and
// ignoring formatting elements.
//
// Parameters:
//   - node: A blackfriday.Node pointer representing the root of the markdown tree.
//     May be nil, in which case an empty string is returned.
//
// Returns:
//   - A string containing the concatenated text from all Text nodes in the tree.
//
// Related: blackfriday.Node, blackfriday.WalkStatus
func getString(node *blackfriday.Node) string {
	if node == nil {
		return ""
	}

	// Pre-allocate builder for better performance
	var result strings.Builder
	result.Grow(64) // Reasonable initial capacity for typical markdown content

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.Text && n.Literal != nil {
			result.Write(n.Literal)
		}
		return blackfriday.GoToNext
	})
	return result.String()
}

// extractEpisodeNumber parses a numerical episode identifier from a file path.
// It looks for paths containing "Episode" followed by digits (e.g., "Episode01").
//
// Parameters:
//   - path: A file or directory path string. May be empty, in which case 0 is returned.
//     The path is expected to contain a base name with the format "EpisodeNN" where
//     NN is a number.
//
// Returns:
//   - An integer representing the episode number, or 0 if:
//   - The path is empty
//   - No episode number is found
//   - The number cannot be parsed
//
// Example paths:
//
//	"Episode01" -> 1
//	"Episode42/content" -> 42
//	"invalid" -> 0
//
// Related: episodeNumberPattern
func extractEpisodeNumber(path string) int {
	if path == "" {
		return 0
	}

	matches := episodeNumberPattern.FindStringSubmatch(filepath.Base(path))
	if len(matches) < 2 {
		return 0
	}

	var num int
	if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
		return 0
	}
	return num
}

// findParent locates the nearest ancestor node with a specified HTML tag.
// The search is case-sensitive and traverses up the DOM tree.
//
// Parameters:
//   - n: The starting HTML node. If nil, returns nil.
//   - tag: The HTML tag name to search for (e.g., "div", "table").
//     Empty tag returns nil.
//
// Returns:
//   - The first ancestor node matching the tag, or nil if:
//   - The input node is nil
//   - The tag is empty
//   - No matching ancestor is found
//
// Related: html.Node, html.ElementNode
func findParent(n *html.Node, tag string) *html.Node {
	if n == nil || tag == "" {
		return nil
	}

	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == tag {
			return p
		}
	}
	return nil
}

// countPreviousSiblings counts HTML element nodes that precede the given node.
// Only considers ElementNode types, ignoring text and comment nodes.
//
// Parameters:
//   - n: The HTML node to count siblings before. If nil, returns 0.
//
// Returns:
//   - The count of ElementNode siblings that come before this node.
//
// Related: html.Node, html.ElementNode
func countPreviousSiblings(n *html.Node) int {
	if n == nil {
		return 0
	}

	count := 0
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.Type == html.ElementNode {
			count++
		}
	}
	return count
}

// getAttr retrieves an attribute value from an HTML node by key.
// Commonly used for extracting href, src, class, and other HTML attributes.
//
// Parameters:
//   - n: The HTML node to examine. If nil, returns empty string.
//   - key: The attribute name to find. If empty, returns empty string.
//
// Returns:
//   - The attribute value as a string, or empty string if:
//   - The node is nil
//   - The key is empty
//   - The attribute is not found
//
// Related: html.Node, html.Attribute
func getAttr(n *html.Node, key string) string {
	if n == nil || key == "" {
		return ""
	}

	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// getTextContent extracts all text content from an HTML node tree.
// Concatenates text from all TextNode descendants in document order.
//
// Parameters:
//   - n: The root HTML node to extract text from. If nil, returns empty string.
//
// Returns:
//   - A string containing all text content from the node tree.
//
// Related: html.Node, html.TextNode
func getTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}

	// Pre-allocate builder for better performance
	var text strings.Builder
	text.Grow(128) // Reasonable initial capacity for typical HTML content

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)
	return text.String()
}

// isJPEGImage checks if a file path has a JPEG image extension.
// The check is case-insensitive and handles both .jpg and .jpeg extensions.
//
// Parameters:
//   - src: The file path to check. If empty, returns false.
//
// Returns:
//   - true if the file path ends with .jpg or .jpeg (case-insensitive)
//   - false if the path is empty or has a different extension
func isJPEGImage(src string) bool {
	if src == "" {
		return false
	}
	src = strings.ToLower(src)
	return strings.HasSuffix(src, jpgExtension) ||
		strings.HasSuffix(src, jpegExtension)
}
