package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

type Exercise struct {
	Number      int
	Title       string
	Emoji       string
	Description string
	Filename    string
	Content     template.HTML
	PrevLink    string
	NextLink    string
}

type IndexData struct {
	Exercises []Exercise
}

var exerciseMetadata = []struct {
	Filename    string
	Title       string
	Emoji       string
	Description string
}{
	{"00-introduction-setup.md", "Introduction and Setup", "🌱", "Get started by cloning and setting up the Go source code environment."},
	{"01-compile-go-unchanged.md", "Compiling Go Without Changes", "🔨", "Learn to build the Go toolchain from source without any modifications."},
	{"02-scanner-arrow-operator.md", "Adding the \"=>\" Arrow Operator for Goroutines", "⚡", "Learn scanner/lexer modification by adding \"=>\" as an alternative syntax for starting goroutines."},
	{"03-parser-multiple-go.md", "Multiple \"go\" Keywords - Parser Enhancement", "🔄", "Learn parser modification by enabling multiple consecutive \"go\" keywords (go go go myFunction)."},
	{"04-compiler-inlining-parameters.md", "Inline Parameters - Function Inlining Experiments", "⚙️", "Explore the inliner behavior by modifying function inlining parameters."},
	{"05-gofmt-ast-transformation.md", "gofmt Transformation - \"hello\" to \"helo\"", "🎨", "Learn about Go's tools by modifying gofmt to modify \"hello\" to \"helo\" in code."},
	{"06-ssa-power-of-two-detector.md", "SSA Pass - Detecting Division by Powers of Two", "🔍", "Create a custom SSA compiler pass that detects division operations by powers of two that could be optimized to bit shifts."},
	{"07-runtime-patient-go.md", "Patient Go - Making Go Wait for Goroutines", "🕰️", "Modify the Go runtime to wait for all goroutines to complete before program termination."},
	{"08-goroutine-sleep-detective.md", "Goroutine Sleep Detective - Runtime State Monitoring", "🕵️‍♂️", "Add logging to the Go scheduler to monitor goroutines going to sleep."},
	{"09-predictable-select.md", "Predictable Select - Removing Randomness from Go's Select Statement", "🎯", "Modify Go's select statement implementation to be deterministic instead of random."},
	{"10-java-style-stack-traces.md", "Java-Style Stack Traces - Making Go Panics Look Familiar", "☕", "Transform Go's verbose stack traces into Java-style formatting."},
}

func main() {
	exercisesDir := flag.String("exercises", "../exercises", "Path to exercises directory")
	outputDir := flag.String("output", "../website", "Path to output directory")
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate exercise pages
	exercises := make([]Exercise, 0, len(exerciseMetadata))
	for i, meta := range exerciseMetadata {
		exercise, err := generateExercisePage(*exercisesDir, *outputDir, meta, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating exercise %s: %v\n", meta.Filename, err)
			os.Exit(1)
		}
		exercises = append(exercises, exercise)
	}

	// Generate index page
	if err := generateIndexPage(*outputDir, exercises); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating index page: %v\n", err)
		os.Exit(1)
	}

	// Copy CSS file
	if err := copyCSSFile(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying CSS file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Website generated successfully!")
	fmt.Printf("📁 Output directory: %s\n", *outputDir)
	fmt.Printf("📄 Generated %d exercise pages + index page\n", len(exercises))
}

func generateExercisePage(exercisesDir, outputDir string, meta struct {
	Filename    string
	Title       string
	Emoji       string
	Description string
}, index int) (Exercise, error) {
	// Read markdown file
	mdPath := filepath.Join(exercisesDir, meta.Filename)
	content, err := os.ReadFile(mdPath)
	if err != nil {
		return Exercise{}, fmt.Errorf("reading markdown file: %w", err)
	}

	// Convert markdown to HTML
	htmlContent := markdownToHTML(content)

	// Generate HTML filename
	htmlFilename := strings.TrimSuffix(meta.Filename, ".md") + ".html"

	// Determine prev/next links
	prevLink := "index.html"
	if index > 0 {
		prevLink = strings.TrimSuffix(exerciseMetadata[index-1].Filename, ".md") + ".html"
	}

	nextLink := ""
	if index < len(exerciseMetadata)-1 {
		nextLink = strings.TrimSuffix(exerciseMetadata[index+1].Filename, ".md") + ".html"
	}

	exercise := Exercise{
		Number:      index,
		Title:       meta.Title,
		Emoji:       meta.Emoji,
		Description: meta.Description,
		Filename:    htmlFilename,
		Content:     template.HTML(htmlContent),
		PrevLink:    prevLink,
		NextLink:    nextLink,
	}

	// Generate HTML page
	tmpl, err := template.New("exercise").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).Parse(exerciseTemplate)
	if err != nil {
		return Exercise{}, fmt.Errorf("parsing template: %w", err)
	}

	outputPath := filepath.Join(outputDir, htmlFilename)
	f, err := os.Create(outputPath)
	if err != nil {
		return Exercise{}, fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, exercise); err != nil {
		return Exercise{}, fmt.Errorf("executing template: %w", err)
	}

	fmt.Printf("✓ Generated %s\n", htmlFilename)
	return exercise, nil
}

func generateIndexPage(outputDir string, exercises []Exercise) error {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	outputPath := filepath.Join(outputDir, "index.html")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	data := IndexData{Exercises: exercises}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	fmt.Printf("✓ Generated index.html\n")
	return nil
}

func copyCSSFile(outputDir string) error {
	cssContent := cssTemplate
	outputPath := filepath.Join(outputDir, "style.css")

	if err := os.WriteFile(outputPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("writing CSS file: %w", err)
	}

	fmt.Printf("✓ Generated style.css\n")
	return nil
}

func markdownToHTML(markdown []byte) string {
	// Use blackfriday to convert markdown to HTML
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})

	// Process the markdown
	html := blackfriday.Run(markdown, blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(blackfriday.CommonExtensions))

	// Post-process to fix relative links
	htmlStr := string(html)
	htmlStr = fixRelativeLinks(htmlStr)

	return htmlStr
}

func fixRelativeLinks(html string) string {
	// Convert markdown links to HTML links
	re := regexp.MustCompile(`href="\.\./(README\.md|exercises/([^"]+)\.md)"`)
	html = re.ReplaceAllStringFunc(html, func(match string) string {
		if strings.Contains(match, "README.md") {
			return `href="index.html"`
		}
		// Extract filename from exercises/XX-name.md
		re2 := regexp.MustCompile(`exercises/([^"]+)\.md`)
		matches := re2.FindStringSubmatch(match)
		if len(matches) > 1 {
			return fmt.Sprintf(`href="%s.html"`, matches[1])
		}
		return match
	})

	// Fix links that are already in the format XX-name.md
	re = regexp.MustCompile(`href="([0-9]{2}-[^"]+)\.md"`)
	html = re.ReplaceAllString(html, `href="$1.html"`)

	return html
}

const exerciseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exercise {{.Number}}: {{.Title}} - Go Source Code Workshop</title>
    <link rel="stylesheet" href="style.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/go.min.js"></script>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            hljs.highlightAll();

            // Add copy buttons to all code blocks
            document.querySelectorAll('pre').forEach(function(pre) {
                const button = document.createElement('button');
                button.className = 'copy-button';
                button.innerHTML = '<i class="far fa-copy"></i>';
                button.title = 'Copy to clipboard';

                button.addEventListener('click', function() {
                    const code = pre.querySelector('code');
                    const text = code.textContent;

                    navigator.clipboard.writeText(text).then(function() {
                        button.innerHTML = '<i class="fas fa-check"></i>';
                        button.classList.add('copied');
                        setTimeout(function() {
                            button.innerHTML = '<i class="far fa-copy"></i>';
                            button.classList.remove('copied');
                        }, 2000);
                    }).catch(function(err) {
                        console.error('Failed to copy:', err);
                    });
                });

                pre.appendChild(button);
            });
        });
    </script>
</head>
<body>
    <nav class="navbar">
        <div class="container">
            <a href="index.html" class="nav-home">🚀 Go Source Code Workshop</a>
            <div class="nav-links">
                <a href="index.html">Home</a>
            </div>
        </div>
    </nav>

    <div class="container">
        <article class="exercise-content">
            {{.Content}}
        </article>

        <nav class="exercise-nav">
            {{if .PrevLink}}
            <a href="{{.PrevLink}}" class="nav-button">{{ if eq .PrevLink "index.html" }}← Home{{ else }}← Previous{{ end }}</a>
            {{end}}
            {{if .NextLink}}
            <a href="{{.NextLink}}" class="nav-button">Next: Exercise {{add .Number 1}} →</a>
            {{end}}
        </nav>
    </div>

    <footer>
        <div class="container">
            <p>🚀 Having Fun with the Go Source Code Workshop</p>
        </div>
    </footer>
</body>
</html>
`

const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Having Fun with the Go Source Code Workshop</title>
    <link rel="stylesheet" href="style.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/go.min.js"></script>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            hljs.highlightAll();

            // Add copy buttons to all code blocks
            document.querySelectorAll('pre').forEach(function(pre) {
                const button = document.createElement('button');
                button.className = 'copy-button';
                button.innerHTML = '<i class="far fa-copy"></i>';
                button.title = 'Copy to clipboard';

                button.addEventListener('click', function() {
                    const code = pre.querySelector('code');
                    const text = code.textContent;

                    navigator.clipboard.writeText(text).then(function() {
                        button.innerHTML = '<i class="fas fa-check"></i>';
                        button.classList.add('copied');
                        setTimeout(function() {
                            button.innerHTML = '<i class="far fa-copy"></i>';
                            button.classList.remove('copied');
                        }, 2000);
                    }).catch(function(err) {
                        console.error('Failed to copy:', err);
                    });
                });

                pre.appendChild(button);
            });
        });
    </script>
</head>
<body>
    <nav class="navbar">
        <div class="container">
            <a href="index.html" class="nav-home">🚀 Go Source Code Workshop</a>
            <div class="nav-links">
                <a href="index.html">Home</a>
            </div>
        </div>
    </nav>

    <div class="container">
        <header class="hero">
            <h1>🚀 Having Fun with the Go Source Code Workshop</h1>
            <p class="lead">Welcome to an interactive workshop where you'll learn how to modify and experiment with the Go programming language source code! This hands-on workshop will guide you through understanding, building, and making changes to the Go compiler and runtime. 🎯</p>
            <p class="version-note"><strong>📌 This workshop uses Go version 1.25.1</strong> - we'll check out the specific release tag to ensure consistency across all exercises.</p>
        </header>

        <section class="prerequisites">
            <h2>📋 Prerequisites</h2>
            <ul>
                <li>🐹 Basic knowledge of Go programming</li>
                <li>💻 Familiarity with command line tools</li>
                <li>🗂️ Git installed on your system</li>
                <li><strong>⚡ Go compiler version 1.24.6 or newer</strong> (required for bootstrapping the build process)</li>
                <li>💾 At least 4GB of free disk space</li>
            </ul>
        </section>

        <section class="overview">
            <h2>🎓 Workshop Overview</h2>
            <p>This workshop consists of {{len .Exercises}} exercises that will take you through the process from building Go from source, and making modifications at different places in the compiler, tooling and runtime. You'll gain some insights about the Go internals, from things like the lexer or parser, to runtime behaviors:</p>

            <div class="exercises-grid">
                {{range .Exercises}}
                <div class="exercise-card">
                    <div class="exercise-number">Exercise {{.Number}}</div>
                    <h3>{{.Emoji}} <a href="{{.Filename}}">{{.Title}}</a></h3>
                    <p>{{.Description}}</p>
                </div>
                {{end}}
            </div>
        </section>

        <section class="getting-started">
            <h2>🚀 Getting Started</h2>
            <ol>
                <li>🌱 Start with <a href="00-introduction-setup.html">Exercise 0</a> to set up your environment</li>
                <li>📚 Work through the exercises in order</li>
                <li>🔗 After exercise 1, you can pick and choose the exercise that you want.</li>
            </ol>
        </section>

        <section class="tips">
            <h2>💡 Tips for Success</h2>
            <ul>
                <li>⏰ Take your time with each exercise - compiler internals are complex!</li>
                <li>🔍 Don't hesitate to explore the Go source code beyond what's required</li>
                <li>🌿 Use <code>git</code> to track your changes and revert when needed</li>
                <li>🧪 Test your modifications thoroughly with various Go programs</li>
            </ul>
        </section>

        <section class="resources">
            <h2>📖 Resources</h2>
            <ul>
                <li>🏗️ <a href="https://github.com/golang/go/tree/master/src/cmd/compile">Go Compiler Overview</a></li>
                <li>📋 <a href="https://go.dev/ref/spec">Go Language Specification</a></li>
                <li>🔧 <a href="https://pkg.go.dev/runtime">Go Runtime Documentation</a></li>
                <li>📚 <a href="https://go.dev/src/">Go Internal Documentation</a></li>
            </ul>

            <h3>🎥 Video References</h3>
            <p>These workshop exercises are based on insights from my talks:</p>
            <ul>
                <li>🎬 <a href="https://www.youtube.com/watch?v=qnmoAA0WRgE">Understanding the Go Compiler</a> - Deep dive into Go's compilation process</li>
                <li>🎬 <a href="https://www.youtube.com/watch?v=YpRNFNFaLGY">Understanding the Go Runtime</a> - Exploration of Go's runtime system</li>
            </ul>
        </section>

        <section class="completion">
            <h2>🏆 Workshop Completion</h2>
            <p>Upon completing all exercises, you'll have:</p>
            <ul>
                <li>✅ <strong>Built Go from source</strong> and understood the bootstrap process</li>
                <li>✅ <strong>Modified language syntax</strong> by changing scanner and parser behavior</li>
                <li>✅ <strong>Customized development tools</strong> like gofmt and compiler optimizations</li>
                <li>✅ <strong>Implemented SSA optimizations</strong> in the compiler backend</li>
                <li>✅ <strong>Modified runtime behavior</strong> including program entry points and scheduler monitoring</li>
                <li>✅ <strong>Altered concurrency algorithms</strong> like select statement randomization</li>
                <li>✅ <strong>Customized error reporting</strong> with Java-style stack trace formatting</li>
            </ul>

            <p><strong>Congratulations!</strong> You'll have gained the confidence to keep exploring the Go source code. This knowledge enables you to:</p>
            <ul>
                <li>Start small contributions to the Go project</li>
                <li>Build custom language variants and tools</li>
                <li>Understand some trade-offs in language and runtime design</li>
            </ul>
        </section>

        <section class="contributing">
            <h2>🤝 Contributing</h2>
            <p>Found an issue, have an improvement idea or want to add more exercises? Please <a href="https://github.com/jespino/having-fun-with-the-go-source-code-workshop/issues">open an issue</a> or submit a pull request!</p>
        </section>

        <div class="cta">
            <a href="00-introduction-setup.html" class="cta-button">Start with Exercise 0 →</a>
        </div>
    </div>

    <footer>
        <div class="container">
            <p>🚀 Having Fun with the Go Source Code Workshop</p>
            <p><strong>Happy coding and welcome to the world of Go internals!</strong> ✨</p>
        </div>
    </footer>
</body>
</html>
`
