// Package bookcompiler provides utilities for PDF document generation from markdown files
package bookcompiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

// getString extracts all text content from a markdown node by walking its tree.
//
// Parameters:
//   - node: A blackfriday Node pointer containing markdown content
//
// Returns:
//   - A string containing all concatenated text from the node and its children
//
// The function walks the entire node tree and only collects Text type nodes,
// ignoring formatting and structural elements.
func getString(node *blackfriday.Node) string {
	// Added nil check for safety while maintaining return signature
	if node == nil {
		return ""
	}

	var result strings.Builder
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.Text && n.Literal != nil {
			result.Write(n.Literal)
		}
		return blackfriday.GoToNext
	})
	return result.String()
}

// extractEpisodeNumber parses a numerical episode identifier from a file path.
//
// Parameters:
//   - path: A string containing a file or directory path
//
// Returns:
//   - An integer representing the episode number, or 0 if no valid number is found
//
// The function expects paths containing "Episode" followed by digits (e.g., "Episode01").
// It returns 0 for paths without a valid episode number or with parsing errors.
func extractEpisodeNumber(path string) int {
	// Added empty path check while maintaining return signature
	if path == "" {
		return 0
	}

	matches := episodeNumberPattern.FindStringSubmatch(filepath.Base(path))
	if len(matches) < 2 {
		return 0
	}

	var num int
	// Simplified error handling while maintaining behavior
	if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
		return 0
	}
	return num
}

// findParent searches up the HTML node tree for a parent with the specified tag.
//
// Parameters:
//   - n: The HTML node to start searching from
//   - tag: The tag name to search for (case-sensitive)
//
// Returns:
//   - A pointer to the first matching parent node, or nil if none is found
//
// The search continues up the tree until either a matching parent is found
// or the root is reached.
func findParent(n *html.Node, tag string) *html.Node {
	// Added input validation while maintaining return signature
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

// countPreviousSiblings counts element nodes that precede the given node.
//
// Parameters:
//   - n: The HTML node to count siblings before
//
// Returns:
//   - The number of element-type siblings that come before this node
//
// Only counts nodes of Type ElementNode, ignoring text and comment nodes.
func countPreviousSiblings(n *html.Node) int {
	// Added nil check while maintaining return signature
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

// getAttr retrieves the value of a specified attribute from an HTML node.
//
// Parameters:
//   - n: The HTML node to examine
//   - key: The attribute name to look for
//
// Returns:
//   - The attribute value as a string, or an empty string if not found
//
// Used primarily for extracting href, src, and other common HTML attributes.
func getAttr(n *html.Node, key string) string {
	// Added input validation while maintaining return signature
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

// getTextContent extracts all text content from an HTML node and its children.
//
// Parameters:
//   - n: The HTML node to extract text from
//
// Returns:
//   - A string containing all concatenated text content
//
// Recursively traverses the entire node tree, collecting text from TextNode types
// while preserving the document order.
func getTextContent(n *html.Node) string {
	// Added nil check while maintaining return signature
	if n == nil {
		return ""
	}

	var text strings.Builder
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

// isJPEGImage checks if the given file path ends with a JPEG extension.
//
// Parameters:
//   - src: The file path to check
//
// Returns:
//   - true if the file has a .jpg or .jpeg extension, false otherwise
func isJPEGImage(src string) bool {
	// Added nil check while maintaining return signature
	if src == "" {
		return false
	}
	src = strings.ToLower(src)
	return strings.HasSuffix(src, ".jpg") || strings.HasSuffix(src, ".jpeg")
}
