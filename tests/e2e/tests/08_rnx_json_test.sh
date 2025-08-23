#!/bin/bash

# Test 08: RNX JSON Flag Tests
# Tests the --json flag functionality for key rnx commands to ensure frontend compatibility

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# ============================================
# RNX JSON Test Helper Functions
# ============================================

# Basic JSON validation without jq dependency
validate_basic_json() {
    local json_output="$1"
    
    # Basic checks for JSON structure (object or array)
    if [[ "$json_output" =~ ^[[:space:]]*\{ && "$json_output" =~ \}[[:space:]]*$ ]]; then
        # JSON object
        return 0
    elif [[ "$json_output" =~ ^[[:space:]]*\[ && "$json_output" =~ \][[:space:]]*$ ]]; then
        # JSON array
        return 0
    else
        echo "    ${RED}Output doesn't appear to be JSON format${NC}"
        echo "    ${RED}Raw output: $json_output${NC}"
        return 1
    fi
}

# Check if JSON contains expected fields using grep
check_json_fields() {
    local json_output="$1"
    local required_fields="$2"  # space-separated list
    
    for field in $required_fields; do
        if ! echo "$json_output" | grep -q "\"$field\""; then
            echo "    ${RED}Missing required field: $field${NC}"
            return 1
        fi
    done
    
    return 0
}

# Execute rnx command with --json flag and validate
execute_rnx_json() {
    local command="$1"
    local expected_fields="$2"
    local timeout="${3:-10}"
    
    local json_output
    # Handle global --json flag (for run command) vs regular --json flag
    if [[ "$command" == "--json "* ]]; then
        json_output=$(timeout "$timeout" "$RNX_BINARY" $command 2>&1)
    else
        json_output=$(timeout "$timeout" "$RNX_BINARY" --json $command 2>&1)
    fi
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        echo "    ${RED}Command failed with exit code: $exit_code${NC}"
        echo "    ${RED}Output: $json_output${NC}"
        return 1
    fi
    
    if validate_basic_json "$json_output" && check_json_fields "$json_output" "$expected_fields"; then
        echo "$json_output"
        return 0
    fi
    
    return 1
}

# Extract simple JSON values using grep/sed (basic implementation)
extract_json_value() {
    local json_output="$1"
    local field="$2"
    
    echo "$json_output" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)"/\1/' | head -1
}

# ============================================
# Test Functions
# ============================================

test_rnx_runtime_list_json() {
    local json_output
    json_output=$(execute_rnx_json "runtime list" "name")
    
    if [[ $? -eq 0 ]]; then
        # Check that it looks like a JSON array
        if echo "$json_output" | grep -q '^\s*\[' && echo "$json_output" | grep -q '\]\s*$'; then
            # Check for runtime structure if any runtimes exist
            if echo "$json_output" | grep -q '"name"'; then
                # Check for expected runtime fields
                local required_runtime_fields="name language version available"
                for field in $required_runtime_fields; do
                    if ! echo "$json_output" | grep -q "\"$field\""; then
                        echo "    ${RED}Runtime object missing field: $field${NC}"
                        return 1
                    fi
                done
            fi
            return 0
        fi
    fi
    
    return 1
}

test_rnx_run_json() {
    local json_output
    json_output=$(execute_rnx_json "--json run echo 'rnx-json-test'" "job_uuid status" 15)
    
    if [[ $? -eq 0 ]]; then
        # Validate job_uuid format (UUID-like) using basic pattern matching
        local job_uuid
        job_uuid=$(extract_json_value "$json_output" "job_uuid")
        
        if [[ "$job_uuid" =~ ^[a-f0-9-]{36}$ ]]; then
            # Check that status field exists
            if echo "$json_output" | grep -q "status"; then
                return 0
            else
                echo "    ${RED}Status field not found${NC}"
                return 1
            fi
        else
            echo "    ${RED}Invalid job_uuid format: $job_uuid${NC}"
            return 1
        fi
    fi
    
    return 1
}

test_rnx_list_json() {
    local json_output
    json_output=$(execute_rnx_json "list" "id")
    
    if [[ $? -eq 0 ]]; then
        # Check that it looks like a JSON array
        if echo "$json_output" | grep -q '^\s*\[' && echo "$json_output" | grep -q '\]\s*$'; then
            # If jobs exist, check for basic job fields
            if echo "$json_output" | grep -q '"id"'; then
                local required_job_fields="id command status"
                for field in $required_job_fields; do
                    if ! echo "$json_output" | grep -q "\"$field\""; then
                        echo "    ${RED}Job object missing field: $field${NC}"
                        return 1
                    fi
                done
            fi
            return 0
        fi
    fi
    
    return 1
}

# Test that non-JSON commands still work normally (regression test)
test_rnx_normal_output() {
    # Test that commands without --json still produce normal output
    local normal_output
    normal_output=$("$RNX_BINARY" runtime list 2>/dev/null | head -5)
    
    if [[ $? -eq 0 ]]; then
        # Normal output should NOT be JSON
        if echo "$normal_output" | grep -qE "^\s*[\[\{]"; then
            echo "    ${RED}Normal output is unexpectedly JSON-like${NC}"
            return 1
        fi
        
        # Should contain some expected text patterns (basic check)
        if echo "$normal_output" | grep -qE "(openjdk|python|Runtime|Available|Name)" || [[ ${#normal_output} -gt 10 ]]; then
            return 0
        else
            echo "    ${RED}Normal output seems too short or missing expected patterns${NC}"
            return 1
        fi
    fi
    
    return 1
}

# ============================================
# Test Suite Execution
# ============================================

main() {
    test_suite_init "RNX JSON Flag Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites not met. Exiting.${NC}"
        exit 1
    fi
    
    # Note: This test uses basic shell tools for JSON validation to avoid external dependencies
    
    test_section "RNX Core Commands with --json"
    run_test "rnx runtime list --json" test_rnx_runtime_list_json
    run_test "rnx run --json" test_rnx_run_json
    run_test "rnx list --json" test_rnx_list_json
    
    test_section "RNX Regression Tests"
    run_test "Normal output still works" test_rnx_normal_output
    
    # Clean up test artifacts
    cleanup_test_artifacts
    
    # Print summary and exit
    test_suite_summary
    exit $?
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi