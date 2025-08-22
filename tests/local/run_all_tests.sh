#!/bin/bash

# Master Test Runner for Joblet Isolation Tests
# Runs all available test scripts and provides a summary

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Joblet Isolation Test Suite Runner${NC}\n"

# Check if we're in the right directory
if [[ ! -f "../../bin/rnx" ]]; then
    echo -e "${RED}Error: Please run this script from the tests/local directory${NC}"
    echo "Usage: cd tests/local && ./run_all_tests.sh"
    exit 1
fi

# Check if rnx binary exists
if [[ ! -x "../../bin/rnx" ]]; then
    echo -e "${RED}Error: rnx binary not found or not executable${NC}"
    echo "Please run 'make all' from the project root first"
    exit 1
fi

echo -e "${YELLOW}Available test scripts:${NC}"
echo "1. final_isolation_test.sh - Quick visual verification (recommended)"
echo "2. test_joblet_principles.sh - Comprehensive automated testing"
echo "3. validate_isolation.sh - Detailed validation with explanations"
echo ""

# Default to running the recommended test
TEST_CHOICE=${1:-1}

case $TEST_CHOICE in
    1)
        echo -e "${BLUE}Running Quick Visual Test...${NC}\n"
        ./final_isolation_test.sh
        ;;
    2) 
        echo -e "${BLUE}Running Comprehensive Automated Test...${NC}\n"
        ./test_joblet_principles.sh
        ;;
    3)
        echo -e "${BLUE}Running Detailed Validation...${NC}\n" 
        ./validate_isolation.sh
        ;;
    "all")
        echo -e "${BLUE}Running ALL test scripts...${NC}\n"
        
        echo -e "${YELLOW}=== 1. Quick Visual Test ===${NC}"
        ./final_isolation_test.sh
        echo -e "\n${YELLOW}Press Enter to continue to comprehensive test...${NC}"
        read -r
        
        echo -e "${YELLOW}=== 2. Comprehensive Automated Test ===${NC}"
        ./test_joblet_principles.sh
        echo -e "\n${YELLOW}Press Enter to continue to detailed validation...${NC}"
        read -r
        
        echo -e "${YELLOW}=== 3. Detailed Validation ===${NC}"
        ./validate_isolation.sh
        ;;
    *)
        echo -e "${RED}Invalid choice. Use: 1, 2, 3, or 'all'${NC}"
        echo "Usage: $0 [1|2|3|all]"
        echo "  1 - Quick visual test (default)"
        echo "  2 - Comprehensive automated test"
        echo "  3 - Detailed validation"
        echo "  all - Run all tests sequentially"
        exit 1
        ;;
esac

echo -e "\n${GREEN}🎉 Test execution completed!${NC}"
echo -e "${BLUE}For more information, see README.md in this directory${NC}"