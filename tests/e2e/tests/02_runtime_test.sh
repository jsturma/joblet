#!/bin/bash

# Test 02: Runtime Management Tests
# Tests runtime installation, listing, info, and job execution

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# Test configuration
PYTHON_RUNTIME="python-3.11-ml"
JAVA_RUNTIME="openjdk-21"

# ============================================
# Test Functions
# ============================================

test_runtime_list() {
    local output=$("$RNX_BINARY" runtime list 2>&1)
    
    if [[ $? -eq 0 ]]; then
        return 0
    else
        echo -e "    ${RED}Failed to list runtimes: $output${NC}"
        return 1
    fi
}

test_python_runtime_available() {
    if runtime_exists "$PYTHON_RUNTIME"; then
        return 0
    else
        echo -e "    ${YELLOW}Runtime not found, installing...${NC}"
        # Run from project root to find runtime sources
        if timeout 300 bash -c "cd '$JOBLET_ROOT' && '$RNX_BINARY' runtime install '$PYTHON_RUNTIME'" >/dev/null 2>&1; then
            return 0
        else
            return 1
        fi
    fi
}

test_runtime_info() {
    local info=$(get_runtime_info "$PYTHON_RUNTIME")
    
    assert_contains "$info" "Runtime:" "Should show runtime name"
    assert_contains "$info" "Version:" "Should show version"
    assert_contains "$info" "Description:" "Should show description"
}

test_python_execution() {
    local job_id=$(run_python_job "print('PYTHON_EXEC_SUCCESS')")
    local logs=$(get_job_logs "$job_id")
    
    assert_contains "$logs" "PYTHON_EXEC_SUCCESS" "Python should execute"
}

test_python_imports() {
    local job_id=$(run_python_job "
import sys
import os
import json
print('IMPORTS_OK')
print(f'Python: {sys.version}')
")
    local logs=$(get_job_logs "$job_id")
    
    assert_contains "$logs" "IMPORTS_OK" "Should import standard libraries"
    assert_contains "$logs" "Python:" "Should show Python version"
}

test_python_ml_packages() {
    local job_id=$(run_python_job "
try:
    import numpy
    print('NUMPY_AVAILABLE')
except ImportError:
    print('NUMPY_NOT_AVAILABLE')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    # ML packages may or may not be available, just check execution works
    if echo "$clean_output" | grep -q "AVAILABLE"; then
        return 0
    else
        return 1
    fi
}

test_concurrent_python_jobs() {
    # Start multiple jobs
    local job1=$(run_python_job "import time; print('JOB1_START'); time.sleep(2); print('JOB1_END')")
    local job2=$(run_python_job "import time; print('JOB2_START'); time.sleep(2); print('JOB2_END')")
    local job3=$(run_python_job "import time; print('JOB3_START'); time.sleep(2); print('JOB3_END')")
    
    # Wait for completion
    sleep 5
    
    # Check all completed
    local status1=$(check_job_status "$job1")
    local status2=$(check_job_status "$job2")
    local status3=$(check_job_status "$job3")
    
    if [[ "$status1" == "COMPLETED" ]] && [[ "$status2" == "COMPLETED" ]] && [[ "$status3" == "COMPLETED" ]]; then
        return 0
    else
        echo -e "    ${RED}Job statuses: $status1, $status2, $status3${NC}"
        return 1
    fi
}

test_runtime_persistence() {
    # Check runtime still exists after tests
    runtime_exists "$PYTHON_RUNTIME"
}

test_runtime_lifecycle() {
    echo -e "    ${BLUE}Testing complete runtime lifecycle (remove/install/test)${NC}"
    
    # Step 1: Remove both runtimes from remote host
    echo -e "    ${BLUE}Step 1: Removing both runtimes from remote host...${NC}"
    
    # Remove Java runtime if it exists
    if runtime_exists "$JAVA_RUNTIME"; then
        if ! "$RNX_BINARY" runtime remove "$JAVA_RUNTIME" >/dev/null 2>&1; then
            echo -e "    ${RED}Failed to remove $JAVA_RUNTIME${NC}"
            return 1
        fi
        echo -e "    ${GREEN}✓ Removed $JAVA_RUNTIME${NC}"
    else
        echo -e "    ${YELLOW}$JAVA_RUNTIME was not installed${NC}"
    fi
    
    # Remove Python runtime if it exists
    if runtime_exists "$PYTHON_RUNTIME"; then
        if ! "$RNX_BINARY" runtime remove "$PYTHON_RUNTIME" >/dev/null 2>&1; then
            echo -e "    ${RED}Failed to remove $PYTHON_RUNTIME${NC}"
            return 1
        fi
        echo -e "    ${GREEN}✓ Removed $PYTHON_RUNTIME${NC}"
    else
        echo -e "    ${YELLOW}$PYTHON_RUNTIME was not installed${NC}"
    fi
    
    # Verify both runtimes are removed
    local runtime_list=$("$RNX_BINARY" runtime list 2>&1)
    if echo "$runtime_list" | grep -q "$JAVA_RUNTIME\|$PYTHON_RUNTIME"; then
        echo -e "    ${RED}Some runtimes still listed after removal${NC}"
        return 1
    fi
    echo -e "    ${GREEN}✓ Both runtimes successfully removed from remote host${NC}"
    
    # Step 2: Install Java runtime individually
    echo -e "    ${BLUE}Step 2: Installing $JAVA_RUNTIME individually...${NC}"
    if ! timeout 300 bash -c "cd '$JOBLET_ROOT' && '$RNX_BINARY' runtime install '$JAVA_RUNTIME'" >/dev/null 2>&1; then
        echo -e "    ${RED}Failed to install $JAVA_RUNTIME${NC}"
        return 1
    fi
    
    # Verify Java runtime is installed
    if ! runtime_exists "$JAVA_RUNTIME"; then
        echo -e "    ${RED}$JAVA_RUNTIME not listed after installation${NC}"
        return 1
    fi
    echo -e "    ${GREEN}✓ Successfully installed $JAVA_RUNTIME${NC}"
    
    # Step 3: Test Java runtime with actual job
    echo -e "    ${BLUE}Step 3: Testing $JAVA_RUNTIME with actual job execution...${NC}"
    local java_job_output
    if java_job_output=$("$RNX_BINARY" run --runtime="$JAVA_RUNTIME" java -version 2>&1); then
        local java_job_id=$(echo "$java_job_output" | grep "ID:" | awk '{print $2}')
        if [[ -n "$java_job_id" ]]; then
            sleep 5  # Wait for job completion
            local java_logs=$(get_job_logs "$java_job_id")
            local java_status=$(check_job_status "$java_job_id")
            
            if [[ "$java_status" == "COMPLETED" ]] && echo "$java_logs" | grep -q "openjdk\|java"; then
                echo -e "    ${GREEN}✓ Java runtime job executed successfully${NC}"
            else
                echo -e "    ${RED}Java runtime job failed - Status: $java_status${NC}"
                echo -e "    ${RED}Java job logs: $java_logs${NC}"
                return 1
            fi
        else
            echo -e "    ${RED}Failed to extract Java job ID${NC}"
            return 1
        fi
    else
        echo -e "    ${RED}Failed to submit Java job${NC}"
        return 1
    fi
    
    # Step 4: Install Python runtime individually
    echo -e "    ${BLUE}Step 4: Installing $PYTHON_RUNTIME individually...${NC}"
    if ! timeout 300 bash -c "cd '$JOBLET_ROOT' && '$RNX_BINARY' runtime install '$PYTHON_RUNTIME'" >/dev/null 2>&1; then
        echo -e "    ${RED}Failed to install $PYTHON_RUNTIME${NC}"
        return 1
    fi
    
    # Verify Python runtime is installed
    if ! runtime_exists "$PYTHON_RUNTIME"; then
        echo -e "    ${RED}$PYTHON_RUNTIME not listed after installation${NC}"
        return 1
    fi
    echo -e "    ${GREEN}✓ Successfully installed $PYTHON_RUNTIME${NC}"
    
    # Step 5: Test Python runtime with actual job
    echo -e "    ${BLUE}Step 5: Testing $PYTHON_RUNTIME with actual job execution...${NC}"
    local python_job_output
    if python_job_output=$("$RNX_BINARY" run --runtime="$PYTHON_RUNTIME" python3 --version 2>&1); then
        local python_job_id=$(echo "$python_job_output" | grep "ID:" | awk '{print $2}')
        if [[ -n "$python_job_id" ]]; then
            sleep 5  # Wait for job completion
            local python_logs=$(get_job_logs "$python_job_id")
            local python_status=$(check_job_status "$python_job_id")
            
            if [[ "$python_status" == "COMPLETED" ]] && echo "$python_logs" | grep -q "Python"; then
                echo -e "    ${GREEN}✓ Python runtime job executed successfully${NC}"
            else
                echo -e "    ${RED}Python runtime job failed - Status: $python_status${NC}"
                echo -e "    ${RED}Python job logs: $python_logs${NC}"
                return 1
            fi
        else
            echo -e "    ${RED}Failed to extract Python job ID${NC}"
            return 1
        fi
    else
        echo -e "    ${RED}Failed to submit Python job${NC}"
        return 1
    fi
    
    # Step 6: Test both runtimes are working
    echo -e "    ${BLUE}Step 6: Verifying both runtimes are working...${NC}"
    local final_runtime_list=$("$RNX_BINARY" runtime list 2>&1)
    if ! echo "$final_runtime_list" | grep -q "$JAVA_RUNTIME"; then
        echo -e "    ${RED}$JAVA_RUNTIME missing from final runtime list${NC}"
        return 1
    fi
    if ! echo "$final_runtime_list" | grep -q "$PYTHON_RUNTIME"; then
        echo -e "    ${RED}$PYTHON_RUNTIME missing from final runtime list${NC}"
        return 1
    fi
    echo -e "    ${GREEN}✓ Both runtimes are installed and working correctly${NC}"
    
    return 0
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "Runtime Management Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites check failed${NC}"
        exit 1
    fi
    
    # Run tests
    test_section "Runtime Operations"
    run_test "Runtime list command" test_runtime_list
    run_test "Python runtime availability" test_python_runtime_available
    run_test "Runtime info retrieval" test_runtime_info
    
    test_section "Python Execution"
    run_test "Basic Python execution" test_python_execution
    run_test "Python standard imports" test_python_imports
    run_test "Python ML packages check" test_python_ml_packages
    
    test_section "Concurrent Execution"
    run_test "Concurrent Python jobs" test_concurrent_python_jobs
    
    test_section "Runtime Persistence"
    run_test "Runtime remains available" test_runtime_persistence
    
    test_section "Full Runtime Lifecycle"
    run_test "Complete runtime lifecycle (install/remove/reinstall)" test_runtime_lifecycle
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi