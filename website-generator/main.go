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
	Description string
	Filename    string
	Content     template.HTML
	PrevLink    string
	NextLink    string
	Lang        string
	AltLangURL  string
	AltLangName string
	CSSPath     string
	HomePath    string
}

type IndexData struct {
	Exercises   []Exercise
	Lang        string
	AltLangURL  string
	AltLangName string
	CSSPath     string
	HomePath    string
}

type exerciseMeta struct {
	Filename    string
	Title       string
	Description string
}

type LangConfig struct {
	Code          string
	FileSuffix    string // ".md" for English, ".es.md" for Spanish
	OutputPrefix  string // "" for English, "es" for Spanish
	AltLangPrefix string // "es" for English, "" for Spanish
	AltLangName   string // "Español" for English, "English" for Spanish
	Metadata      []exerciseMeta
	UIStrings     UIStrings
}

type UIStrings struct {
	Home                string
	Previous            string
	Next                string
	Exercise            string
	HeroTitle           string
	HeroLead            string
	HeroVersionNote     string
	Prerequisites       string
	PrereqItems         []string
	Overview            string
	OverviewText        string
	GettingStarted      string
	GettingStartedItems []string
	Tips                string
	TipItems            []string
	Resources           string
	VideoReferences     string
	VideoRefsIntro      string
	VideoCompiler       string
	VideoCompilerDesc   string
	VideoRuntime        string
	VideoRuntimeDesc    string
	Completion          string
	CompletionIntro     string
	CompletionItems     []string
	CompletionCongrats  string
	CompletionEnables   []string
	Contributing        string
	ContributingText    string
	CTAButton           string
	FooterTitle         string
	FooterCreatedBy     string
}

var englishConfig = LangConfig{
	Code:          "en",
	FileSuffix:    ".md",
	OutputPrefix:  "",
	AltLangPrefix: "es",
	AltLangName:   "Español",
	Metadata: []exerciseMeta{
		{"00-introduction-setup", "Introduction and Setup", "Get started by cloning and setting up the Go source code environment."},
		{"01-compile-go-unchanged", "Compiling Go Without Changes", "Learn to build the Go toolchain from source without any modifications."},
		{"02-scanner-arrow-operator", "Adding the \"=>\" Arrow Operator for Goroutines", "Learn scanner/lexer modification by adding \"=>\" as an alternative syntax for starting goroutines."},
		{"03-parser-multiple-go", "Multiple \"go\" Keywords - Parser Enhancement", "Learn parser modification by enabling multiple consecutive \"go\" keywords (go go go myFunction)."},
		{"04-compiler-inlining-parameters", "Inline Parameters - Function Inlining Experiments", "Explore the inliner behavior by modifying function inlining parameters."},
		{"05-gofmt-ast-transformation", "gofmt Modification - Indentation & AST Transformation", "Modify gofmt to use 4 spaces instead of tabs and add a custom AST transformation replacing \"hello\" with \"helo\"."},
		{"06-ssa-power-of-two-detector", "SSA Pass - Detecting Division by Powers of Two", "Create a custom SSA compiler pass that detects division operations by powers of two that could be optimized to bit shifts."},
		{"07-runtime-patient-go", "Patient Go - Making Go Wait for Goroutines", "Modify the Go runtime to wait for all goroutines to complete before program termination."},
		{"08-goroutine-sleep-detective", "Goroutine Sleep Detective - Runtime State Monitoring", "Add logging to the Go scheduler to monitor goroutines going to sleep."},
		{"09-predictable-select", "Predictable Select - Removing Randomness from Go's Select Statement", "Modify Go's select statement implementation to be deterministic instead of random."},
		{"10-java-style-stack-traces", "Java-Style Stack Traces - Making Go Panics Look Familiar", "Transform Go's verbose stack traces into Java-style formatting."},
		{"11-dnd-work-stealing", "D&D Work Stealing - Rolling for Goroutines", "Add a d20 dice roll to Go's work stealing scheduler to gate goroutine theft between processors."},
	},
	UIStrings: UIStrings{
		Home:            "Home",
		Previous:        "Previous",
		Next:            "Next",
		Exercise:        "Exercise",
		HeroTitle:       "Having fun with the Go Source Code",
		HeroLead:        "Welcome to an interactive workshop where you'll learn how to modify and experiment with the Go programming language source code! This hands-on workshop will guide you through understanding, building, and making changes to the Go compiler and runtime.",
		HeroVersionNote: "<strong>This workshop uses Go version 1.26.1</strong> - we'll check out the specific release tag to ensure consistency across all exercises.",
		Prerequisites:   "Prerequisites",
		PrereqItems: []string{
			"Basic knowledge of Go programming",
			"Familiarity with command line tools",
			"Git installed on your system",
			"<strong>Go compiler version 1.24 or newer</strong> (required for bootstrapping the build process)",
			"At least 4GB of free disk space",
		},
		Overview:       "Workshop Overview",
		OverviewText:   "This workshop consists of %d exercises that will take you through the process from building Go from source, and making modifications at different places in the compiler, tooling and runtime. You'll gain some insights about the Go internals, from things like the lexer or parser, to runtime behaviors:",
		GettingStarted: "Getting Started",
		GettingStartedItems: []string{
			`Start with <a href="%s00-introduction-setup.html">Exercise 0</a> to set up your environment`,
			"Work through the exercises in order",
			"After exercise 1, you can pick and choose the exercise that you want.",
		},
		Tips: "Tips for Success",
		TipItems: []string{
			"Take your time with each exercise - compiler internals are complex!",
			"Don't hesitate to explore the Go source code beyond what's required",
			"Use <code>git</code> to track your changes and revert when needed",
			"Test your modifications thoroughly with various Go programs",
		},
		Resources:         "Resources",
		VideoReferences:   "Video References",
		VideoRefsIntro:    "These workshop exercises are based on insights from my talks:",
		VideoCompiler:     "Understanding the Go Compiler",
		VideoCompilerDesc: "Deep dive into Go's compilation process",
		VideoRuntime:      "Understanding the Go Runtime",
		VideoRuntimeDesc:  "Exploration of Go's runtime system",
		Completion:        "Workshop Completion",
		CompletionIntro:   "Upon completing all exercises, you'll have:",
		CompletionItems: []string{
			"<strong>Built Go from source</strong> and understood the bootstrap process",
			"<strong>Modified language syntax</strong> by changing scanner and parser behavior",
			"<strong>Customized development tools</strong> like gofmt and compiler optimizations",
			"<strong>Implemented SSA optimizations</strong> in the compiler backend",
			"<strong>Modified runtime behavior</strong> including program entry points and scheduler monitoring",
			"<strong>Altered concurrency algorithms</strong> like select statement randomization",
			"<strong>Customized error reporting</strong> with Java-style stack trace formatting",
		},
		CompletionCongrats: "<strong>Congratulations!</strong> You'll have gained the confidence to keep exploring the Go source code. This knowledge enables you to:",
		CompletionEnables: []string{
			"Start small contributions to the Go project",
			"Build custom language variants and tools",
			"Understand some trade-offs in language and runtime design",
		},
		Contributing:     "Contributing",
		ContributingText: `Found an issue, have an improvement idea or want to add more exercises? Please <a href="https://github.com/jespino/having-fun-with-the-go-source-code-workshop/issues">open an issue</a> or submit a pull request!`,
		CTAButton:        "Start with Exercise 0 →",
		FooterTitle:      "Having fun with the Go Source Code",
		FooterCreatedBy:  "Created by <strong>Jesús Espino</strong>",
	},
}

var spanishConfig = LangConfig{
	Code:          "es",
	FileSuffix:    ".es.md",
	OutputPrefix:  "es",
	AltLangPrefix: "",
	AltLangName:   "English",
	Metadata: []exerciseMeta{
		{"00-introduction-setup", "Introducción y Configuración", "Comienza clonando y configurando el entorno del código fuente de Go."},
		{"01-compile-go-unchanged", "Compilando Go Sin Cambios", "Aprende a compilar el toolchain de Go desde el código fuente sin modificaciones."},
		{"02-scanner-arrow-operator", "Añadiendo el Operador Flecha \"=>\" para Goroutines", "Aprende a modificar el scanner/lexer añadiendo \"=>\" como sintaxis alternativa para iniciar goroutines."},
		{"03-parser-multiple-go", "Múltiples Keywords \"go\" - Mejora del Parser", "Aprende a modificar el parser permitiendo múltiples keywords \"go\" consecutivos (go go go myFunction)."},
		{"04-compiler-inlining-parameters", "Parámetros de Inlining - Experimentos con Function Inlining", "Explora el comportamiento del inliner modificando los parámetros de inlining de funciones."},
		{"05-gofmt-ast-transformation", "Modificación de gofmt - Indentación y Transformación AST", "Modifica gofmt para usar 4 espacios en lugar de tabs y añade una transformación AST personalizada reemplazando \"hello\" con \"helo\"."},
		{"06-ssa-power-of-two-detector", "Pase SSA - Detectando División por Potencias de Dos", "Crea un pase SSA personalizado en el compilador que detecta operaciones de división por potencias de dos que podrían optimizarse con bit shifts."},
		{"07-runtime-patient-go", "Go Paciente - Haciendo que Go Espere a las Goroutines", "Modifica el runtime de Go para esperar a que todas las goroutines terminen antes de finalizar el programa."},
		{"08-goroutine-sleep-detective", "Detective de Goroutines Dormidas - Monitoreo del Estado del Runtime", "Añade logging al scheduler de Go para monitorear goroutines que se van a dormir."},
		{"09-predictable-select", "Select Predecible - Eliminando la Aleatoriedad del Select de Go", "Modifica la implementación del select de Go para que sea determinista en lugar de aleatorio."},
		{"10-java-style-stack-traces", "Stack Traces Estilo Java - Haciendo los Panics de Go Familiares", "Transforma los stack traces verbosos de Go al formato estilo Java."},
	},
	UIStrings: UIStrings{
		Home:            "Inicio",
		Previous:        "Anterior",
		Next:            "Siguiente",
		Exercise:        "Ejercicio",
		HeroTitle:       "Divirtiéndonos con el Código Fuente de Go",
		HeroLead:        "¡Bienvenido a un taller interactivo donde aprenderás a modificar y experimentar con el código fuente del lenguaje de programación Go! Este taller práctico te guiará a través de la comprensión, compilación y modificación del compilador y runtime de Go.",
		HeroVersionNote: "<strong>Este taller usa Go versión 1.26.1</strong> - haremos checkout del tag de release específico para asegurar consistencia en todos los ejercicios.",
		Prerequisites:   "Prerrequisitos",
		PrereqItems: []string{
			"Conocimientos básicos de programación en Go",
			"Familiaridad con herramientas de línea de comandos",
			"Git instalado en tu sistema",
			"<strong>Compilador de Go versión 1.24 o superior</strong> (necesario para el proceso de bootstrapping)",
			"Al menos 4GB de espacio libre en disco",
		},
		Overview:       "Descripción General del Taller",
		OverviewText:   "Este taller consta de %d ejercicios que te llevarán a través del proceso desde compilar Go desde el código fuente hasta hacer modificaciones en diferentes partes del compilador, herramientas y runtime. Obtendrás conocimientos sobre los internos de Go, desde cosas como el lexer o parser, hasta comportamientos del runtime:",
		GettingStarted: "Cómo Empezar",
		GettingStartedItems: []string{
			`Comienza con el <a href="%s00-introduction-setup.html">Ejercicio 0</a> para configurar tu entorno`,
			"Trabaja los ejercicios en orden",
			"Después del ejercicio 1, puedes elegir el ejercicio que quieras.",
		},
		Tips: "Consejos para el Éxito",
		TipItems: []string{
			"Tómate tu tiempo con cada ejercicio - ¡los internos del compilador son complejos!",
			"No dudes en explorar el código fuente de Go más allá de lo requerido",
			"Usa <code>git</code> para rastrear tus cambios y revertir cuando sea necesario",
			"Prueba tus modificaciones a fondo con varios programas Go",
		},
		Resources:         "Recursos",
		VideoReferences:   "Referencias en Video",
		VideoRefsIntro:    "Los ejercicios de este taller están basados en ideas de mis charlas:",
		VideoCompiler:     "Entendiendo el Compilador de Go",
		VideoCompilerDesc: "Profundización en el proceso de compilación de Go",
		VideoRuntime:      "Entendiendo el Runtime de Go",
		VideoRuntimeDesc:  "Exploración del sistema de runtime de Go",
		Completion:        "Completando el Taller",
		CompletionIntro:   "Al completar todos los ejercicios, habrás:",
		CompletionItems: []string{
			"<strong>Compilado Go desde el código fuente</strong> y entendido el proceso de bootstrap",
			"<strong>Modificado la sintaxis del lenguaje</strong> cambiando el comportamiento del scanner y parser",
			"<strong>Personalizado herramientas de desarrollo</strong> como gofmt y optimizaciones del compilador",
			"<strong>Implementado optimizaciones SSA</strong> en el backend del compilador",
			"<strong>Modificado el comportamiento del runtime</strong> incluyendo puntos de entrada del programa y monitoreo del scheduler",
			"<strong>Alterado algoritmos de concurrencia</strong> como la aleatorización del select",
			"<strong>Personalizado el reporte de errores</strong> con formato de stack traces estilo Java",
		},
		CompletionCongrats: "<strong>¡Felicidades!</strong> Habrás ganado la confianza para seguir explorando el código fuente de Go. Este conocimiento te permite:",
		CompletionEnables: []string{
			"Comenzar pequeñas contribuciones al proyecto Go",
			"Construir variantes personalizadas del lenguaje y herramientas",
			"Entender algunas decisiones de diseño del lenguaje y runtime",
		},
		Contributing:     "Contribuir",
		ContributingText: `¿Encontraste un problema, tienes una idea de mejora o quieres añadir más ejercicios? ¡Por favor <a href="https://github.com/jespino/having-fun-with-the-go-source-code-workshop/issues">abre un issue</a> o envía un pull request!`,
		CTAButton:        "Comenzar con el Ejercicio 0 →",
		FooterTitle:      "Divirtiéndonos con el Código Fuente de Go",
		FooterCreatedBy:  "Creado por <strong>Jesús Espino</strong>",
	},
}

var languages = []LangConfig{englishConfig, spanishConfig}

// exerciseMetadata is kept for backward compatibility with serve.go
var exerciseMetadata = englishConfig.Metadata

func main() {
	exercisesDir := flag.String("exercises", "../exercises", "Path to exercises directory")
	outputDir := flag.String("output", "../website", "Path to output directory")
	serve := flag.Bool("serve", false, "Start dev server with live reload")
	port := flag.Int("port", 8000, "Dev server port (used with -serve)")
	flag.Parse()

	if *serve {
		srv := newDevServer(*exercisesDir, *outputDir, *port)
		if err := srv.run(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	totalPages := 0
	for _, lang := range languages {
		// Determine output directory for this language
		langOutputDir := *outputDir
		if lang.OutputPrefix != "" {
			langOutputDir = filepath.Join(*outputDir, lang.OutputPrefix)
			if err := os.MkdirAll(langOutputDir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory for %s: %v\n", lang.Code, err)
				os.Exit(1)
			}
		}

		// Determine CSS path relative to output dir
		cssPath := "style.css"
		if lang.OutputPrefix != "" {
			cssPath = "../style.css"
		}

		// Determine the home path prefix
		homePath := ""
		if lang.OutputPrefix != "" {
			homePath = ""
		}

		// Determine alt lang URL prefix
		altLangURLPrefix := "../"
		if lang.OutputPrefix == "" {
			altLangURLPrefix = "es/"
		}

		// Generate exercise pages
		exercises := make([]Exercise, 0, len(lang.Metadata))
		for i, meta := range lang.Metadata {
			exercise, err := generateExercisePage(*exercisesDir, langOutputDir, lang, meta, i, cssPath, homePath, altLangURLPrefix)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating exercise %s (%s): %v\n", meta.Filename, lang.Code, err)
				os.Exit(1)
			}
			exercises = append(exercises, exercise)
		}

		// Generate index page
		if err := generateIndexPage(langOutputDir, lang, exercises, cssPath, homePath, altLangURLPrefix); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating index page (%s): %v\n", lang.Code, err)
			os.Exit(1)
		}

		totalPages += len(exercises) + 1
	}

	// Copy CSS file (only at root level, shared by all languages)
	if err := copyCSSFile(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying CSS file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Website generated successfully!")
	fmt.Printf("📁 Output directory: %s\n", *outputDir)
	fmt.Printf("📄 Generated %d pages total (including all languages)\n", totalPages)
}

func generateExercisePage(exercisesDir, outputDir string, lang LangConfig, meta exerciseMeta, index int, cssPath, homePath, altLangURLPrefix string) (Exercise, error) {
	// Read markdown file
	mdFilename := meta.Filename + lang.FileSuffix
	mdPath := filepath.Join(exercisesDir, mdFilename)
	content, err := os.ReadFile(mdPath)
	if err != nil {
		return Exercise{}, fmt.Errorf("reading markdown file: %w", err)
	}

	// Convert markdown to HTML
	htmlContent := markdownToHTML(content)

	// Generate HTML filename
	htmlFilename := meta.Filename + ".html"

	// Determine prev/next links
	prevLink := homePath + "index.html"
	if index > 0 {
		prevLink = lang.Metadata[index-1].Filename + ".html"
	}

	nextLink := ""
	if index < len(lang.Metadata)-1 {
		nextLink = lang.Metadata[index+1].Filename + ".html"
	}

	// Alt language URL for the same exercise
	altLangURL := altLangURLPrefix + htmlFilename

	exercise := Exercise{
		Number:      index,
		Title:       meta.Title,
		Description: meta.Description,
		Filename:    htmlFilename,
		Content:     template.HTML(htmlContent),
		PrevLink:    prevLink,
		NextLink:    nextLink,
		Lang:        lang.Code,
		AltLangURL:  altLangURL,
		AltLangName: lang.AltLangName,
		CSSPath:     cssPath,
		HomePath:    homePath,
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

	fmt.Printf("✓ Generated %s [%s]\n", htmlFilename, lang.Code)
	return exercise, nil
}

func generateIndexPage(outputDir string, lang LangConfig, exercises []Exercise, cssPath, homePath, altLangURLPrefix string) error {
	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	outputPath := filepath.Join(outputDir, "index.html")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	// Format overview text with exercise count
	ui := lang.UIStrings
	ui.OverviewText = fmt.Sprintf(ui.OverviewText, len(exercises))

	// Format getting started items with link prefix
	formattedGSItems := make([]string, len(ui.GettingStartedItems))
	for i, item := range ui.GettingStartedItems {
		if strings.Contains(item, "%s") {
			formattedGSItems[i] = fmt.Sprintf(item, "")
		} else {
			formattedGSItems[i] = item
		}
	}
	ui.GettingStartedItems = formattedGSItems

	data := struct {
		IndexData
		UI              UIStrings
		AltLangURLIndex string
	}{
		IndexData: IndexData{
			Exercises:   exercises,
			Lang:        lang.Code,
			AltLangURL:  altLangURLPrefix + "index.html",
			AltLangName: lang.AltLangName,
			CSSPath:     cssPath,
			HomePath:    homePath,
		},
		UI:              ui,
		AltLangURLIndex: altLangURLPrefix + "index.html",
	}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	fmt.Printf("✓ Generated index.html [%s]\n", lang.Code)
	return nil
}

func copyCSSFile(outputDir string) error {
	cssContent := cssTemplate
	outputPath := filepath.Join(outputDir, "style.css")

	if err := os.WriteFile(outputPath, []byte(cssContent), 0o644); err != nil {
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

	// Fix links that are already in the format XX-name.md or XX-name.es.md
	re = regexp.MustCompile(`href="(?:\./)?([0-9]{2}-[^"]+?)(?:\.es)?\.md"`)
	html = re.ReplaceAllString(html, `href="$1.html"`)

	return html
}

const exerciseTemplate = `<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exercise {{.Number}}: {{.Title}} - Go Source Code Workshop</title>
    <link rel="stylesheet" href="{{.CSSPath}}">
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
            <a href="{{.HomePath}}index.html" class="nav-home">Having fun with the Go Source Code</a>
            <div class="nav-links">
                <a href="{{.HomePath}}index.html">{{if eq .Lang "es"}}Inicio{{else}}Home{{end}}</a>
                <a href="{{.AltLangURL}}" class="lang-switch"><i class="fas fa-globe"></i> {{.AltLangName}}</a>
                <a href="https://github.com/jespino/having-fun-with-the-go-source-code-workshop" target="_blank"><i class="fab fa-github"></i> Repository</a>
            </div>
        </div>
    </nav>

    <div class="container">
        <article class="exercise-content">
            {{.Content}}
        </article>

        <nav class="exercise-nav">
            {{if .PrevLink}}
            <a href="{{.PrevLink}}" class="nav-button">{{ if eq .PrevLink "index.html" }}{{if eq .Lang "es"}}← Inicio{{else}}← Home{{end}}{{ else }}{{if eq .Lang "es"}}← Anterior{{else}}← Previous{{end}}{{ end }}</a>
            {{end}}
            {{if .NextLink}}
            <a href="{{.NextLink}}" class="nav-button">{{if eq .Lang "es"}}Siguiente{{else}}Next{{end}}: {{if eq .Lang "es"}}Ejercicio{{else}}Exercise{{end}} {{add .Number 1}} →</a>
            {{end}}
        </nav>
    </div>

    <footer>
        <div class="container">
            <p>Having fun with the Go Source Code</p>
            <p>{{if eq .Lang "es"}}Creado por{{else}}Created by{{end}} <strong>Jesús Espino</strong></p>
            <div class="footer-links">
                <a href="https://github.com/jespino" target="_blank"><i class="fab fa-github"></i> GitHub</a>
                <a href="https://x.com/jespinog" target="_blank"><i class="fab fa-x-twitter"></i> @jespinog</a>
                <a href="https://linkedin.com/in/jesus-espino" target="_blank"><i class="fab fa-linkedin"></i> LinkedIn</a>
            </div>
        </div>
    </footer>
</body>
</html>
`

const indexTemplate = `<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.UI.HeroTitle}}</title>
    <link rel="stylesheet" href="{{.CSSPath}}">
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
            <a href="{{.HomePath}}index.html" class="nav-home">Having fun with the Go Source Code</a>
            <div class="nav-links">
                <a href="{{.HomePath}}index.html">{{.UI.Home}}</a>
                <a href="{{.AltLangURL}}" class="lang-switch"><i class="fas fa-globe"></i> {{.AltLangName}}</a>
                <a href="https://github.com/jespino/having-fun-with-the-go-source-code-workshop" target="_blank"><i class="fab fa-github"></i> Repository</a>
            </div>
        </div>
    </nav>

    <div class="container">
        <header class="hero">
            <h1>{{.UI.HeroTitle}}</h1>
            <p class="lead">{{.UI.HeroLead}}</p>
            <p class="version-note">{{safeHTML .UI.HeroVersionNote}}</p>
        </header>

        <section class="prerequisites">
            <h2>{{.UI.Prerequisites}}</h2>
            <ul>
                {{range .UI.PrereqItems}}<li>{{safeHTML .}}</li>
                {{end}}
            </ul>
        </section>

        <section class="overview">
            <h2>{{.UI.Overview}}</h2>
            <p>{{safeHTML .UI.OverviewText}}</p>

            <div class="exercises-grid">
                {{range .Exercises}}
                <a href="{{.Filename}}" class="exercise-card-link">
                    <div class="exercise-card">
                        <div class="exercise-number">{{if eq .Lang "es"}}Ejercicio{{else}}Exercise{{end}} {{.Number}}</div>
                        <h3>{{.Title}}</h3>
                        <p>{{.Description}}</p>
                    </div>
                </a>
                {{end}}
            </div>
        </section>

        <section class="getting-started">
            <h2>{{.UI.GettingStarted}}</h2>
            <ol>
                {{range .UI.GettingStartedItems}}<li>{{safeHTML .}}</li>
                {{end}}
            </ol>
        </section>

        <section class="tips">
            <h2>{{.UI.Tips}}</h2>
            <ul>
                {{range .UI.TipItems}}<li>{{safeHTML .}}</li>
                {{end}}
            </ul>
        </section>

        <section class="resources">
            <h2>{{.UI.Resources}}</h2>
            <ul>
                <li><a href="https://github.com/golang/go/tree/master/src/cmd/compile">Go Compiler Overview</a></li>
                <li><a href="https://go.dev/ref/spec">Go Language Specification</a></li>
                <li><a href="https://pkg.go.dev/runtime">Go Runtime Documentation</a></li>
            </ul>

            <h3>{{.UI.VideoReferences}}</h3>
            <p>{{.UI.VideoRefsIntro}}</p>
            <div class="video-grid">
                <div class="video-container">
                    <h4>{{.UI.VideoCompiler}}</h4>
                    <iframe src="https://www.youtube.com/embed/qnmoAA0WRgE" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
                    <p>{{.UI.VideoCompilerDesc}}</p>
                </div>
                <div class="video-container">
                    <h4>{{.UI.VideoRuntime}}</h4>
                    <iframe src="https://www.youtube.com/embed/YpRNFNFaLGY" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
                    <p>{{.UI.VideoRuntimeDesc}}</p>
                </div>
            </div>
        </section>

        <section class="completion">
            <h2>{{.UI.Completion}}</h2>
            <p>{{.UI.CompletionIntro}}</p>
            <ul>
                {{range .UI.CompletionItems}}<li>{{safeHTML .}}</li>
                {{end}}
            </ul>

            <p>{{safeHTML .UI.CompletionCongrats}}</p>
            <ul>
                {{range .UI.CompletionEnables}}<li>{{.}}</li>
                {{end}}
            </ul>
        </section>

        <section class="contributing">
            <h2>{{.UI.Contributing}}</h2>
            <p>{{safeHTML .UI.ContributingText}}</p>
        </section>

        <div class="cta">
            <a href="00-introduction-setup.html" class="cta-button">{{.UI.CTAButton}}</a>
        </div>
    </div>

    <footer>
        <div class="container">
            <p>{{.UI.FooterTitle}}</p>
            <p>{{safeHTML .UI.FooterCreatedBy}}</p>
            <div class="footer-links">
                <a href="https://github.com/jespino" target="_blank"><i class="fab fa-github"></i> GitHub</a>
                <a href="https://x.com/jespinog" target="_blank"><i class="fab fa-x-twitter"></i> @jespinog</a>
                <a href="https://linkedin.com/in/jesus-espino" target="_blank"><i class="fab fa-linkedin"></i> LinkedIn</a>
            </div>
        </div>
    </footer>
</body>
</html>
`
