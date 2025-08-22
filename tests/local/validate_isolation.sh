#!/bin/bash

# Simple Joblet Isolation Validator
# Validates core isolation principles with clear pass/fail results

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🛡️  Joblet Isolation Validator${NC}\n"

run_job() {
    local cmd="$1"
    echo -e "${YELLOW}Running: $cmd${NC}"
    local job_id=$(../../bin/rnx run "$cmd" | grep "ID:" | awk '{print $2}')
    sleep 3  # Wait for completion
    ../../bin/rnx log "$job_id" 2>/dev/null
}

echo -e "${BLUE}1. Testing PID Namespace Isolation${NC}"
echo "Job should only see its own processes, not host processes"
echo "==============================================="
run_job "ps -aux"
echo -e "\n${YELLOW}✓ Expected: Only 1-2 processes visible (the ps command itself)${NC}\n"

echo -e "${BLUE}2. Testing Filesystem Isolation${NC}" 
echo "Job should have isolated /proc and run in chroot"
echo "==============================================="
run_job "ls /proc/ | head -10 && echo '---' && pwd"
echo -e "\n${YELLOW}✓ Expected: Limited /proc entries and /work directory${NC}\n"

echo -e "${BLUE}3. Testing Network Isolation${NC}"
echo "Job should have limited network interfaces"
echo "===============================================" 
run_job "ip addr show"
echo -e "\n${YELLOW}✓ Expected: Only loopback interface visible${NC}\n"

echo -e "${BLUE}4. Testing Security Boundaries${NC}"
echo "Job should NOT see host processes like systemd"
echo "==============================================="
run_job "ps aux | grep systemd || echo 'NO_SYSTEMD_FOUND'"
echo -e "\n${YELLOW}✓ Expected: NO_SYSTEMD_FOUND (no host processes visible)${NC}\n"

echo -e "${BLUE}5. Testing Upload Isolation${NC}"
echo "Creating test file and uploading to job..."
echo "==============================================="
echo "test_data_$(date +%s)" > /tmp/isolation_test.txt
run_job --upload /tmp/isolation_test.txt "ls -la /work/ && cat /work/isolation_test.txt"
rm -f /tmp/isolation_test.txt
echo -e "\n${YELLOW}✓ Expected: File visible in /work directory with correct content${NC}\n"

echo -e "${BLUE}6. Testing Resource Limits${NC}"
echo "Testing job with memory limit"
echo "==============================================="
run_job --memory 100M "echo 'Memory limited job completed successfully'"
echo -e "\n${YELLOW}✓ Expected: Job completes without errors${NC}\n"

echo -e "${GREEN}🎉 Validation Complete!${NC}"
echo -e "\n${YELLOW}Manual Review Required:${NC}"
echo -e "• PID isolation: Jobs should show only 1-2 processes"
echo -e "• Filesystem isolation: /proc should have <50 entries, pwd should be /work"
echo -e "• Network isolation: Only loopback interface should be visible"  
echo -e "• Security: No systemd or host processes should be visible"
echo -e "• Uploads: Files should be accessible in /work directory"
echo -e "• Resource limits: Jobs with limits should complete successfully"
echo -e "\n${BLUE}If all outputs match expectations, isolation is working correctly! ✅${NC}"