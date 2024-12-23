package bookie

import (
	"errors"
	"os"
)

// DirectoryToPDF converts a directory containing markdown files into a PDF byte slice.
// The directory should follow the Bookie chapter structure (Episode01, Episode02, etc.).
//
// Parameters:
//   - dirPath: Path to the directory containing markdown files organized in chapters
//
// Returns:
//   - []byte: The PDF file contents
//   - error: Any error that occurred during processing
//
// Example directory structure:
//
//	inputDir/
//	  ├── Episode01/
//	  │   └── content.md
//	  └── Episode02/
//	      └── content.md
func DirectoryToPDF(dirPath string) ([]byte, error) {
	// Validate input directory
	if dirPath == "" {
		return nil, errors.New("directory path cannot be empty")
	}

	// Create a temporary file for the PDF output
	tmpFile, err := os.CreateTemp("", "bookie-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name()) // Clean up temporary file
	defer tmpFile.Close()

	// Create a new book compiler
	compiler := NewBookCompiler(dirPath, tmpFile.Name())

	// Compile the PDF
	if err := compiler.Compile(); err != nil {
		return nil, err
	}

	// Read the generated PDF file
	pdfBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	return pdfBytes, nil
}

func DirectoryToPDFFile(directoryPath, filePath string) error {
	bytes, err := DirectoryToPDF(directoryPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, bytes, 0644); err != nil {
		return err
	}
	return nil
}
