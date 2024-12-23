package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/opd-ai/bookie"
)

// Configuration defaults
const (
	defaultInDir     = "tmp"
	defaultOutFile   = "tmp.pdf"
	defaultToCTitle  = "Contents"
	defaultLogPrefix = "[BookCompiler] "
)

// Command line flags
var (
	inDir   = flag.String("indir", defaultInDir, "Input directory containing markdown files")
	outFile = flag.String("outfile", defaultOutFile, "Output PDF filename")
	debug   = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%sError: %v", defaultLogPrefix, err)
	}
}

func run() error {
	// Parse and validate flags
	flag.Parse()
	if err := validateFlags(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup logging
	setupLogging()

	// Initialize compiler
	compiler := initializeCompiler()

	// Configure compiler options
	configureCompiler(compiler)

	// Run compilation
	if err := compiler.Compile(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	log.Printf("%sSuccessfully compiled PDF: %s", defaultLogPrefix, *outFile)
	return nil
}

// validateFlags checks command line arguments for validity
func validateFlags() error {
	// Validate input directory
	if *inDir == "" {
		return fmt.Errorf("input directory cannot be empty")
	}

	// Check if input directory exists
	if info, err := os.Stat(*inDir); err != nil {
		return fmt.Errorf("cannot access input directory: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", *inDir)
	}

	// Set default output filename if not specified
	if *outFile == defaultOutFile && *inDir != defaultInDir {
		*outFile = *inDir + ".pdf"
	}

	// Ensure output directory exists
	outDir := filepath.Dir(*outFile)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// setupLogging configures the logging system
func setupLogging() {
	logFlags := log.LstdFlags
	if *debug {
		logFlags |= log.Lshortfile
	}
	log.SetFlags(logFlags)
	log.SetPrefix(defaultLogPrefix)
}

// initializeCompiler creates and returns a new BookCompiler instance
func initializeCompiler() *bookie.BookCompiler {
	compiler := bookie.NewBookCompiler(*inDir, *outFile)

	return compiler
}

// configureCompiler sets up the compiler options
func configureCompiler(compiler *bookie.BookCompiler) {
	compiler.SetToCTitle(defaultToCTitle)
	compiler.SetPageNumbers(true)

	// Additional configuration can be added here
}
