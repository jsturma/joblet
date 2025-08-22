#!/bin/bash

# Quick Joblet Isolation Test
# Fast validation of core isolation principles

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}🚀 Quick Joblet Isolation Test${NC}\n"

# Test 1: PID Namespace Isolation
echo -e "${YELLOW}Testing PID namespace isolation...${NC}"
JOB_OUTPUT=$(timeout 10s ../../bin/rnx run ps -aux 2>/dev/null | grep "ID:" | awk '{print $2}')
if [[ -n "$JOB_OUTPUT" ]]; then
    sleep 2
    PROCESSES=$(../../bin/rnx log "$JOB_OUTPUT" 2>/dev/null | grep -E "^\s*[0-9]+.*ps" | wc -l)
    if [[ $PROCESSES -eq 1 ]]; then
        echo -e "${GREEN}✅ PID isolation working - only 1 process visible${NC}"
    else
        echo -e "${RED}❌ PID isolation failed - $PROCESSES processes visible${NC}"
    fi
fi

# Test 2: Basic filesystem isolation
echo -e "${YELLOW}Testing filesystem isolation...${NC}"
JOB_OUTPUT=$(timeout 10s ../../bin/rnx run "ls /proc/ | wc -l" 2>/dev/null | grep "ID:" | awk '{print $2}')
if [[ -n "$JOB_OUTPUT" ]]; then
    sleep 2
    PROC_COUNT=$(../../bin/rnx log "$JOB_OUTPUT" 2>/dev/null | tail -1 | grep -E "^[0-9]+$")
    if [[ $PROC_COUNT -lt 50 ]]; then
        echo -e "${GREEN}✅ Filesystem isolation working - /proc has $PROC_COUNT entries${NC}"
    else
        echo -e "${RED}❌ Filesystem isolation failed - /proc has $PROC_COUNT entries${NC}"
    fi
fi

# Test 3: Network namespace
echo -e "${YELLOW}Testing network isolation...${NC}"
JOB_OUTPUT=$(timeout 10s ../../bin/rnx run "ip addr | grep -c 'inet '" 2>/dev/null | grep "ID:" | awk '{print $2}')
if [[ -n "$JOB_OUTPUT" ]]; then
    sleep 2
    INET_COUNT=$(../../bin/rnx log "$JOB_OUTPUT" 2>/dev/null | tail -1 | grep -E "^[0-9]+$")
    if [[ $INET_COUNT -le 3 ]]; then
        echo -e "${GREEN}✅ Network isolation working - $INET_COUNT network interfaces${NC}"
    else
        echo -e "${RED}❌ Network isolation failed - $INET_COUNT interfaces${NC}"
    fi
fi

# Test 4: Security boundary
echo -e "${YELLOW}Testing security boundaries...${NC}"
JOB_OUTPUT=$(timeout 10s ../../bin/rnx run "ps aux | grep systemd | wc -l" 2>/dev/null | grep "ID:" | awk '{print $2}')
if [[ -n "$JOB_OUTPUT" ]]; then
    sleep 2
    SYSTEMD_COUNT=$(../../bin/rnx log "$JOB_OUTPUT" 2>/dev/null | tail -1 | grep -E "^[0-9]+$")
    if [[ $SYSTEMD_COUNT -eq 0 ]]; then
        echo -e "${GREEN}✅ Security boundaries working - no host processes visible${NC}"
    else
        echo -e "${RED}❌ Security boundary breach - $SYSTEMD_COUNT host processes visible${NC}"
    fi
fi

echo -e "\n${GREEN}🎉 Quick test completed!${NC}"