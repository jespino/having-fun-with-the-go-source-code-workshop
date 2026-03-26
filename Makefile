.PHONY: website clean help publish iximiuz

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
	@rm -rf website/es
	@echo "✅ Website cleaned"

serve: ## Serve the website locally with live reload
	@cd website-generator && go run . -exercises ../exercises -output ../website -serve

iximiuz: ## Generate the iximiuz Labs tutorial from markdown exercises
	@echo "🚀 Generating iximiuz Labs tutorial..."
	@cd iximiuz-generator && go run . -exercises ../exercises -output ../fun-with-go-code-8fc2f532
	@echo "✅ iximiuz tutorial generated in ./fun-with-go-code-8fc2f532/"
