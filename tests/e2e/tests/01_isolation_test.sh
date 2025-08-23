#!/bin/bash

# Test 01: Core Isolation Tests
# Tests PID, filesystem, network, and security isolation

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# ============================================
# Test Functions
# ============================================

test_pid_isolation() {
    local job_id=$(run_python_job "import os; print(f'PID:{os.getpid()}:PPID:{os.getppid()}')")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if echo "$clean_output" | grep -q "PID:.*:PPID:"; then
        local pid=$(echo "$clean_output" | grep -o 'PID:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
        local ppid=$(echo "$clean_output" | grep -o 'PPID:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
        
        # Job should run as PID 1 (true isolation)
        assert_equals "$pid" "1" "Process should be PID 1 in isolated namespace"
        assert_equals "$ppid" "0" "Parent process should be PID 0 (kernel)"
    else
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        return 1
    fi
}

test_process_count() {
    local job_id=$(run_python_job "import os; pids=[p for p in os.listdir('/proc') if p.isdigit()]; print(f'PROC_COUNT:{len(pids)}:PIDS:{sorted([int(p) for p in pids])}'[:100])")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if echo "$clean_output" | grep -q "PROC_COUNT:"; then
        local proc_count=$(echo "$clean_output" | grep -o 'PROC_COUNT:[0-9]*' | cut -d: -f2)
        
        # True isolation should show only 1 process (the job itself)
        assert_equals "$proc_count" "1" "Should see exactly 1 process in isolated namespace"
    else
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        return 1
    fi
}

test_working_directory() {
    local job_id=$(run_python_job "import os; print(os.getcwd())")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if [[ -z "$clean_output" ]]; then
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        echo "    ${RED}Raw logs: $(echo "$logs" | head -5)...${NC}"
    fi
    
    assert_contains "$clean_output" "/work" "Should be in /work directory"
}

test_proc_filesystem() {
    local job_id=$(run_python_job "import os; print(len(os.listdir('/proc')))")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    local entries=$(echo "$clean_output" | grep -E '^[0-9]+$' | tail -1)
    
    assert_numeric_le "$entries" 100 "/proc should have limited entries"
}

test_network_interfaces() {
    local job_id=$(run_python_job "
import os
with open('/proc/net/dev') as f:
    lines = f.readlines()
    ifaces = [line.split()[0].rstrip(':') for line in lines[2:] if line.strip()]
    print(f'INTERFACES:{len(ifaces)}:LIST:{ifaces}')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if echo "$clean_output" | grep -q "INTERFACES:"; then
        local iface_count=$(echo "$clean_output" | grep -o 'INTERFACES:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
        
        # Should have exactly 2 interfaces: lo and veth-p-<jobid>
        assert_equals "$iface_count" "2" "Should have exactly 2 network interfaces (lo + veth)"
        assert_contains "$clean_output" "'lo'" "Should have loopback interface"
        assert_contains "$clean_output" "'veth-p-" "Should have isolated veth interface"
    else
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        return 1
    fi
}

test_cgroup_assignment() {
    local job_id=$(run_python_job "
with open('/proc/self/cgroup') as f:
    lines = f.readlines()
    print(f'CGROUPS:{len(lines)}')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "CGROUPS:" "Should be in cgroups"
}

test_resource_limits() {
    local job_id=$("$RNX_BINARY" run --runtime="$DEFAULT_RUNTIME" --max-memory=100 --max-cpu=50 \
        python3 -c "print('RESOURCE_TEST_OK')" | grep "ID:" | awk '{print $2}')
    local logs=$(get_job_logs "$job_id")
    
    assert_contains "$logs" "RESOURCE_TEST_OK" "Should run with resource limits"
}

test_file_operations() {
    local job_id=$(run_python_job "
import os
# Write test file
with open('/work/test.txt', 'w') as f:
    f.write('test_data')
# Read it back
with open('/work/test.txt', 'r') as f:
    print(f'FILE_CONTENT:{f.read()}')
# List directory
print(f'FILES:{os.listdir(\"/work\")}')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "FILE_CONTENT:test_data" "Should write and read files"
    assert_contains "$clean_output" "test.txt" "Should see created file"
}

test_user_isolation() {
    local job_id=$(run_python_job "
import os
print(f'UID:{os.getuid()}:GID:{os.getgid()}:USER:{os.getenv(\"USER\", \"none\")}')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if echo "$clean_output" | grep -q "UID:.*:GID:"; then
        local uid=$(echo "$clean_output" | grep -o 'UID:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
        local gid=$(echo "$clean_output" | grep -o 'GID:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
        
        # Jobs should run as root in isolated namespace
        assert_equals "$uid" "0" "Process should run as UID 0 (root in namespace)"
        assert_equals "$gid" "0" "Process should run as GID 0 (root group in namespace)"
    else
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        return 1
    fi
}

test_filesystem_structure() {
    local job_id=$(run_python_job "
import os
root_dirs = sorted(os.listdir('/'))
print(f'ROOT_DIRS:{root_dirs}')
# Check if host filesystem is NOT visible
if 'home' in root_dirs or 'root' in root_dirs:
    print('HOST_VISIBLE:YES')
else:
    print('HOST_VISIBLE:NO')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    # Should have isolated root filesystem, not host filesystem
    assert_contains "$clean_output" "HOST_VISIBLE:NO" "Host filesystem should not be visible"
    assert_contains "$clean_output" "work" "Should have work directory for job files"
    
    # Should have basic system directories in isolated root
    assert_contains "$clean_output" "'usr'" "Should have /usr in isolated root"
    assert_contains "$clean_output" "'bin'" "Should have /bin in isolated root"
}

test_mount_isolation() {
    local job_id=$(run_python_job "
import os
# Check if specific mounts exist
with open('/proc/mounts') as f:
    mounts = f.readlines()
    mount_targets = [line.split()[1] for line in mounts]
    
print(f'MOUNT_COUNT:{len(mount_targets)}')

# Check for key isolated mounts
readonly_mounts = [m for m in mounts if 'ro,' in m or ',ro' in m]
print(f'READONLY_COUNT:{len(readonly_mounts)}')

# Check for work directory mount
work_mounts = [m for m in mount_targets if '/work' in m]
print(f'WORK_MOUNTS:{len(work_mounts)}')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    if echo "$clean_output" | grep -q "MOUNT_COUNT:"; then
        local mount_count=$(echo "$clean_output" | grep -o 'MOUNT_COUNT:[0-9]*' | cut -d: -f2)
        local readonly_count=$(echo "$clean_output" | grep -o 'READONLY_COUNT:[0-9]*' | cut -d: -f2)
        
        # Should have many mounts (isolated filesystem with specific mounts)
        assert_numeric_le "10" "$mount_count" "Should have multiple mount points for isolation"
        assert_numeric_le "1" "$readonly_count" "Should have readonly mounts for security"
    else
        echo "    ${RED}Clean output: [$clean_output]${NC}"
        return 1
    fi
}

test_cgroup_isolation() {
    local job_id=$(run_python_job "
with open('/proc/self/cgroup') as f:
    cgroup_info = f.read().strip()
    print(f'CGROUP_INFO:{cgroup_info[:150]}')
    
    # Check for joblet-specific cgroup path
    if 'joblet' in cgroup_info.lower():
        print('JOBLET_CGROUP:YES')
    else:
        print('JOBLET_CGROUP:NO')
        
    # Check for job-specific cgroup
    if 'job-' in cgroup_info:
        print('JOB_SPECIFIC:YES')
    else:
        print('JOB_SPECIFIC:NO')
")
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "JOBLET_CGROUP:YES" "Should be in joblet-specific cgroup"
    assert_contains "$clean_output" "JOB_SPECIFIC:YES" "Should be in job-specific cgroup"
    assert_contains "$clean_output" "CGROUP_INFO:" "Should have cgroup information"
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "Core Isolation Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites check failed${NC}"
        exit 1
    fi
    
    # Ensure runtime is available
    ensure_runtime "$DEFAULT_RUNTIME"
    
    # Run tests
    test_section "Process Isolation"
    run_test "True PID 1 namespace isolation" test_pid_isolation
    run_test "Complete process visibility isolation" test_process_count
    
    test_section "Filesystem Isolation"
    run_test "Working directory isolation" test_working_directory
    run_test "/proc filesystem isolation" test_proc_filesystem
    run_test "File operations in /work" test_file_operations
    run_test "Isolated filesystem structure" test_filesystem_structure
    run_test "Mount namespace isolation" test_mount_isolation
    
    test_section "Network Isolation" 
    run_test "Proper network interface isolation" test_network_interfaces
    
    test_section "User and Security Isolation"
    run_test "User/group namespace isolation" test_user_isolation
    
    test_section "Resource Management"
    run_test "Cgroup assignment" test_cgroup_assignment
    run_test "Job-specific cgroup isolation" test_cgroup_isolation
    run_test "Resource limits enforcement" test_resource_limits
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi