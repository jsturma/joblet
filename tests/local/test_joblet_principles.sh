#!/bin/bash

# Joblet Core Principles Validation Script
# Tests: PID namespaces, filesystem isolation, cgroups, network isolation, security boundaries

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

print_test() {
    echo -e "\n${BLUE}🧪 $1${NC}"
}

test_result() {
    if [[ $1 -eq 0 ]]; then
        echo -e "${GREEN}✅ PASS: $2${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL: $2${NC}"
        ((FAILED++))
    fi
}

run_test_job() {
    local cmd="$1"
    local job_id=$(../../bin/rnx run "$cmd" | grep "ID:" | awk '{print $2}')
    echo "$job_id"
}

wait_and_get_logs() {
    local job_id="$1"
    sleep 2  # Give job time to complete
    ../../bin/rnx log "$job_id" 2>/dev/null || echo "LOG_ERROR"
}

echo -e "${YELLOW}🚀 Joblet Core Principles Test Suite${NC}"
echo -e "Testing isolation, security, and resource management...\n"

# Test 1: PID Namespace Isolation
print_test "PID Namespace Isolation"
job_id=$(run_test_job "ps -aux")
output=$(wait_and_get_logs "$job_id")
# Look for ps process in output
ps_line=$(echo "$output" | grep "ps -aux" | head -1)
if [[ -n "$ps_line" ]]; then
    # Count total processes shown by ps command
    process_count=$(echo "$output" | grep -E "^\s*[A-Za-z0-9]+\s+[0-9]+" | wc -l)
    test_result $((process_count <= 3 ? 0 : 1)) "Jobs see only their own processes ($process_count visible)"
else
    test_result 1 "Could not find ps command output"
fi

# Test 2: Process runs as PID 1
pid1_check=$(echo "$output" | grep -c "0.*1.*ps")
test_result $((pid1_check > 0 ? 0 : 1)) "Job process runs as PID 1 in namespace"

# Test 3: Filesystem Isolation (/proc)
print_test "Filesystem Isolation"
job_id=$(run_test_job "ls /proc/ | wc -l")
output=$(wait_and_get_logs "$job_id")
proc_entries=$(echo "$output" | tail -1 | grep -E "^[0-9]+$")
test_result $((proc_entries < 50 ? 0 : 1)) "/proc filesystem isolated ($proc_entries entries)"

# Test 4: Working directory isolation
job_id=$(run_test_job "pwd")
output=$(wait_and_get_logs "$job_id")
work_dir_check=$(echo "$output" | grep -c "/work")
test_result $((work_dir_check > 0 ? 0 : 1)) "Job runs in isolated work directory"

# Test 5: Network Namespace
print_test "Network Isolation"
job_id=$(run_test_job "ip addr show | grep -c 'inet '")
output=$(wait_and_get_logs "$job_id")
inet_count=$(echo "$output" | tail -1 | grep -E "^[0-9]+$")
test_result $((inet_count <= 3 ? 0 : 1)) "Network interfaces isolated ($inet_count interfaces)"

# Test 6: Security Boundaries - Cannot see host processes
print_test "Security Boundaries"
job_id=$(run_test_job "ps aux | grep -c systemd")
output=$(wait_and_get_logs "$job_id")
systemd_count=$(echo "$output" | tail -1 | grep -E "^[0-9]+$")
test_result $((systemd_count == 0 ? 0 : 1)) "Cannot see host processes (systemd: $systemd_count)"

# Test 7: Cgroup Integration
print_test "Cgroup Resource Management"
job_id=$(run_test_job --memory 100M "cat /proc/self/cgroup | grep -c memory")
output=$(wait_and_get_logs "$job_id")
cgroup_check=$(echo "$output" | tail -1 | grep -E "^[0-9]+$")
test_result $((cgroup_check > 0 ? 0 : 1)) "Process assigned to cgroups with limits"

# Test 8: Upload/Volume Isolation
print_test "Upload & Volume Isolation"
echo "test_content_$(date)" > /tmp/test_file.txt
job_id=$(run_test_job --upload /tmp/test_file.txt "ls /work/test_file.txt && cat /work/test_file.txt")
output=$(wait_and_get_logs "$job_id")
upload_check=$(echo "$output" | grep -c "test_content_")
test_result $((upload_check > 0 ? 0 : 1)) "File uploads work in isolated workspace"
rm -f /tmp/test_file.txt

# Test 9: Job Lifecycle Management
print_test "Job Lifecycle & Resource Cleanup"
job_id=$(run_test_job "echo 'lifecycle_test' && sleep 1")
sleep 3
status=$(../../bin/rnx status "$job_id" | grep "Status:" | sed 's/.*Status: \([^[]*\).*/\1/' | tr -d ' ')
test_result $([[ "$status" == "COMPLETED" ]] && echo 0 || echo 1) "Job completes successfully"

# Cleanup test
../../bin/rnx delete "$job_id" >/dev/null 2>&1
deletion_check=$?
test_result $deletion_check "Job resources cleaned up after completion"

# Test 10: Concurrent Job Isolation
print_test "Concurrent Job Isolation"
job1_id=$(run_test_job "echo 'job1' && ps aux | wc -l")
job2_id=$(run_test_job "echo 'job2' && ps aux | wc -l")
sleep 3

output1=$(wait_and_get_logs "$job1_id")
output2=$(wait_and_get_logs "$job2_id")
count1=$(echo "$output1" | tail -1 | grep -E "^[0-9]+$")
count2=$(echo "$output2" | tail -1 | grep -E "^[0-9]+$")

concurrent_isolated=$((count1 <= 5 && count2 <= 5 ? 0 : 1))
test_result $concurrent_isolated "Concurrent jobs isolated from each other (job1: $count1, job2: $count2 processes)"

# Cleanup
../../bin/rnx delete "$job1_id" >/dev/null 2>&1 || true
../../bin/rnx delete "$job2_id" >/dev/null 2>&1 || true

# Summary
echo -e "\n${BLUE}================================================${NC}"
echo -e "${BLUE}Test Results Summary${NC}"  
echo -e "${BLUE}================================================${NC}"
echo -e "Total Tests: $((PASSED + FAILED))"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"

if [[ $FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}Joblet isolation and security principles are working correctly.${NC}"
    echo -e "\n${YELLOW}✅ Verified Principles:${NC}"
    echo -e "  • PID namespace isolation (jobs see only own processes)"
    echo -e "  • Filesystem isolation (chroot + /proc remount)"  
    echo -e "  • Network namespace isolation"
    echo -e "  • Cgroup resource management"
    echo -e "  • Security boundaries (no host process visibility)"
    echo -e "  • Upload/volume workspace isolation"
    echo -e "  • Job lifecycle management"
    echo -e "  • Concurrent job isolation"
    exit 0
else
    echo -e "\n${RED}⚠️  $FAILED test(s) failed. Please review isolation implementation.${NC}"
    exit 1
fi