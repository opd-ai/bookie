# Bookie

Bookie is a powerful Go library for converting markdown documents into professionally formatted PDF books with support for chapters, table of contents, and rich formatting. It's designed to help authors and technical writers create publication-ready documents from markdown content.

## Features

- **Chapter Organization**
  - Automatic chapter discovery and numbering
  - Support for episode-based content structure
  - Flexible markdown file organization

- **Rich Content Support**
  - Full markdown syntax support including tables
  - Image handling with automatic scaling
  - Code blocks with syntax highlighting
  - Nested lists (ordered and unordered)
  - Blockquotes and horizontal rules

- **Professional PDF Output**
  - Automatic table of contents generation
  - Configurable page numbering
  - Consistent typography and spacing
  - A4 page format with customizable margins
  - Header and footer support

- **Advanced Formatting**
  - Custom font styles and sizes
  - Table support with header styling
  - Flexible text alignment options
  - Link highlighting
  - Image captions

## Installation

```bash
go get github.com/opd-ai/bookie
```

Requires Go 1.21.3 or later.

## Usage

### Basic Example

```go
package main

import "github.com/opd-ai/bookie"

func main() {
    // Create a new book compiler
    compiler := bookie.NewBookCompiler(
        "path/to/markdown/files",  // Root directory containing chapters
        "output.pdf",             // Output PDF path
    )

    // Compile the book
    if err := compiler.Compile(); err != nil {
        log.Fatal(err)
    }
}
```

### Directory Structure

Organize your markdown files in episode-based chapters:

```
root/
  ├── Episode01/
  │   ├── content.md
  │   └── images/
  ├── Episode02/
  │   ├── intro.md
  │   └── details.md
  └── Episode03/
      └── final.md
```

## Configuration

Configure the book compiler with these options:

```go
compiler.SetPageNumbers(true)           // Enable/disable page numbers
compiler.SetToCTitle("Table of Contents") // Custom ToC title

// Font settings are configurable
compiler.SetChapterFont("Arial")
compiler.SetTextFont("Times")
```

### Default Settings

- Page Size: A4 (210x297mm)
- Margins: 20mm
- Chapter Font: Arial
- Body Text Font: Times
- Default Line Height: 5mm

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./...`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Code Style

- Follow standard Go formatting guidelines
- Include comprehensive comments
- Add tests for new functionality
- Update documentation when needed

## Testing

Run the test suite:

```bash
go test -v ./...
```

Ensure all tests pass before submitting pull requests.

## License

This project is licensed under the MIT License - see the `LICENSE` file for details.

## Acknowledgments

Built with these excellent libraries:
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation
- [blackfriday/v2](https://github.com/russross/blackfriday) - Markdown processing
- [golang.org/x/net](https://golang.org/x/net) - HTML processing

## Documentation

For detailed API documentation, visit the [Go package documentation](https://pkg.go.dev/github.com/opd-ai/bookie).

For more examples and detailed usage, see the [Wiki](https://github.com/opd-ai/bookie/wiki).

## Support the Project

If you find this project useful, consider supporting the developer:

Monero Address: `43H3Uqnc9rfEsJjUXZYmam45MbtWmREFSANAWY5hijY4aht8cqYaT2BCNhfBhua5XwNdx9Tb6BEdt4tjUHJDwNW5H7mTiwe`
Bitcoin Address: `bc1qew5kx0srtp8c4hlpw8ax0gllhnpsnp9ylthpas`