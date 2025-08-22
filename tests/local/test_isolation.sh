#!/bin/bash

# Joblet Isolation & Security Test Script
# Tests all core isolation principles: PID namespaces, filesystem isolation,
# cgroups resource limits, network isolation, and security boundaries.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# RNX binary path
RNX_BIN="../../bin/rnx"

# Check if rnx binary exists
if [[ ! -f "$RNX_BIN" ]]; then
    echo -e "${RED}Error: rnx binary not found at $RNX_BIN${NC}"
    echo "Please run 'make all' first"
    exit 1
fi

print_header() {
    echo -e "\n${BLUE}================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================${NC}"
}

print_test() {
    echo -e "\n${YELLOW}🧪 Test: $1${NC}"
    ((TOTAL_TESTS++))
}

test_pass() {
    echo -e "${GREEN}✅ PASS: $1${NC}"
    ((PASSED_TESTS++))
}

test_fail() {
    echo -e "${RED}❌ FAIL: $1${NC}"
    ((FAILED_TESTS++))
}

run_job_and_get_output() {
    local cmd="$1"
    local job_id=$(timeout 30s $RNX_BIN run $cmd | grep "ID:" | awk '{print $2}')
    
    if [[ -z "$job_id" ]]; then
        echo "ERROR: Failed to get job ID"
        return 1
    fi
    
    # Wait for job completion
    local timeout=30
    local elapsed=0
    while [[ $elapsed -lt $timeout ]]; do
        local status=$($RNX_BIN status "$job_id" | grep "Status:" | awk '{print $2}' | sed 's/\x1b\[[0-9;]*m//g')
        if [[ "$status" == "COMPLETED" || "$status" == "FAILED" ]]; then
            break
        fi
        sleep 1
        ((elapsed++))
    done
    
    # Get job output
    $RNX_BIN log "$job_id" 2>/dev/null | tail -50
}

wait_for_job_completion() {
    local job_id="$1"
    local timeout=30
    local elapsed=0
    
    while [[ $elapsed -lt $timeout ]]; do
        local status=$($RNX_BIN status "$job_id" | grep "Status:" | awk '{print $2}' | sed 's/\x1b\[[0-9;]*m//g')
        if [[ "$status" == "COMPLETED" || "$status" == "FAILED" ]]; then
            echo "$status"
            return 0
        fi
        sleep 1
        ((elapsed++))
    done
    
    echo "TIMEOUT"
    return 1
}

test_pid_namespace_isolation() {
    print_test "PID Namespace Isolation"
    
    # Test 1: Job should only see its own processes
    local output=$(run_job_and_get_output "ps -aux")
    local process_count=$(echo "$output" | grep -E "^\s*[0-9]+" | wc -l)
    
    if [[ $process_count -le 3 ]]; then
        test_pass "Job sees only its own processes ($process_count processes visible)"
    else
        test_fail "Job can see too many processes ($process_count processes visible)"
        echo "$output" | head -10
    fi
    
    # Test 2: Job should run as PID 1 in namespace
    local pid1_process=$(echo "$output" | grep "^\s*0\s*1\s" | head -1)
    if [[ -n "$pid1_process" ]]; then
        test_pass "Job process runs as PID 1 in namespace"
    else
        test_fail "Job process does not run as PID 1"
    fi
}

test_filesystem_isolation() {
    print_test "Filesystem Isolation"
    
    # Test 1: Job should have isolated /proc
    local output=$(run_job_and_get_output "ls /proc/ | head -5")
    local proc_entries=$(echo "$output" | grep -E "^[0-9]+$" | wc -l)
    
    if [[ $proc_entries -gt 0 && $proc_entries -le 5 ]]; then
        test_pass "Filesystem has isolated /proc with $proc_entries process entries"
    else
        test_fail "Filesystem isolation failed - /proc shows $proc_entries entries"
    fi
    
    # Test 2: Job should be in chroot environment
    local output=$(run_job_and_get_output "pwd && ls -la /")
    if echo "$output" | grep -q "/work"; then
        test_pass "Job runs in chroot environment with work directory"
    else
        test_fail "Job not in proper chroot environment"
    fi
}

test_cgroup_limits() {
    print_test "Cgroup Resource Limits"
    
    # Test with memory limit
    local job_id=$(timeout 30s $RNX_BIN run --memory 100M "cat /proc/self/cgroup" | grep "ID:" | awk '{print $2}')
    
    if [[ -n "$job_id" ]]; then
        local status=$(wait_for_job_completion "$job_id")
        if [[ "$status" == "COMPLETED" ]]; then
            local output=$($RNX_BIN log "$job_id" 2>/dev/null)
            if echo "$output" | grep -q "memory"; then
                test_pass "Job has cgroup memory limits configured"
            else
                test_fail "Cgroup limits not properly configured"
            fi
        else
            test_fail "Job with memory limit failed to complete"
        fi
    else
        test_fail "Failed to start job with memory limit"
    fi
    
    # Test CPU limit
    local job_id=$(timeout 30s $RNX_BIN run --cpu 50 "sleep 1 && echo 'CPU limit test completed'" | grep "ID:" | awk '{print $2}')
    
    if [[ -n "$job_id" ]]; then
        local status=$(wait_for_job_completion "$job_id")
        if [[ "$status" == "COMPLETED" ]]; then
            test_pass "Job with CPU limit completed successfully"
        else
            test_fail "Job with CPU limit failed"
        fi
    else
        test_fail "Failed to start job with CPU limit"
    fi
}

test_network_isolation() {
    print_test "Network Isolation"
    
    # Test 1: Default isolation
    local output=$(run_job_and_get_output "ip addr show | grep -E 'inet|link'")
    
    # Should have loopback but limited network interfaces
    if echo "$output" | grep -q "lo:"; then
        test_pass "Job has loopback interface in network namespace"
    else
        test_fail "Job missing loopback interface"
    fi
    
    # Test 2: Network connectivity test
    local output=$(run_job_and_get_output "ping -c 1 127.0.0.1 >/dev/null 2>&1 && echo 'LOOPBACK_OK' || echo 'LOOPBACK_FAIL'")
    if echo "$output" | grep -q "LOOPBACK_OK"; then
        test_pass "Loopback networking works in namespace"
    else
        test_fail "Loopback networking failed"
    fi
}

test_security_boundaries() {
    print_test "Security Boundaries"
    
    # Test 1: Job cannot access host filesystem
    local output=$(run_job_and_get_output "ls /host 2>&1 || echo 'HOST_ACCESS_DENIED'")
    if echo "$output" | grep -q "HOST_ACCESS_DENIED\|No such file"; then
        test_pass "Job cannot access host filesystem"
    else
        test_fail "Job can access host filesystem - security boundary breached"
    fi
    
    # Test 2: Job cannot see host processes
    local output=$(run_job_and_get_output "ps aux | grep systemd | wc -l")
    local systemd_count=$(echo "$output" | grep -E "^[0-9]+$" | head -1)
    if [[ "$systemd_count" -eq 0 ]]; then
        test_pass "Job cannot see host systemd processes"
    else
        test_fail "Job can see $systemd_count host systemd processes"
    fi
    
    # Test 3: User isolation
    local output=$(run_job_and_get_output "id && whoami")
    if echo "$output" | grep -q "uid=0"; then
        test_pass "Job runs as expected user (root in namespace)"
    else
        test_fail "Job user isolation failed"
    fi
}

test_resource_cleanup() {
    print_test "Resource Cleanup"
    
    # Start a job and let it complete
    local job_id=$(timeout 30s $RNX_BIN run "echo 'cleanup test' && sleep 2" | grep "ID:" | awk '{print $2}')
    
    if [[ -n "$job_id" ]]; then
        local status=$(wait_for_job_completion "$job_id")
        if [[ "$status" == "COMPLETED" ]]; then
            # Check if job can be deleted
            $RNX_BIN delete "$job_id" >/dev/null 2>&1
            if [[ $? -eq 0 ]]; then
                test_pass "Job resources cleaned up successfully"
            else
                test_fail "Job resource cleanup failed"
            fi
        else
            test_fail "Cleanup test job failed to complete"
        fi
    else
        test_fail "Failed to start cleanup test job"
    fi
}

test_volume_isolation() {
    print_test "Volume and Upload Isolation"
    
    # Create a test file to upload
    echo "test content for isolation" > /tmp/test_upload.txt
    
    # Test job with upload
    local job_id=$(timeout 30s $RNX_BIN run --upload /tmp/test_upload.txt "ls -la /work && cat /work/test_upload.txt" | grep "ID:" | awk '{print $2}')
    
    if [[ -n "$job_id" ]]; then
        local status=$(wait_for_job_completion "$job_id")
        if [[ "$status" == "COMPLETED" ]]; then
            local output=$($RNX_BIN log "$job_id" 2>/dev/null)
            if echo "$output" | grep -q "test content for isolation"; then
                test_pass "Upload isolation working - file accessible in job workspace"
            else
                test_fail "Upload isolation failed - file not found in workspace"
            fi
        else
            test_fail "Upload test job failed to complete"
        fi
    else
        test_fail "Failed to start upload test job"
    fi
    
    # Cleanup
    rm -f /tmp/test_upload.txt
}

test_concurrent_isolation() {
    print_test "Concurrent Job Isolation"
    
    # Start multiple jobs concurrently
    local job1_id=$(timeout 30s $RNX_BIN run "echo 'job1' && sleep 3 && ps aux | wc -l" | grep "ID:" | awk '{print $2}')
    local job2_id=$(timeout 30s $RNX_BIN run "echo 'job2' && sleep 3 && ps aux | wc -l" | grep "ID:" | awk '{print $2}')
    
    if [[ -n "$job1_id" && -n "$job2_id" ]]; then
        local status1=$(wait_for_job_completion "$job1_id")
        local status2=$(wait_for_job_completion "$job2_id")
        
        if [[ "$status1" == "COMPLETED" && "$status2" == "COMPLETED" ]]; then
            local output1=$($RNX_BIN log "$job1_id" 2>/dev/null | tail -1)
            local output2=$($RNX_BIN log "$job2_id" 2>/dev/null | tail -1)
            
            # Each job should see only its own processes (low count)
            local count1=$(echo "$output1" | grep -E "^[0-9]+$" | head -1)
            local count2=$(echo "$output2" | grep -E "^[0-9]+$" | head -1)
            
            if [[ "$count1" -le 5 && "$count2" -le 5 ]]; then
                test_pass "Concurrent jobs properly isolated from each other"
            else
                test_fail "Concurrent jobs not properly isolated (job1: $count1, job2: $count2 processes)"
            fi
        else
            test_fail "Concurrent jobs failed to complete"
        fi
        
        # Cleanup
        $RNX_BIN delete "$job1_id" >/dev/null 2>&1 || true
        $RNX_BIN delete "$job2_id" >/dev/null 2>&1 || true
    else
        test_fail "Failed to start concurrent jobs"
    fi
}

test_error_handling() {
    print_test "Error Handling and Limits"
    
    # Test invalid command
    local job_id=$(timeout 30s $RNX_BIN run "nonexistent_command_12345" | grep "ID:" | awk '{print $2}' || true)
    
    if [[ -n "$job_id" ]]; then
        local status=$(wait_for_job_completion "$job_id")
        if [[ "$status" == "FAILED" ]]; then
            test_pass "Invalid command properly handled with FAILED status"
        else
            test_fail "Invalid command not properly handled (status: $status)"
        fi
        $RNX_BIN delete "$job_id" >/dev/null 2>&1 || true
    else
        test_fail "Failed to start invalid command test"
    fi
}

# Main test execution
main() {
    print_header "Joblet Isolation & Security Test Suite"
    echo -e "Testing all core isolation principles..."
    echo -e "RNX Binary: $RNX_BIN"
    
    # Run all tests
    test_pid_namespace_isolation
    test_filesystem_isolation
    test_cgroup_limits
    test_network_isolation
    test_security_boundaries
    test_resource_cleanup
    test_volume_isolation
    test_concurrent_isolation
    test_error_handling
    
    # Print summary
    print_header "Test Results Summary"
    echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "\n${GREEN}🎉 ALL TESTS PASSED! Joblet isolation is working correctly.${NC}"
        exit 0
    else
        echo -e "\n${RED}⚠️  Some tests failed. Please check the isolation implementation.${NC}"
        exit 1
    fi
}

# Handle script interruption
trap 'echo -e "\n${YELLOW}Test interrupted. Cleaning up...${NC}"; exit 130' INT

# Run main function
main "$@"