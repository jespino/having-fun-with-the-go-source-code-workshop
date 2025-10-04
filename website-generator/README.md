# Website Generator

This Go program generates the workshop website from the markdown exercise files.

## Features

- ✅ Converts markdown exercises to HTML pages
- ✅ Generates index page with exercise overview
- ✅ Includes CSS styling
- ✅ Automatic navigation links (previous/next)
- ✅ Preserves all markdown formatting, emojis, and code blocks
- ✅ Fixes relative links to work in HTML format

## Usage

### Basic Usage

```bash
# From the website-generator directory
go run . -exercises ../exercises -output ../website
```

### Installation

```bash
# Install dependencies
go mod download

# Build the generator
go build -o website-generator

# Run the built binary
./website-generator -exercises ../exercises -output ../website
```

### Command Line Flags

- `-exercises` - Path to the exercises directory (default: `../exercises`)
- `-output` - Path to the output directory (default: `../website`)

### Examples

```bash
# Generate to default output directory
go run .

# Generate to custom location
go run . -output /path/to/output

# Use custom exercises directory
go run . -exercises /path/to/exercises -output /path/to/output
```

## How It Works

1. **Reads Markdown Files**: Scans the exercises directory for all `.md` files
2. **Converts to HTML**: Uses [blackfriday](https://github.com/russross/blackfriday) to convert markdown to HTML
3. **Applies Templates**: Wraps content in HTML templates with navigation and styling
4. **Fixes Links**: Converts relative markdown links to HTML links
5. **Generates Index**: Creates an index page with all exercises listed
6. **Copies CSS**: Includes the CSS stylesheet

## Project Structure

```
website-generator/
├── main.go          # Main program logic
├── templates.go     # HTML and CSS templates
├── go.mod          # Go module definition
└── README.md       # This file
```

## Dependencies

- [blackfriday v2](https://github.com/russross/blackfriday) - Markdown processor

## Generated Output

The generator creates:

- `index.html` - Homepage with exercise overview
- `00-introduction-setup.html` through `10-java-style-stack-traces.html` - Exercise pages
- `style.css` - Stylesheet

## Customization

### Exercise Metadata

Edit the `exerciseMetadata` array in `main.go` to customize:
- Exercise titles
- Emojis
- Descriptions
- Filenames

### Templates

Modify the templates in `templates.go`:
- `exerciseTemplate` - Individual exercise page layout
- `indexTemplate` - Homepage layout
- `cssTemplate` - Styling

### Markdown Processing

The `markdownToHTML()` function can be customized to add:
- Custom markdown extensions
- Post-processing steps
- Link transformations

## Regenerating the Website

After making changes to the markdown files:

```bash
cd website-generator
go run .
```

The website will be regenerated in the `../website` directory.

## Development

### Testing

```bash
# Generate and test locally
go run . -output /tmp/test-website
open /tmp/test-website/index.html
```

### Adding New Exercises

1. Add the markdown file to `../exercises/`
2. Add metadata to `exerciseMetadata` array in `main.go`
3. Run the generator
4. Verify the output

## License

Part of the "Having Fun with the Go Source Code Workshop" project.
