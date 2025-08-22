#!/bin/bash

# Final Joblet Isolation Test - Simple & Clear
# Tests core isolation principles with clear output

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🛡️  Joblet Isolation Test Results${NC}\n"

echo -e "${YELLOW}1. PID Namespace Isolation${NC}"
echo "========================================="
../../bin/rnx run ps -aux
sleep 2
echo -e "\n${GREEN}✓ PASS if only 1-2 processes visible (ps command itself)${NC}\n"

echo -e "${YELLOW}2. Filesystem Isolation (/proc and working directory)${NC}"
echo "========================================="
../../bin/rnx run "ls /proc/ | wc -l && echo '---' && pwd"
sleep 2
echo -e "\n${GREEN}✓ PASS if /proc has <50 entries and pwd shows /work${NC}\n"

echo -e "${YELLOW}3. Network Isolation${NC}"
echo "========================================="
../../bin/rnx run "ip link show | grep -c '^[0-9]'"
sleep 2
echo -e "\n${GREEN}✓ PASS if only 1-2 network interfaces visible${NC}\n"

echo -e "${YELLOW}4. Security Boundaries - Host Process Visibility${NC}"
echo "========================================="
../../bin/rnx run sh -c "ps aux | grep systemd | wc -l"
sleep 2  
echo -e "\n${GREEN}✓ PASS if result is 0 (no systemd processes visible)${NC}\n"

echo -e "${YELLOW}5. Resource Management${NC}"
echo "========================================="
../../bin/rnx run --memory 100M echo "Memory limit test completed"
sleep 2
echo -e "\n${GREEN}✓ PASS if job completes successfully with resource limits${NC}\n"

echo -e "${YELLOW}6. Cgroup Integration${NC}"  
echo "========================================="
../../bin/rnx run "cat /proc/self/cgroup | head -3"
sleep 2
echo -e "\n${GREEN}✓ PASS if cgroup paths are shown (process is in cgroups)${NC}\n"

echo -e "${BLUE}🎯 Summary${NC}"
echo "================================"
echo -e "${YELLOW}All tests completed!${NC}"
echo ""
echo -e "${GREEN}✅ PID Isolation:${NC} Jobs should see only their own processes" 
echo -e "${GREEN}✅ Filesystem:${NC} /proc isolated, running in /work directory"
echo -e "${GREEN}✅ Network:${NC} Limited network interfaces visible"
echo -e "${GREEN}✅ Security:${NC} No host processes (systemd) visible"  
echo -e "${GREEN}✅ Resources:${NC} Resource limits working"
echo -e "${GREEN}✅ Cgroups:${NC} Process assigned to cgroups"
echo ""
echo -e "${BLUE}If all outputs match the expected results above, isolation is working correctly! 🎉${NC}"