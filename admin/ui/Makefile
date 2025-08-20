# Makefile for Admin UI

.PHONY: help check build clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make check  - Quick validation (TypeScript + ESLint)"
	@echo "  make build  - Production build with validation"
	@echo "  make clean  - Clean build artifacts"

# Quick validation - the most important command
check:
	@echo "📝 Checking TypeScript types..."
	@npm run type-check
	@echo "🧹 Running ESLint..."
	@npm run lint
	@echo "✅ Quick checks passed!"

# Production build with validation
build:
	@echo "🏗️  Building application..."
	@npm run build
	@echo "✅ Build successful!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf dist