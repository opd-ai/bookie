// Package bookie provides functionality for compiling markdown files into structured PDF documents.
package bookie

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// File system constants define expected file extensions and naming patterns.
const (
	markdownExt   = ".md"     // Extension for markdown files
	episodePrefix = "Episode" // Directory prefix for chapter folders
)

// Package-level errors define common failure conditions during chapter processing.
var (
	// ErrInvalidRoot indicates the root directory is empty or inaccessible
	ErrInvalidRoot = errors.New("invalid root directory")

	// ErrNoChapters indicates no valid episode chapters were found
	ErrNoChapters = errors.New("no episode chapters found")

	// ErrNoMarkdown indicates a chapter directory contains no markdown files
	ErrNoMarkdown = errors.New("no markdown files found")

	// ErrInvalidChapter indicates a chapter directory is malformed
	ErrInvalidChapter = errors.New("invalid chapter directory")
)

// episodeNumberPattern matches and extracts episode numbers from directory names.
// Example: "Episode 1" -> "1"
var episodeNumberPattern = regexp.MustCompile(`Episode\s*(\d+)`)

// getChapters scans the root directory for episode folders and builds an ordered
// slice of chapters for processing.
//
// Returns:
//   - []Chapter: Ordered slice of chapters found in the root directory
//   - error: Root directory validation or scanning errors
//
// Errors:
//   - ErrInvalidRoot if root directory is invalid
//   - ErrNoChapters if no valid chapters are found
//
// The chapters are sorted by episode number extracted from directory names.
func (bc *BookCompiler) getChapters() ([]Chapter, error) {
	if err := bc.validateRootDir(); err != nil {
		return nil, fmt.Errorf("root directory validation failed: %w", err)
	}

	chapters, err := bc.collectChapters()
	if err != nil {
		return nil, err
	}

	bc.sortChapters(chapters)
	return chapters, nil
}

// validateRootDir ensures the root directory exists and is accessible.
//
// Returns:
//   - error: Validation errors including access and type checks
//
// The root directory must be:
// 1. Non-empty string path
// 2. Existing directory
// 3. Accessible with current permissions
func (bc *BookCompiler) validateRootDir() error {
	if bc.RootDir == "" {
		return ErrInvalidRoot
	}

	info, err := os.Stat(bc.RootDir)
	if err != nil {
		return fmt.Errorf("failed to access root directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", bc.RootDir)
	}

	return nil
}

// collectChapters gathers and validates all episode chapters from the root directory.
//
// Returns:
//   - []Chapter: Slice of valid chapters found
//   - error: Directory reading or validation errors
//
// Each directory entry is processed if it:
// 1. Is a directory
// 2. Contains the episode prefix
// 3. Contains at least one markdown file
func (bc *BookCompiler) collectChapters() ([]Chapter, error) {
	var chapters []Chapter

	entries, err := os.ReadDir(bc.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if chapter, ok := bc.processDirectoryEntry(entry); ok {
			chapters = append(chapters, chapter)
		}
	}

	if len(chapters) == 0 {
		return nil, ErrNoChapters
	}

	return chapters, nil
}

// processDirectoryEntry validates and processes a single directory entry into a Chapter.
//
// Parameters:
//   - entry: Directory entry to process
//
// Returns:
//   - Chapter: Processed chapter if valid
//   - bool: true if entry was processed successfully
//
// Handles image discovery and markdown file collection for each chapter.
func (bc *BookCompiler) processDirectoryEntry(entry fs.DirEntry) (Chapter, bool) {
	if !entry.IsDir() || !strings.Contains(entry.Name(), episodePrefix) {
		return Chapter{}, false
	}

	chapterPath := filepath.Join(bc.RootDir, entry.Name())
	files, err := bc.getMarkdownFiles(chapterPath)
	if err != nil {
		bc.logWarning("Skipping chapter %s: %v", entry.Name(), err)
		return Chapter{}, false
	}

	images := make(map[string]string)
	filepath.Walk(chapterPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && isImageFile(path) {
			images[filepath.Base(path)] = path
		}
		return nil
	})

	return Chapter{
		Path:   chapterPath,
		Files:  files,
		Images: images,
	}, true
}

// isImageFile checks if a file has a supported image extension.
//
// Parameters:
//   - path: File path to check
//
// Returns:
//   - bool: true if file has a supported image extension
//
// Supported extensions: .jpg, .jpeg, .png, .gif
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// getMarkdownFiles retrieves all markdown files from a directory.
//
// Parameters:
//   - path: Directory path to scan
//
// Returns:
//   - []string: Sorted slice of markdown file paths
//   - error: Directory reading errors or if no markdown files found
//
// Files are sorted alphabetically for consistent processing order.
func (bc *BookCompiler) getMarkdownFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read chapter directory: %w", err)
	}

	files := bc.collectMarkdownFiles(entries, path)
	if len(files) == 0 {
		return nil, ErrNoMarkdown
	}

	sort.Strings(files)
	return files, nil
}

// collectMarkdownFiles filters and collects markdown files from directory entries.
//
// Parameters:
//   - entries: Directory entries to process
//   - basePath: Base path for constructing full file paths
//
// Returns:
//   - []string: Slice of full paths to markdown files
func (bc *BookCompiler) collectMarkdownFiles(entries []fs.DirEntry, basePath string) []string {
	var files []string
	for _, entry := range entries {
		if isMarkdownFile(entry) {
			filePath := filepath.Join(basePath, entry.Name())
			files = append(files, filePath)
			bc.logDebug("Found markdown file: %s", entry.Name())
		}
	}
	return files
}

// isMarkdownFile checks if a file entry is a markdown file.
//
// Parameters:
//   - entry: File entry to check
//
// Returns:
//   - bool: true if entry is a non-directory file with .md extension
func isMarkdownFile(entry fs.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(
		strings.ToLower(entry.Name()),
		markdownExt,
	)
}

// sortChapters sorts chapters by their episode numbers in ascending order.
//
// Parameters:
//   - chapters: Slice of chapters to sort in-place
func (bc *BookCompiler) sortChapters(chapters []Chapter) {
	sort.Slice(chapters, func(i, j int) bool {
		numI := extractEpisodeNumber(chapters[i].Path)
		numJ := extractEpisodeNumber(chapters[j].Path)
		return numI < numJ
	})
}

// logWarning logs a warning message with formatting.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for format string
func (bc *BookCompiler) logWarning(format string, args ...interface{}) {
	log.Printf("WARNING: "+format, args...)
}

// logDebug logs a debug message with formatting.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for format string
func (bc *BookCompiler) logDebug(format string, args ...interface{}) {
	log.Printf("DEBUG: "+format, args...)
}
