.PHONY: generate clean check-breaking

# Generate proto files for Go and TypeScript
generate:
	@echo "Generating proto files..."
	buf generate

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf gen/go
	rm -rf gen/ts

# Check for breaking changes using buf
check-breaking:
	@echo "Checking for breaking changes..."
	buf breaking --against '.git#branch=main'

# Format proto files
format:
	@echo "Formatting proto files..."
	buf format -w

# Lint proto files
lint:
	@echo "Linting proto files..."
	buf lint
