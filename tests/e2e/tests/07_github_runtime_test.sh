#!/bin/bash

# Test 07: GitHub Runtime Installation Tests  
# Tests GitHub repository-based runtime installation features

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# Test configuration
GITHUB_REPO="ehsaniara/joblet/tree/main/runtimes"
PYTHON_RUNTIME="python-3.11-ml"
JAVA_RUNTIME="openjdk-21"
TEST_RUNTIME_INSTALLED=false

# ============================================
# Helper Functions
# ============================================

cleanup_test_runtime() {
    if [[ "$TEST_RUNTIME_INSTALLED" == "true" ]]; then
        echo -e "  ${YELLOW}Cleaning up test runtime...${NC}"
        "$RNX_BINARY" runtime remove "$PYTHON_RUNTIME" >/dev/null 2>&1 || true
        TEST_RUNTIME_INSTALLED=false
    fi
}

# Trap to ensure cleanup on exit
trap cleanup_test_runtime EXIT

# ============================================
# Test Functions
# ============================================

test_github_runtime_list() {
    echo -e "  ${BLUE}Testing GitHub repository runtime listing...${NC}"
    
    local output=$("$RNX_BINARY" runtime list --github-repo="$GITHUB_REPO" 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        assert_contains "$output" "$PYTHON_RUNTIME" "Should list Python ML runtime"
        assert_contains "$output" "$JAVA_RUNTIME" "Should list Java runtime" 
        return 0
    else
        echo -e "    ${RED}Failed to list GitHub runtimes: $output${NC}"
        return 1
    fi
}

test_github_runtime_list_json() {
    echo -e "  ${BLUE}Testing GitHub repository runtime listing with JSON output...${NC}"
    
    local output=$("$RNX_BINARY" runtime list --github-repo="$GITHUB_REPO" --json 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        # Verify JSON structure
        if echo "$output" | jq . >/dev/null 2>&1; then
            assert_contains "$output" "\"$PYTHON_RUNTIME\"" "JSON should contain Python runtime"
            assert_contains "$output" "\"$JAVA_RUNTIME\"" "JSON should contain Java runtime"
            return 0
        else
            echo -e "    ${RED}Output is not valid JSON${NC}"
            return 1
        fi
    else
        echo -e "    ${RED}Failed to get JSON runtime list: $output${NC}"
        return 1
    fi
}

test_github_runtime_install() {
    echo -e "  ${BLUE}Testing GitHub repository runtime installation...${NC}"
    
    # Ensure runtime is not already installed
    cleanup_test_runtime
    
    # Install from GitHub repository
    local output=$(timeout 300 "$RNX_BINARY" runtime install "$PYTHON_RUNTIME" --github-repo="$GITHUB_REPO" 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        TEST_RUNTIME_INSTALLED=true
        
        # Verify installation success messages
        assert_contains "$output" "Installing runtime: $PYTHON_RUNTIME" "Should show installation message"
        assert_contains "$output" "Installing from GitHub repository" "Should show GitHub source"
        assert_contains "$output" "GitHub runtime installation started" "Should start installation"
        
        # Verify runtime is now available locally
        if runtime_exists "$PYTHON_RUNTIME"; then
            return 0
        else
            echo -e "    ${RED}Runtime not found after installation${NC}"
            return 1
        fi
    else
        echo -e "    ${RED}Failed to install runtime from GitHub: $output${NC}"
        return 1
    fi
}

test_github_manifest_validation() {
    echo -e "  ${BLUE}Testing GitHub manifest system validation...${NC}"
    
    local output=$("$RNX_BINARY" runtime install "$PYTHON_RUNTIME" --github-repo="$GITHUB_REPO" 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        # Should show GitHub installation process
        assert_contains "$output" "Installing from GitHub repository" "Should show GitHub installation"
        assert_contains "$output" "Repository:" "Should show repository info"
        assert_contains "$output" "Branch:" "Should show branch info"
        assert_contains "$output" "GitHub runtime installation started" "Should start installation"
        return 0
    else
        echo -e "    ${RED}GitHub installation failed: $output${NC}"
        return 1
    fi
}

test_github_archive_download() {
    echo -e "  ${BLUE}Testing GitHub archive download process...${NC}"
    
    local output=$("$RNX_BINARY" runtime install "$PYTHON_RUNTIME" --github-repo="$GITHUB_REPO" 2>&1)
    
    # Should show build process (GitHub download happens in background job)
    assert_contains "$output" "Build Job:" "Should show build job ID"
    assert_contains "$output" "Status: building" "Should show building status"
    assert_contains "$output" "Runtime build started" "Should start build process"
    assert_contains "$output" "real-time log streaming" "Should provide log streaming"
}

test_runtime_execution_after_github_install() {
    echo -e "  ${BLUE}Testing runtime execution after GitHub installation...${NC}"
    
    # Ensure runtime is installed
    if ! runtime_exists "$PYTHON_RUNTIME"; then
        echo -e "    ${YELLOW}Runtime not installed, installing first...${NC}"
        if ! test_github_runtime_install; then
            return 1
        fi
    fi
    
    # Test Python execution
    local job_id=$(run_python_job "print('GITHUB_RUNTIME_SUCCESS')")
    local logs=$(get_job_logs "$job_id")
    
    assert_contains "$logs" "GITHUB_RUNTIME_SUCCESS" "Python should execute after GitHub install"
}

test_host_contamination_prevention() {
    echo -e "  ${BLUE}Testing host system contamination prevention...${NC}"
    
    # Get pip package count before
    local pip_before=$(pip3 list 2>/dev/null | wc -l || echo "0")
    
    # Install runtime from GitHub
    "$RNX_BINARY" runtime install "$PYTHON_RUNTIME" --github-repo="$GITHUB_REPO" >/dev/null 2>&1
    TEST_RUNTIME_INSTALLED=true
    
    # Get pip package count after
    local pip_after=$(pip3 list 2>/dev/null | wc -l || echo "0")
    
    # Host pip packages should not change
    if [[ "$pip_before" -eq "$pip_after" ]]; then
        echo -e "    ${GREEN}✓ Host pip packages unchanged ($pip_before -> $pip_after)${NC}"
        return 0
    else
        echo -e "    ${RED}✗ Host contamination detected: pip packages changed ($pip_before -> $pip_after)${NC}"
        return 1
    fi
}

test_invalid_github_repo() {
    echo -e "  ${BLUE}Testing invalid GitHub repository handling...${NC}"
    
    local output=$("$RNX_BINARY" runtime list --github-repo="invalid/nonexistent" 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        # Should handle gracefully with error message
        assert_contains "$output" "ERROR" "Should show error for invalid repo"
        return 0
    else
        echo -e "    ${RED}Should have failed for invalid repository${NC}"
        return 1
    fi
}

test_runtime_isolation_verification() {
    echo -e "  ${BLUE}Testing runtime isolation after GitHub installation...${NC}"
    
    # Ensure runtime is installed
    if ! runtime_exists "$PYTHON_RUNTIME"; then
        if ! test_github_runtime_install; then
            return 1
        fi
    fi
    
    # Test that runtime files are in correct isolation location
    local runtime_path="/opt/joblet/runtimes/$PYTHON_RUNTIME"
    
    # Check runtime configuration exists
    if [[ -f "$runtime_path/runtime.yml" ]]; then
        echo -e "    ${GREEN}✓ Runtime configuration found at $runtime_path/runtime.yml${NC}"
    else
        echo -e "    ${RED}✗ Runtime configuration not found${NC}"
        return 1
    fi
    
    # Check isolated directory structure
    if [[ -d "$runtime_path/isolated" ]]; then
        echo -e "    ${GREEN}✓ Isolated directory structure created${NC}"
        return 0
    else
        echo -e "    ${RED}✗ Isolated directory structure not found${NC}"
        return 1
    fi
}

test_concurrent_github_installs() {
    echo -e "  ${BLUE}Testing concurrent GitHub runtime installations...${NC}"
    
    # Clean up first
    "$RNX_BINARY" runtime remove "$PYTHON_RUNTIME" >/dev/null 2>&1 || true
    "$RNX_BINARY" runtime remove "$JAVA_RUNTIME" >/dev/null 2>&1 || true
    
    # Start concurrent installations
    "$RNX_BINARY" runtime install "$PYTHON_RUNTIME" --github-repo="$GITHUB_REPO" >/dev/null 2>&1 &
    local python_pid=$!
    
    "$RNX_BINARY" runtime install "$JAVA_RUNTIME" --github-repo="$GITHUB_REPO" >/dev/null 2>&1 &
    local java_pid=$!
    
    # Wait for both to complete
    wait $python_pid
    local python_exit=$?
    wait $java_pid  
    local java_exit=$?
    
    # Check results
    if [[ $python_exit -eq 0 && $java_exit -eq 0 ]]; then
        if runtime_exists "$PYTHON_RUNTIME" && runtime_exists "$JAVA_RUNTIME"; then
            echo -e "    ${GREEN}✓ Both runtimes installed successfully${NC}"
            
            # Clean up
            "$RNX_BINARY" runtime remove "$PYTHON_RUNTIME" >/dev/null 2>&1 || true
            "$RNX_BINARY" runtime remove "$JAVA_RUNTIME" >/dev/null 2>&1 || true
            
            return 0
        else
            echo -e "    ${RED}✗ Runtimes not found after concurrent install${NC}"
            return 1
        fi
    else
        echo -e "    ${RED}✗ Concurrent installations failed (python: $python_exit, java: $java_exit)${NC}"
        return 1
    fi
}

# ============================================
# Prerequisites Check
# ============================================

check_github_prerequisites() {
    echo -e "${BLUE}Checking GitHub runtime test prerequisites...${NC}"
    
    # Check basic requirements using framework function
    local prereqs_met=true
    
    # Check RNX binary
    if [[ ! -x "$RNX_BINARY" ]]; then
        echo -e "${RED}Error: RNX binary not found or not executable: $RNX_BINARY${NC}"
        prereqs_met=false
    fi
    
    # Check joblet service
    if ! "$RNX_BINARY" runtime list &>/dev/null; then
        echo -e "${RED}Error: Cannot connect to joblet service${NC}"
        prereqs_met=false
    fi
    
    # Check network connectivity to GitHub
    if ! timeout 10 bash -c "wget --spider --quiet https://github.com" >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Cannot reach GitHub - some tests may be skipped${NC}"
        prereqs_met=false
    fi
    
    # Check if jq is available for JSON tests
    if ! command -v jq >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: jq not available - JSON tests will be skipped${NC}"
    fi
    
    if [[ "$prereqs_met" == "false" ]]; then
        return 1
    fi
    
    echo -e "${GREEN}✓ Prerequisites satisfied${NC}"
    return 0
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "GitHub Runtime Installation Tests"
    
    # Check prerequisites
    if ! check_github_prerequisites; then
        echo -e "${RED}Prerequisites check failed - skipping GitHub runtime tests${NC}"
        exit 0  # Skip rather than fail
    fi
    
    # Run tests
    test_section "GitHub Repository Runtime Listing"
    run_test "GitHub runtime list" test_github_runtime_list
    
    if command -v jq >/dev/null 2>&1; then
        run_test "GitHub runtime list (JSON)" test_github_runtime_list_json
    else
        skip_test "GitHub runtime list (JSON)" "jq not available"
    fi
    
    test_section "GitHub Repository Runtime Installation"
    run_test "GitHub runtime installation" test_github_runtime_install
    run_test "GitHub manifest validation" test_github_manifest_validation
    run_test "GitHub archive download" test_github_archive_download
    
    test_section "Runtime Execution After GitHub Install"
    run_test "Runtime execution after GitHub install" test_runtime_execution_after_github_install
    
    test_section "Security and Isolation"
    run_test "Host contamination prevention" test_host_contamination_prevention
    run_test "Runtime isolation verification" test_runtime_isolation_verification
    
    test_section "Error Handling"
    run_test "Invalid GitHub repository handling" test_invalid_github_repo
    
    test_section "Concurrency"
    run_test "Concurrent GitHub installations" test_concurrent_github_installs
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi