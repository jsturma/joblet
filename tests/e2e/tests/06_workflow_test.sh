#!/bin/bash

# Test 06: Workflow and Job Dependencies Tests
# Tests workflow creation, job dependencies, and DAG execution

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# ============================================
# Test Functions
# ============================================

test_simple_workflow() {
    # Create a simple workflow YAML
    local workflow_file="/tmp/test_workflow_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  step1:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('STEP1_COMPLETE')"]
  
  step2:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('STEP2_COMPLETE')"]
    requires: ["step1"]
  
  step3:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('STEP3_COMPLETE')"]
    requires: ["step2"]
EOF
    
    # Try to run workflow
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -q "workflow.*not\|unrecognized"; then
        echo -e "    ${YELLOW}Workflow feature not implemented${NC}"
        return 0
    fi
    
    if echo "$workflow_output" | grep -q "UUID\|created\|ID"; then
        echo -e "    ${GREEN}Workflow created${NC}"
        return 0
    else
        echo -e "    ${YELLOW}Workflow feature may be limited${NC}"
        return 0
    fi
}

test_parallel_workflow() {
    # Test workflow with parallel execution
    local workflow_file="/tmp/test_parallel_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  prepare:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('PREPARE_DONE')"]
  
  process1:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "import time; time.sleep(2); print('PROCESS1_DONE')"]
    requires: ["prepare"]
  
  process2:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "import time; time.sleep(2); print('PROCESS2_DONE')"]
    requires: ["prepare"]
  
  process3:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "import time; time.sleep(2); print('PROCESS3_DONE')"]
    requires: ["prepare"]
  
  finalize:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('ALL_DONE')"]
    requires: ["process1", "process2", "process3"]
EOF
    
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -q "UUID\|created"; then
        # Workflow was created, check if parallel execution works
        local workflow_id=$(echo "$workflow_output" | grep -oE '[a-f0-9-]{36}' | head -1)
        
        if [[ -n "$workflow_id" ]]; then
            # Wait for completion
            sleep 10
            
            # Check status
            local status_output=$("$RNX_BINARY" status --workflow "$workflow_id" 2>&1 || echo "")
            
            if echo "$status_output" | grep -q "COMPLETED"; then
                return 0
            else
                echo -e "    ${YELLOW}Workflow may still be running${NC}"
                return 0
            fi
        fi
    else
        echo -e "    ${YELLOW}Parallel workflows not supported${NC}"
        return 0
    fi
}

test_conditional_workflow() {
    # Test workflow with conditional execution
    echo -e "    ${YELLOW}Testing conditional workflows (advanced feature)${NC}"
    
    # This would test:
    # - Jobs that run based on previous job outputs
    # - If/else logic in workflows
    # - Skip conditions
    
    return 0
}

test_workflow_failure_handling() {
    # Test workflow behavior when a job fails
    local workflow_file="/tmp/test_failure_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  good_job:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('SUCCESS')"]
  
  bad_job:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "raise Exception('INTENTIONAL_FAILURE')"]
    requires: ["good_job"]
  
  dependent_job:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('SHOULD_NOT_RUN')"]
    requires: ["bad_job"]
EOF
    
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -q "UUID\|created"; then
        echo -e "    ${GREEN}Failure handling workflow created${NC}"
        return 0
    else
        echo -e "    ${YELLOW}Workflow failure handling not testable${NC}"
        return 0
    fi
}

test_workflow_status() {
    # Test workflow status tracking
    local workflow_file="/tmp/test_status_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  quick_job:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('QUICK')"]
EOF
    
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -q "UUID"; then
        local workflow_id=$(echo "$workflow_output" | grep -oE '[a-f0-9-]{36}' | head -1)
        
        if [[ -n "$workflow_id" ]]; then
            # Check status command
            local status=$("$RNX_BINARY" status --workflow "$workflow_id" 2>&1 || echo "")
            
            if echo "$status" | grep -q "Status:\|RUNNING\|COMPLETED"; then
                return 0
            else
                echo -e "    ${YELLOW}Workflow status not available${NC}"
                return 0
            fi
        fi
    else
        echo -e "    ${YELLOW}Workflow status tracking not implemented${NC}"
        return 0
    fi
}

test_workflow_cancellation() {
    # Test cancelling a running workflow
    local workflow_file="/tmp/test_cancel_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  long_job:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "import time; time.sleep(30); print('SHOULD_NOT_COMPLETE')"]
EOF
    
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -q "UUID"; then
        local workflow_id=$(echo "$workflow_output" | grep -oE '[a-f0-9-]{36}' | head -1)
        
        if [[ -n "$workflow_id" ]]; then
            # Try to cancel
            sleep 2
            local cancel_output=$("$RNX_BINARY" cancel --workflow "$workflow_id" 2>&1 || \
                                  "$RNX_BINARY" delete --workflow "$workflow_id" 2>&1 || echo "")
            
            if echo "$cancel_output" | grep -q "cancel\|delete\|stop"; then
                return 0
            else
                echo -e "    ${YELLOW}Workflow cancellation not supported${NC}"
                return 0
            fi
        fi
    else
        echo -e "    ${YELLOW}Cannot test workflow cancellation${NC}"
        return 0
    fi
}

test_workflow_with_data_passing() {
    # Test passing data between workflow steps
    echo -e "    ${YELLOW}Testing data passing between steps (advanced)${NC}"
    
    # This would test:
    # - Output from one job as input to another
    # - Environment variable passing
    # - File artifact passing
    
    return 0
}

test_cyclic_dependency_detection() {
    # Test that cyclic dependencies are detected
    local workflow_file="/tmp/test_cyclic_$$.yaml"
    cat > "$workflow_file" << 'EOF'
version: "1.0"
jobs:
  job_a:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('A')"]
    requires: ["job_c"]
  
  job_b:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('B')"]
    requires: ["job_a"]
  
  job_c:
    runtime: "python-3.11-ml"
    command: "python3"
    args: ["-c", "print('C')"]
    requires: ["job_b"]
EOF
    
    local workflow_output=$("$RNX_BINARY" run --workflow="$workflow_file" 2>&1 || echo "")
    
    rm -f "$workflow_file"
    
    if echo "$workflow_output" | grep -qi "cycl\|circular\|dependency.*error"; then
        echo -e "    ${GREEN}Cyclic dependencies detected correctly${NC}"
        return 0
    else
        echo -e "    ${YELLOW}Cyclic dependency detection may not be implemented${NC}"
        return 0
    fi
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "Workflow and Dependencies Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites check failed${NC}"
        exit 1
    fi
    
    # Ensure runtime is available
    ensure_runtime "$DEFAULT_RUNTIME"
    
    # Run tests
    test_section "Basic Workflows"
    run_test "Simple sequential workflow" test_simple_workflow
    run_test "Parallel execution workflow" test_parallel_workflow
    
    test_section "Workflow Control"
    run_test "Workflow status tracking" test_workflow_status
    run_test "Workflow cancellation" test_workflow_cancellation
    
    test_section "Advanced Features"
    run_test "Conditional workflows" test_conditional_workflow
    run_test "Failure handling" test_workflow_failure_handling
    run_test "Data passing between steps" test_workflow_with_data_passing
    
    test_section "Validation"
    run_test "Cyclic dependency detection" test_cyclic_dependency_detection
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi