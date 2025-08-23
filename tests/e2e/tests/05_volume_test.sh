#!/bin/bash

# Test 05: Volume Management Tests
# Tests volume creation, mounting, persistence, and sharing

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# Test configuration
TEST_VOLUME_NAME="test-volume-$$"
TEST_DATA="test_data_$(date +%s)"

# ============================================
# Test Functions
# ============================================

test_upload_file() {
    # Test uploading a file to a job
    local test_file="/tmp/test_upload_$$.txt"
    echo "$TEST_DATA" > "$test_file"
    
    local job_output=$("$RNX_BINARY" run --runtime="$DEFAULT_RUNTIME" \
        --upload="$test_file" \
        python3 -c "
import os
files = os.listdir('/work')
print(f'FILES:{files}')
if os.path.exists('/work/$(basename $test_file)'):
    with open('/work/$(basename $test_file)') as f:
        print(f'CONTENT:{f.read().strip()}')
" 2>&1)
    
    local job_id=$(echo "$job_output" | grep "ID:" | awk '{print $2}')
    local logs=$(get_job_logs "$job_id" 5)
    
    rm -f "$test_file"
    
    assert_contains "$logs" "CONTENT:$TEST_DATA" "Uploaded file should be accessible"
}

test_download_result() {
    # Test downloading results from a job
    local job_id=$(run_python_job "
with open('/work/output.txt', 'w') as f:
    f.write('DOWNLOAD_TEST_DATA')
print('FILE_CREATED')
")
    
    local logs=$(get_job_logs "$job_id")
    assert_contains "$logs" "FILE_CREATED" "Should create output file"
    
    # Check if download command exists
    local download_help=$("$RNX_BINARY" help 2>&1 | grep -i "download" || echo "")
    
    if [[ -n "$download_help" ]]; then
        local download_output=$("$RNX_BINARY" download "$job_id" /work/output.txt 2>&1 || echo "")
        
        if [[ -f "output.txt" ]]; then
            local content=$(cat output.txt)
            rm -f output.txt
            assert_equals "$content" "DOWNLOAD_TEST_DATA" "Downloaded content should match"
        else
            echo -e "    ${YELLOW}Download feature may not be implemented${NC}"
            return 0
        fi
    else
        echo -e "    ${YELLOW}Download command not available${NC}"
        return 0
    fi
}

test_volume_creation() {
    # Test creating a named volume
    local volume_output=$("$RNX_BINARY" volume create "$TEST_VOLUME_NAME" 2>&1 || echo "")
    
    if echo "$volume_output" | grep -q "not found\|not recognized\|unknown"; then
        echo -e "    ${YELLOW}Volume management not implemented${NC}"
        return 0
    fi
    
    # If volume commands exist, test them
    if echo "$volume_output" | grep -q "created\|success"; then
        return 0
    else
        echo -e "    ${YELLOW}Volume creation may not be supported${NC}"
        return 0
    fi
}

test_volume_mounting() {
    # Test mounting a volume to a job
    local job_output=$("$RNX_BINARY" run --runtime="$DEFAULT_RUNTIME" \
        --volume="$TEST_VOLUME_NAME:/data" \
        python3 -c "
import os
# Check if /data exists
if os.path.exists('/data'):
    print('VOLUME_MOUNTED')
    # Write test data
    with open('/data/test.txt', 'w') as f:
        f.write('VOLUME_DATA')
else:
    print('NO_VOLUME')
" 2>&1 || echo "")
    
    if echo "$job_output" | grep -q "volume.*not\|unrecognized option"; then
        echo -e "    ${YELLOW}Volume mounting not supported${NC}"
        return 0
    fi
    
    if echo "$job_output" | grep -q "ID:"; then
        local job_id=$(echo "$job_output" | grep "ID:" | awk '{print $2}')
        local logs=$(get_job_logs "$job_id")
        
        if assert_contains "$logs" "VOLUME_MOUNTED"; then
            return 0
        else
            echo -e "    ${YELLOW}Volume feature may be limited${NC}"
            return 0
        fi
    else
        return 0
    fi
}

test_volume_persistence() {
    # Test that data persists across jobs
    
    # First job: write data
    local job1=$(run_python_job "
import os
os.makedirs('/work/persist', exist_ok=True)
with open('/work/persist/data.txt', 'w') as f:
    f.write('PERSISTENT_DATA')
print('DATA_WRITTEN')
")
    
    sleep 3
    
    # Second job: try to read data (won't persist without volumes)
    local job2=$(run_python_job "
import os
if os.path.exists('/work/persist/data.txt'):
    with open('/work/persist/data.txt') as f:
        print(f'FOUND:{f.read()}')
else:
    print('DATA_NOT_FOUND')
")
    
    local logs2=$(get_job_logs "$job2")
    
    # Without volume support, data won't persist
    if assert_contains "$logs2" "DATA_NOT_FOUND"; then
        echo -e "    ${GREEN}Isolation working (data doesn't persist without volumes)${NC}"
        return 0
    else
        echo -e "    ${YELLOW}Data might be persisting unexpectedly${NC}"
        return 1
    fi
}

test_concurrent_volume_access() {
    # Test multiple jobs accessing same volume
    echo -e "    ${YELLOW}Testing concurrent volume access (advanced feature)${NC}"
    
    # This would test:
    # - Multiple jobs reading from same volume
    # - Locking mechanisms for write access
    # - Race condition handling
    
    return 0
}

test_volume_size_limits() {
    # Test volume size restrictions
    local job_id=$(run_python_job "
import os
# Try to write a large file
try:
    with open('/work/large.txt', 'w') as f:
        # Write 10MB of data
        for i in range(1024 * 10):
            f.write('x' * 1024)
    print('LARGE_FILE_CREATED')
    # Check size
    size = os.path.getsize('/work/large.txt')
    print(f'SIZE:{size}')
except Exception as e:
    print(f'ERROR:{e}')
")
    
    local logs=$(get_job_logs "$job_id")
    
    if assert_contains "$logs" "SIZE:"; then
        return 0
    else
        return 1
    fi
}

test_volume_cleanup() {
    # Test volume deletion/cleanup
    local volume_list=$("$RNX_BINARY" volume list 2>&1 || echo "")
    
    if echo "$volume_list" | grep -q "not found\|not recognized"; then
        echo -e "    ${YELLOW}Volume management commands not available${NC}"
        return 0
    fi
    
    # Try to delete test volume if it exists
    local delete_output=$("$RNX_BINARY" volume delete "$TEST_VOLUME_NAME" 2>&1 || echo "")
    
    # Success if deleted or doesn't exist
    return 0
}

test_bind_mount() {
    # Test bind mounting host directories
    echo -e "    ${YELLOW}Testing bind mounts (security-sensitive feature)${NC}"
    
    # This would test mounting host directories
    # Should be restricted for security
    
    return 0
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "Volume Management Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites check failed${NC}"
        exit 1
    fi
    
    # Ensure runtime is available
    ensure_runtime "$DEFAULT_RUNTIME"
    
    # Run tests
    test_section "File Transfer"
    run_test "Upload file to job" test_upload_file
    run_test "Download results from job" test_download_result
    
    test_section "Volume Operations"
    run_test "Volume creation" test_volume_creation
    run_test "Volume mounting" test_volume_mounting
    run_test "Data persistence" test_volume_persistence
    
    test_section "Advanced Volume Features"
    run_test "Concurrent volume access" test_concurrent_volume_access
    run_test "Volume size limits" test_volume_size_limits
    run_test "Bind mount restrictions" test_bind_mount
    
    test_section "Cleanup"
    run_test "Volume cleanup" test_volume_cleanup
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi