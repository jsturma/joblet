#!/bin/bash
set -e

# RNX User Installation Verification Script
# This script helps users verify their RNX homebrew installation

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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
    echo -e "\n${CYAN}🔍 $1${NC}"
    echo "=========================================="
}

log_tip() {
    echo -e "${CYAN}💡 $1${NC}"
}

# Verification results
CHECKS_PASSED=0
CHECKS_FAILED=0
ISSUES_FOUND=()
RECOMMENDATIONS=()

check_item() {
    local check_name="$1"
    local check_command="$2"
    local success_msg="$3"
    local failure_msg="$4"
    local is_critical="${5:-true}"
    
    if eval "$check_command" &>/dev/null; then
        log_success "$success_msg"
        ((CHECKS_PASSED++))
        return 0
    else
        if [ "$is_critical" = "true" ]; then
            log_error "$failure_msg"
            ISSUES_FOUND+=("$check_name: $failure_msg")
        else
            log_warning "$failure_msg"
            RECOMMENDATIONS+=("$check_name: $failure_msg")
        fi
        ((CHECKS_FAILED++))
        return 1
    fi
}

get_system_info() {
    echo "🖥️  System Information:"
    echo "   OS: $(sw_vers -productName) $(sw_vers -productVersion)"
    echo "   Architecture: $(uname -m)"
    echo "   Homebrew: $(brew --version | head -1)"
    echo ""
}

check_homebrew_setup() {
    log_header "Homebrew Setup"
    
    check_item "homebrew_installed" \
        "command -v brew" \
        "Homebrew is installed" \
        "Homebrew is not installed"
    
    check_item "homebrew_tap" \
        "brew tap | grep -q ehsaniara/joblet" \
        "RNX tap is added (ehsaniara/joblet)" \
        "RNX tap is not added - run: brew tap ehsaniara/joblet"
    
    # Get tap info
    if brew tap | grep -q ehsaniara/joblet; then
        local tap_path=$(brew --repository ehsaniara/joblet)
        if [ -d "$tap_path" ]; then
            echo "   Tap location: $tap_path"
            local last_updated=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$tap_path/.git/FETCH_HEAD" 2>/dev/null || echo "Unknown")
            echo "   Last updated: $last_updated"
        fi
    fi
}

check_rnx_installation() {
    log_header "RNX Installation"
    
    check_item "rnx_command" \
        "command -v rnx" \
        "RNX command is available" \
        "RNX command not found - run: brew install rnx"
    
    if command -v rnx &>/dev/null; then
        # Get version info
        local rnx_version=$(rnx --version 2>/dev/null || echo "Unknown")
        echo "   Version: $rnx_version"
        
        # Get installation path
        local rnx_path=$(which rnx)
        echo "   Location: $rnx_path"
        
        # Check if it's properly linked
        local brew_prefix=$(brew --prefix)
        if [[ "$rnx_path" == "$brew_prefix"* ]]; then
            log_success "RNX is properly installed via Homebrew"
            ((CHECKS_PASSED++))
        else
            log_warning "RNX is not in Homebrew path - may be installed elsewhere"
            RECOMMENDATIONS+=("RNX path: RNX found at $rnx_path but expected in $brew_prefix/bin")
        fi
        
        # Test basic functionality
        check_item "rnx_help" \
            "rnx --help | grep -q 'Usage:'" \
            "RNX help command works" \
            "RNX help command failed"
    fi
}

check_admin_ui_installation() {
    log_header "Admin UI Installation"
    
    local brew_prefix=$(brew --prefix)
    local admin_dir="$brew_prefix/share/rnx/admin"
    
    if [ -d "$admin_dir" ]; then
        log_success "Admin UI is installed"
        ((CHECKS_PASSED++))
        
        # Check admin UI components
        check_item "admin_server" \
            "test -f '$admin_dir/server/package.json'" \
            "Admin server files are present" \
            "Admin server files are missing"
        
        check_item "admin_ui_build" \
            "test -f '$admin_dir/ui/dist/index.html'" \
            "Admin UI build is present" \
            "Admin UI build is missing"
        
        check_item "admin_launcher" \
            "command -v rnx-admin" \
            "Admin launcher (rnx-admin) is available" \
            "Admin launcher is missing"
        
        # Check Node.js dependencies
        if [ -d "$admin_dir/server/node_modules" ]; then
            log_success "Admin server dependencies are installed"
            ((CHECKS_PASSED++))
            
            # Count installed packages
            local pkg_count=$(find "$admin_dir/server/node_modules" -maxdepth 1 -type d | wc -l)
            echo "   Installed packages: $((pkg_count - 1))"
        else
            log_error "Admin server dependencies are missing"
            ISSUES_FOUND+=("Admin dependencies: Node modules not found in $admin_dir/server")
            ((CHECKS_FAILED++))
        fi
        
        # Check Node.js version
        if command -v node &>/dev/null; then
            local node_version=$(node --version)
            local node_major=$(echo "$node_version" | sed 's/v\([0-9]*\).*/\1/')
            
            if [ "$node_major" -ge 18 ]; then
                log_success "Node.js version is compatible ($node_version)"
                ((CHECKS_PASSED++))
            else
                log_warning "Node.js version may be too old ($node_version, recommend 18+)"
                RECOMMENDATIONS+=("Node.js version: Consider upgrading to Node.js 18+ for optimal admin UI experience")
            fi
        else
            log_error "Node.js is not available"
            ISSUES_FOUND+=("Node.js: Required for admin UI but not found")
            ((CHECKS_FAILED++))
        fi
        
    else
        log_info "Admin UI is not installed (CLI-only installation)"
        echo "   To install admin UI: brew reinstall rnx --with-admin"
        
        # Check if Node.js is available for future admin UI installation
        if command -v node &>/dev/null; then
            local node_version=$(node --version)
            echo "   Node.js available: $node_version"
            log_tip "You can install the admin UI with: brew reinstall rnx --with-admin"
        else
            echo "   Node.js: Not available"
            log_tip "Install Node.js first, then: brew reinstall rnx --with-admin"
        fi
    fi
}

check_configuration() {
    log_header "Configuration"
    
    local config_dir="$HOME/.rnx"
    local config_file="$config_dir/rnx-config.yml"
    
    check_item "config_directory" \
        "test -d '$config_dir'" \
        "Configuration directory exists (~/.rnx)" \
        "Configuration directory missing - create with: mkdir -p ~/.rnx" \
        false
    
    if [ -f "$config_file" ]; then
        log_success "Configuration file exists"
        ((CHECKS_PASSED++))
        
        # Basic config validation
        if grep -q "server:" "$config_file" 2>/dev/null; then
            log_success "Configuration appears valid (contains server section)"
            ((CHECKS_PASSED++))
        else
            log_warning "Configuration may be incomplete"
            RECOMMENDATIONS+=("Config validation: Check that $config_file contains proper server configuration")
        fi
        
        # Check for certificate files mentioned in config
        if grep -q "cert" "$config_file" 2>/dev/null; then
            local cert_dir="$config_dir/certs"
            if [ -d "$cert_dir" ]; then
                log_success "Certificate directory exists"
                ((CHECKS_PASSED++))
                
                local cert_files=(client.crt client.key ca.crt)
                for cert_file in "${cert_files[@]}"; do
                    if [ -f "$cert_dir/$cert_file" ]; then
                        log_success "Certificate file exists: $cert_file"
                        ((CHECKS_PASSED++))
                    else
                        log_warning "Certificate file missing: $cert_file"
                        RECOMMENDATIONS+=("Certificate: $cert_dir/$cert_file not found")
                    fi
                done
            else
                log_warning "Certificate directory missing"
                RECOMMENDATIONS+=("Certificates: Expected certificate directory at $cert_dir")
            fi
        fi
        
    else
        log_warning "Configuration file missing"
        RECOMMENDATIONS+=("Configuration: Create $config_file with your joblet server settings")
        echo "   Copy from your joblet server: scp server:/opt/joblet/config/rnx-config.yml ~/.rnx/"
    fi
}

test_connectivity() {
    log_header "Connectivity Test"
    
    local config_file="$HOME/.rnx/rnx-config.yml"
    
    if [ -f "$config_file" ] && command -v rnx &>/dev/null; then
        log_info "Testing connection to joblet server..."
        
        # Try to run a simple command with timeout
        if timeout 10s rnx list &>/dev/null; then
            log_success "Successfully connected to joblet server"
            ((CHECKS_PASSED++))
            
            # Get job count
            local job_output=$(timeout 10s rnx list 2>/dev/null || echo "")
            if [ -n "$job_output" ]; then
                local job_count=$(echo "$job_output" | tail -n +2 | wc -l)
                echo "   Jobs visible: $job_count"
            fi
            
        else
            log_warning "Cannot connect to joblet server"
            RECOMMENDATIONS+=("Connectivity: Check that your joblet server is running and accessible")
            echo "   Possible issues:"
            echo "   - Joblet server is not running"
            echo "   - Network connectivity issues"
            echo "   - Certificate problems"
            echo "   - Configuration errors"
        fi
    else
        log_info "Skipping connectivity test (missing config or RNX not installed)"
    fi
}

test_admin_ui() {
    log_header "Admin UI Test"
    
    if command -v rnx-admin &>/dev/null; then
        log_info "Admin UI launcher is available"
        
        # Check if we can start the admin UI (don't actually start it)
        local brew_prefix=$(brew --prefix)
        local admin_server="$brew_prefix/share/rnx/admin/server"
        
        if [ -f "$admin_server/server.js" ]; then
            log_success "Admin server script exists"
            ((CHECKS_PASSED++))
            
            # Test Node.js can load the server script (syntax check)
            if cd "$admin_server" && node -c server.js 2>/dev/null; then
                log_success "Admin server script is valid"
                ((CHECKS_PASSED++))
            else
                log_error "Admin server script has syntax errors"
                ISSUES_FOUND+=("Admin server: server.js has syntax errors")
                ((CHECKS_FAILED++))
            fi
            
        else
            log_error "Admin server script is missing"
            ISSUES_FOUND+=("Admin server: server.js not found at $admin_server")
            ((CHECKS_FAILED++))
        fi
        
        log_tip "To test admin UI: run 'rnx admin' and check http://localhost:5173"
        
    else
        log_info "Admin UI is not installed"
        log_tip "Install with: brew reinstall rnx --with-admin"
    fi
}

show_summary() {
    log_header "Verification Summary"
    
    local total_checks=$((CHECKS_PASSED + CHECKS_FAILED))
    
    echo "📊 Overall Results:"
    echo "   Total checks: $total_checks"
    echo "   Passed: $CHECKS_PASSED"
    echo "   Issues found: ${#ISSUES_FOUND[@]}"
    echo "   Recommendations: ${#RECOMMENDATIONS[@]}"
    echo ""
    
    if [ ${#ISSUES_FOUND[@]} -eq 0 ]; then
        log_success "No critical issues found! 🎉"
        echo ""
        echo "✅ Your RNX installation appears to be working correctly."
        
        if command -v rnx &>/dev/null; then
            echo ""
            echo "🚀 Quick start:"
            echo "   rnx --help                    # View available commands"
            echo "   rnx list                      # List jobs on server"
            echo "   rnx run \"echo hello\"          # Run a simple job"
            
            if command -v rnx-admin &>/dev/null; then
                echo "   rnx admin                     # Launch web interface"
            fi
        fi
        
    else
        log_error "Found ${#ISSUES_FOUND[@]} critical issue(s)"
        echo ""
        echo "❌ Critical Issues:"
        for issue in "${ISSUES_FOUND[@]}"; do
            echo "   • $issue"
        done
    fi
    
    if [ ${#RECOMMENDATIONS[@]} -gt 0 ]; then
        echo ""
        echo "💡 Recommendations:"
        for rec in "${RECOMMENDATIONS[@]}"; do
            echo "   • $rec"
        done
    fi
    
    echo ""
    echo "🔗 For more help:"
    echo "   • Documentation: https://github.com/ehsaniara/joblet/docs"
    echo "   • Troubleshooting: https://github.com/ehsaniara/homebrew-joblet/docs/TROUBLESHOOTING.md"
    echo "   • Issues: https://github.com/ehsaniara/joblet/issues"
}

generate_report() {
    local report_file="$HOME/rnx-verification-report.txt"
    
    {
        echo "RNX Installation Verification Report"
        echo "Generated on: $(date)"
        echo "======================================"
        echo ""
        
        get_system_info
        echo ""
        
        echo "Verification Results:"
        echo "- Total checks: $((CHECKS_PASSED + CHECKS_FAILED))"
        echo "- Passed: $CHECKS_PASSED"
        echo "- Issues: ${#ISSUES_FOUND[@]}"
        echo "- Recommendations: ${#RECOMMENDATIONS[@]}"
        echo ""
        
        if [ ${#ISSUES_FOUND[@]} -gt 0 ]; then
            echo "Critical Issues:"
            for issue in "${ISSUES_FOUND[@]}"; do
                echo "• $issue"
            done
            echo ""
        fi
        
        if [ ${#RECOMMENDATIONS[@]} -gt 0 ]; then
            echo "Recommendations:"
            for rec in "${RECOMMENDATIONS[@]}"; do
                echo "• $rec"
            done
            echo ""
        fi
        
        echo "Command Outputs:"
        echo "=================="
        
        if command -v rnx &>/dev/null; then
            echo ""
            echo "rnx --version:"
            rnx --version 2>&1 || echo "Failed to get version"
            
            echo ""
            echo "rnx --help:"
            rnx --help 2>&1 | head -20 || echo "Failed to get help"
        fi
        
        echo ""
        echo "brew list rnx:"
        brew list rnx 2>&1 | head -20 || echo "RNX not installed via Homebrew"
        
        if [ -f "$HOME/.rnx/rnx-config.yml" ]; then
            echo ""
            echo "Configuration file (sensitive data removed):"
            sed 's/\(password\|key\|secret\|token\):.*$/\1: [REDACTED]/' "$HOME/.rnx/rnx-config.yml" || echo "Could not read config"
        fi
        
    } > "$report_file"
    
    log_info "Detailed report saved to: $report_file"
}

main() {
    echo "🔍 RNX Installation Verification"
    echo "=================================="
    echo ""
    
    # Check if running on macOS
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "This script is for macOS only"
        exit 1
    fi
    
    get_system_info
    
    # Run all verification checks
    check_homebrew_setup
    check_rnx_installation
    check_admin_ui_installation
    check_configuration
    test_connectivity
    test_admin_ui
    
    # Show results
    show_summary
    
    # Generate detailed report
    if [ "$1" = "--report" ]; then
        generate_report
    fi
    
    # Exit with appropriate code
    if [ ${#ISSUES_FOUND[@]} -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help|--report]"
        echo ""
        echo "This script verifies your RNX homebrew installation."
        echo ""
        echo "Options:"
        echo "  --help    Show this help message"
        echo "  --report  Generate a detailed report file"
        echo ""
        echo "The script checks:"
        echo "• Homebrew and tap setup"
        echo "• RNX CLI installation"
        echo "• Admin UI installation (if present)"
        echo "• Configuration files"
        echo "• Basic connectivity"
        exit 0
        ;;
    --report|"")
        main "$@"
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac