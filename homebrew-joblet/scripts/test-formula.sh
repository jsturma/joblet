#!/bin/bash
set -e

# RNX Homebrew Formula Testing Script
# Comprehensive testing for the RNX homebrew formula

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
FORMULA_PATH="$REPO_ROOT/Formula/rnx.rb"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "\n${BLUE}🧪 $1${NC}"
    echo "=========================================="
}

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    local test_command="$2"
    local should_pass="${3:-true}"
    
    log_info "Running: $test_name"
    
    if [ "$should_pass" = "true" ]; then
        if eval "$test_command" &>/dev/null; then
            log_success "$test_name"
            ((TESTS_PASSED++))
            return 0
        else
            log_error "$test_name"
            FAILED_TESTS+=("$test_name")
            ((TESTS_FAILED++))
            return 1
        fi
    else
        # Test should fail
        if eval "$test_command" &>/dev/null; then
            log_error "$test_name (expected to fail but passed)"
            FAILED_TESTS+=("$test_name")
            ((TESTS_FAILED++))
            return 1
        else
            log_success "$test_name (correctly failed as expected)"
            ((TESTS_PASSED++))
            return 0
        fi
    fi
}

run_test_with_output() {
    local test_name="$1"
    local test_command="$2"
    
    log_info "Running: $test_name"
    
    local output
    if output=$(eval "$test_command" 2>&1); then
        log_success "$test_name"
        if [ -n "$output" ]; then
            echo "  Output: $output"
        fi
        ((TESTS_PASSED++))
        return 0
    else
        log_error "$test_name"
        echo "  Error: $output"
        FAILED_TESTS+=("$test_name")
        ((TESTS_FAILED++))
        return 1
    fi
}

cleanup_test_installations() {
    log_info "Cleaning up any existing test installations..."
    
    # Uninstall RNX if installed
    if brew list | grep -q "^rnx$"; then
        brew uninstall rnx --ignore-dependencies || true
    fi
    
    # Remove tap if added
    if brew tap | grep -q "ehsaniara/joblet"; then
        brew untap ehsaniara/joblet || true
    fi
    
    # Clean homebrew cache
    brew cleanup --prune=0 || true
}

# Main testing function
main() {
    log_header "RNX Homebrew Formula Testing Suite"
    
    # Check if running on macOS
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "This script must be run on macOS"
        exit 1
    fi
    
    # Check prerequisites
    log_header "Prerequisites Check"
    
    run_test "Homebrew installed" "command -v brew"
    run_test "Formula file exists" "test -f '$FORMULA_PATH'"
    run_test "Ruby available" "command -v ruby"
    
    if [ $TESTS_FAILED -gt 0 ]; then
        log_error "Prerequisites not met. Aborting."
        exit 1
    fi
    
    # Formula syntax validation
    log_header "Formula Syntax Validation"
    
    run_test "Formula syntax check" "ruby -c '$FORMULA_PATH'"
    run_test_with_output "Homebrew audit (basic)" "brew audit --formula '$FORMULA_PATH'"
    run_test "Homebrew audit (strict)" "brew audit --strict --formula '$FORMULA_PATH'"
    
    # Formula content validation
    log_header "Formula Content Validation"
    
    run_test "Has required class name" "grep -q 'class Rnx < Formula' '$FORMULA_PATH'"
    run_test "Has description" "grep -q 'desc ' '$FORMULA_PATH'"
    run_test "Has homepage" "grep -q 'homepage ' '$FORMULA_PATH'"
    run_test "Has license" "grep -q 'license ' '$FORMULA_PATH'"
    run_test "Has URL for Intel" "grep -q 'if Hardware::CPU.intel?' '$FORMULA_PATH'"
    run_test "Has URL for ARM64" "grep -q 'else' '$FORMULA_PATH'"
    run_test "Has SHA256 placeholders or values" "grep -q 'sha256 ' '$FORMULA_PATH'"
    run_test "Has install method" "grep -q 'def install' '$FORMULA_PATH'"
    run_test "Has test method" "grep -q 'def test' '$FORMULA_PATH'"
    run_test "Has Node.js dependency" "grep -q 'depends_on \"node\"' '$FORMULA_PATH'"
    run_test "Has installation options" "grep -q 'option \"' '$FORMULA_PATH'"
    
    # URL validation (if not placeholders)
    log_header "URL Validation"
    
    # Extract URLs from formula
    INTEL_URL=$(grep -A 1 'if Hardware::CPU.intel?' "$FORMULA_PATH" | grep 'url ' | sed 's/.*url "\(.*\)".*/\1/')
    ARM64_URL=$(grep -A 3 'else' "$FORMULA_PATH" | grep 'url ' | sed 's/.*url "\(.*\)".*/\1/')
    
    if [[ "$INTEL_URL" != *"PLACEHOLDER"* ]]; then
        run_test "Intel URL accessible" "curl -sI '$INTEL_URL' | grep -q '200 OK'"
    else
        log_warning "Intel URL contains placeholder, skipping accessibility test"
    fi
    
    if [[ "$ARM64_URL" != *"PLACEHOLDER"* ]]; then
        run_test "ARM64 URL accessible" "curl -sI '$ARM64_URL' | grep -q '200 OK'"
    else
        log_warning "ARM64 URL contains placeholder, skipping accessibility test"
    fi
    
    # Installation testing
    log_header "Installation Testing"
    
    cleanup_test_installations
    
    # Test CLI-only installation
    log_info "Testing CLI-only installation..."
    if HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source "$FORMULA_PATH" --without-admin --quiet; then
        log_success "CLI-only installation"
        ((TESTS_PASSED++))
        
        # Test CLI functionality
        run_test_with_output "RNX binary installed" "command -v rnx"
        run_test_with_output "RNX version output" "rnx --version"
        run_test "RNX help output" "rnx --help | grep -q 'Usage:'"
        
        # Check that admin UI is not installed
        BREW_PREFIX=$(brew --prefix)
        run_test "Admin UI not installed (CLI-only)" "test ! -d '$BREW_PREFIX/share/rnx/admin'" false
        
        brew uninstall rnx --ignore-dependencies
    else
        log_error "CLI-only installation failed"
        FAILED_TESTS+=("CLI-only installation")
        ((TESTS_FAILED++))
    fi
    
    cleanup_test_installations
    
    # Test with admin UI (if Node.js available)
    if command -v node &>/dev/null; then
        log_info "Node.js detected, testing admin UI installation..."
        
        if HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source "$FORMULA_PATH" --with-admin --quiet; then
            log_success "Admin UI installation"
            ((TESTS_PASSED++))
            
            # Test CLI functionality
            run_test_with_output "RNX binary installed (with admin)" "command -v rnx"
            run_test_with_output "RNX version output (with admin)" "rnx --version"
            
            # Test admin UI files
            BREW_PREFIX=$(brew --prefix)
            run_test "Admin UI directory exists" "test -d '$BREW_PREFIX/share/rnx/admin'"
            run_test "Admin server files exist" "test -f '$BREW_PREFIX/share/rnx/admin/server/package.json'"
            run_test "Admin UI build exists" "test -f '$BREW_PREFIX/share/rnx/admin/ui/dist/index.html'"
            run_test "Admin launcher script exists" "test -x '$BREW_PREFIX/bin/rnx-admin'"
            
            # Test Node.js dependencies
            run_test "Admin server dependencies installed" "test -d '$BREW_PREFIX/share/rnx/admin/server/node_modules'"
            
            brew uninstall rnx --ignore-dependencies
        else
            log_error "Admin UI installation failed"
            FAILED_TESTS+=("Admin UI installation")
            ((TESTS_FAILED++))
        fi
    else
        log_warning "Node.js not available, skipping admin UI installation test"
    fi
    
    cleanup_test_installations
    
    # Formula option validation
    log_header "Formula Options Validation"
    
    run_test "Has with-admin option" "grep -q 'option \"with-admin\"' '$FORMULA_PATH'"
    run_test "Has without-admin option" "grep -q 'option \"without-admin\"' '$FORMULA_PATH'"
    run_test "Has determine_admin_installation method" "grep -q 'def determine_admin_installation' '$FORMULA_PATH'"
    run_test "Has setup_admin_ui method" "grep -q 'def setup_admin_ui' '$FORMULA_PATH'"
    run_test "Has create_admin_launcher method" "grep -q 'def create_admin_launcher' '$FORMULA_PATH'"
    run_test "Has caveats method" "grep -q 'def caveats' '$FORMULA_PATH'"
    
    # Archive structure validation (if URLs are not placeholders)
    if [[ "$INTEL_URL" != *"PLACEHOLDER"* ]] && [[ "$ARM64_URL" != *"PLACEHOLDER"* ]]; then
        log_header "Archive Structure Validation"
        
        TEMP_DIR=$(mktemp -d)
        
        # Test Intel archive
        log_info "Testing Intel archive structure..."
        if curl -sL "$INTEL_URL" -o "$TEMP_DIR/intel.tar.gz"; then
            if tar -tzf "$TEMP_DIR/intel.tar.gz" | grep -q "^rnx$"; then
                log_success "Intel archive contains rnx binary"
                ((TESTS_PASSED++))
            else
                log_error "Intel archive missing rnx binary"
                FAILED_TESTS+=("Intel archive structure")
                ((TESTS_FAILED++))
            fi
            
            if tar -tzf "$TEMP_DIR/intel.tar.gz" | grep -q "^admin/"; then
                log_success "Intel archive contains admin directory"
                ((TESTS_PASSED++))
            else
                log_error "Intel archive missing admin directory"
                FAILED_TESTS+=("Intel archive admin directory")
                ((TESTS_FAILED++))
            fi
        else
            log_error "Failed to download Intel archive"
            FAILED_TESTS+=("Intel archive download")
            ((TESTS_FAILED++))
        fi
        
        # Test ARM64 archive
        log_info "Testing ARM64 archive structure..."
        if curl -sL "$ARM64_URL" -o "$TEMP_DIR/arm64.tar.gz"; then
            if tar -tzf "$TEMP_DIR/arm64.tar.gz" | grep -q "^rnx$"; then
                log_success "ARM64 archive contains rnx binary"
                ((TESTS_PASSED++))
            else
                log_error "ARM64 archive missing rnx binary"
                FAILED_TESTS+=("ARM64 archive structure")
                ((TESTS_FAILED++))
            fi
            
            if tar -tzf "$TEMP_DIR/arm64.tar.gz" | grep -q "^admin/"; then
                log_success "ARM64 archive contains admin directory"
                ((TESTS_PASSED++))
            else
                log_error "ARM64 archive missing admin directory"
                FAILED_TESTS+=("ARM64 archive admin directory")
                ((TESTS_FAILED++))
            fi
        else
            log_error "Failed to download ARM64 archive"
            FAILED_TESTS+=("ARM64 archive download")
            ((TESTS_FAILED++))
        fi
        
        rm -rf "$TEMP_DIR"
    else
        log_warning "URLs contain placeholders, skipping archive structure validation"
    fi
    
    # Final cleanup
    cleanup_test_installations
    
    # Test results summary
    log_header "Test Results Summary"
    
    local total_tests=$((TESTS_PASSED + TESTS_FAILED))
    
    echo "📊 Test Statistics:"
    echo "   Total tests: $total_tests"
    echo "   Passed: $TESTS_PASSED"
    echo "   Failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed! 🎉"
        echo ""
        echo "✅ The RNX homebrew formula is ready for use!"
        echo ""
        echo "📋 Next steps:"
        echo "1. Push the formula to the homebrew tap repository"
        echo "2. Test installation from the tap:"
        echo "   brew tap ehsaniara/joblet"
        echo "   brew install rnx"
        echo "3. Test the auto-update workflow with a release"
        exit 0
    else
        log_error "Some tests failed!"
        echo ""
        echo "❌ Failed tests:"
        for test in "${FAILED_TESTS[@]}"; do
            echo "   - $test"
        done
        echo ""
        echo "💡 Please fix the issues above before deploying the formula"
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help|--cleanup-only]"
        echo ""
        echo "Options:"
        echo "  --help         Show this help message"
        echo "  --cleanup-only Only run cleanup without testing"
        echo ""
        echo "This script tests the RNX homebrew formula comprehensively."
        exit 0
        ;;
    --cleanup-only)
        log_info "Running cleanup only..."
        cleanup_test_installations
        log_success "Cleanup complete"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac