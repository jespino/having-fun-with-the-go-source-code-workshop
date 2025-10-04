.PHONY: website clean help

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
