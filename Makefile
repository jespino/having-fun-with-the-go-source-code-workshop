.PHONY: website clean help publish

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

website: ## Generate the website from markdown exercises
	@echo "🚀 Generating website..."
	@cd website-generator && go run . -exercises ../exercises -output ../website
	@echo "✅ Website generated successfully in ./website/"

clean: ## Remove generated website files
	@echo "🧹 Cleaning website directory..."
	@rm -f website/*.html website/*.css
	@echo "✅ Website cleaned"

serve: ## Serve the website locally (requires Python)
	@echo "🌐 Starting local web server on http://localhost:8000"
	@cd website && python3 -m http.server 8000

publish: website ## Publish website to GitHub Pages
	@echo "📤 Publishing website to GitHub Pages..."
	@if [ ! -d ".git" ]; then \
		echo "❌ Error: Not in a git repository"; \
		exit 1; \
	fi
	@echo "🔄 Switching to gh-pages branch..."
	@git checkout -B gh-pages
	@echo "📁 Moving website files to root..."
	@git checkout main -- website
	@cp -r website/* .
	@rm -rf website
	@git add .
	@if git diff --cached --quiet; then \
		echo "ℹ️  No changes to publish"; \
	else \
		git commit -m "Update GitHub Pages"; \
		echo "✅ Changes committed to gh-pages branch"; \
	fi
	@echo "⬆️  Pushing to gh-pages..."
	@git push -f origin gh-pages
	@echo "🔄 Switching back to main branch..."
	@git checkout main
	@echo "✅ Website published successfully!"
	@echo "🌐 Your site will be available at: https://jespino.github.io/having-fun-with-the-go-source-code-workshop/"
