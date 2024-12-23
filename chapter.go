package bookcompiler

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

// File-related constants
const (
	markdownExt   = ".md"
	episodePrefix = "Episode"
)

// Common errors
var (
	ErrInvalidRoot    = errors.New("invalid root directory")
	ErrNoChapters     = errors.New("no episode chapters found")
	ErrNoMarkdown     = errors.New("no markdown files found")
	ErrInvalidChapter = errors.New("invalid chapter directory")
)

// episodeNumberPattern matches episode numbers in directory names
var episodeNumberPattern = regexp.MustCompile(`Episode\s*(\d+)`)

// getChapters scans the root directory for episode folders and builds a slice of chapters.
func (bc *BookCompiler) getChapters() ([]Chapter, error) {
	// Validate root directory
	if err := bc.validateRootDir(); err != nil {
		return nil, fmt.Errorf("root directory validation failed: %w", err)
	}

	// Collect chapters
	chapters, err := bc.collectChapters()
	if err != nil {
		return nil, err
	}

	// Sort chapters by episode number
	bc.sortChapters(chapters)

	return chapters, nil
}

// validateRootDir ensures the root directory exists and is accessible
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

// collectChapters gathers all valid episode chapters from the root directory
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

// processDirectoryEntry handles a single directory entry and returns a chapter if valid
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

	// Scan for images in the chapter directory
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

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// getMarkdownFiles retrieves all markdown files from a specified directory
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

// collectMarkdownFiles gathers markdown files from directory entries
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

// isMarkdownFile checks if a file entry is a markdown file
func isMarkdownFile(entry fs.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(
		strings.ToLower(entry.Name()),
		markdownExt,
	)
}

// sortChapters sorts chapters by their episode numbers
func (bc *BookCompiler) sortChapters(chapters []Chapter) {
	sort.Slice(chapters, func(i, j int) bool {
		numI := extractEpisodeNumber(chapters[i].Path)
		numJ := extractEpisodeNumber(chapters[j].Path)
		return numI < numJ
	})
}

// logWarning logs a warning message with formatting
func (bc *BookCompiler) logWarning(format string, args ...interface{}) {
	//if bc.Logger != nil {
	log.Printf("WARNING: "+format, args...)
	//}
}

// logDebug logs a debug message with formatting
func (bc *BookCompiler) logDebug(format string, args ...interface{}) {
	//if bc.Logger != nil {
	log.Printf("DEBUG: "+format, args...)
	//}
}
