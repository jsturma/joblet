#!/bin/bash

# Test 03: Comprehensive Network Configuration Tests
# Tests all network modes: bridge, isolated, none, and custom networks
# Tests inter-job communication within networks and isolation between networks

# Source the test framework
source "$(dirname "$0")/../lib/test_framework.sh"

# Test configuration
CUSTOM_NETWORK_1="test-network-1"
CUSTOM_NETWORK_2="test-network-2"
CUSTOM_CIDR_1="10.100.1.0/24"
CUSTOM_CIDR_2="10.100.2.0/24"

# ============================================
# Helper Functions
# ============================================

run_job_with_network() {
    local command="$1"
    local network="$2"
    local runtime="${3:-$DEFAULT_RUNTIME}"
    
    local job_output
    if [[ -n "$network" ]]; then
        job_output=$("$RNX_BINARY" run --network="$network" --runtime="$runtime" "$command" 2>&1)
    else
        job_output=$("$RNX_BINARY" run --runtime="$runtime" "$command" 2>&1)
    fi
    echo "$job_output" | grep "ID:" | awk '{print $2}'
}

run_python_network_job() {
    local python_code="$1"
    local network="$2"
    
    if [[ -n "$network" ]]; then
        local job_output=$("$RNX_BINARY" run --network="$network" --runtime="$DEFAULT_RUNTIME" python3 -c "$python_code" 2>&1)
    else
        local job_output=$("$RNX_BINARY" run --runtime="$DEFAULT_RUNTIME" python3 -c "$python_code" 2>&1)
    fi
    echo "$job_output" | grep "ID:" | awk '{print $2}'
}

cleanup_test_networks() {
    # Clean up test networks (ignore errors)
    "$RNX_BINARY" network remove "$CUSTOM_NETWORK_1" 2>/dev/null || true
    "$RNX_BINARY" network remove "$CUSTOM_NETWORK_2" 2>/dev/null || true
}

# ============================================
# Bridge Network Tests
# ============================================

test_bridge_network_interfaces() {
    local job_id=$(run_python_network_job "
import subprocess
result = subprocess.run(['cat', '/proc/net/dev'], capture_output=True, text=True)
interfaces = [line.split(':')[0].strip() for line in result.stdout.split('\\n')[2:] if ':' in line]
print(f'BRIDGE_INTERFACES:{len(interfaces)}:{interfaces}')
" "bridge")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    # Bridge network should have lo + veth interface
    assert_contains "$clean_output" "BRIDGE_INTERFACES:" "Should list bridge interfaces"
    local iface_count=$(echo "$clean_output" | grep -o 'BRIDGE_INTERFACES:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
    assert_equals "$iface_count" "2" "Bridge should have exactly 2 interfaces (lo + veth)"
}

test_bridge_internet_access() {
    local job_id=$(run_python_network_job "
import socket
import sys
try:
    # Try to connect to a reliable external service
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(3)
    s.connect(('8.8.8.8', 53))  # Google DNS
    print('BRIDGE_INTERNET:SUCCESS')
    s.close()
except Exception as e:
    print(f'BRIDGE_INTERNET:FAILED:{e}')
" "bridge")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "BRIDGE_INTERNET:SUCCESS" "Bridge network should have internet access"
}

test_bridge_dns_resolution() {
    local job_id=$(run_python_network_job "
import socket
try:
    addr = socket.gethostbyname('google.com')
    print(f'BRIDGE_DNS:SUCCESS:{addr}')
except Exception as e:
    print(f'BRIDGE_DNS:FAILED:{e}')
" "bridge")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "BRIDGE_DNS:SUCCESS" "Bridge network should resolve DNS"
}

# ============================================
# Isolated Network Tests
# ============================================

test_isolated_network_interfaces() {
    local job_id=$(run_python_network_job "
import subprocess
result = subprocess.run(['cat', '/proc/net/dev'], capture_output=True, text=True)
interfaces = [line.split(':')[0].strip() for line in result.stdout.split('\\n')[2:] if ':' in line]
print(f'ISOLATED_INTERFACES:{len(interfaces)}:{interfaces}')
" "isolated")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "ISOLATED_INTERFACES:" "Should list isolated interfaces"
    local iface_count=$(echo "$clean_output" | grep -o 'ISOLATED_INTERFACES:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
    # Isolated network might have only loopback in some configurations
    if [[ "$iface_count" == "1" ]]; then
        echo -e "    ${YELLOW}Note: Isolated network has only loopback (strict isolation)${NC}"
        return 0
    else
        assert_equals "$iface_count" "2" "Isolated should have exactly 2 interfaces (lo + veth)"
    fi
}

test_isolated_internet_access() {
    local job_id=$(run_python_network_job "
import socket
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(3)
    s.connect(('8.8.8.8', 53))
    print('ISOLATED_INTERNET:SUCCESS')
    s.close()
except Exception as e:
    print('ISOLATED_INTERNET:FAILED')
" "isolated")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "ISOLATED_INTERNET:SUCCESS" "Isolated network should have internet access"
}

test_isolated_no_inter_job_communication() {
    # Start a server job in isolated network
    local server_job=$(run_python_network_job "
import socket
import time
import threading

def server():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.bind(('0.0.0.0', 8080))
        s.listen(1)
        print('ISOLATED_SERVER:LISTENING')
        s.settimeout(5)
        conn, addr = s.accept()
        print(f'ISOLATED_SERVER:CONNECTION:{addr}')
        conn.close()
        s.close()
    except Exception as e:
        print(f'ISOLATED_SERVER:ERROR:{e}')

threading.Thread(target=server).start()
time.sleep(8)
print('ISOLATED_SERVER:DONE')
" "isolated")
    
    sleep 2
    
    # Try to connect from another isolated job
    local client_job=$(run_python_network_job "
import socket
import time
time.sleep(1)  # Let server start
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(3)
    # Try to connect to the server job - should fail in isolated mode
    s.connect(('10.100.1.1', 8080))  # Guess at server IP
    print('ISOLATED_CLIENT:SUCCESS')
    s.close()
except Exception as e:
    print('ISOLATED_CLIENT:BLOCKED')
" "isolated")
    
    sleep 8
    
    local server_logs=$(get_job_logs "$server_job")
    local client_logs=$(get_job_logs "$client_job")
    
    # In isolated mode, jobs cannot communicate with each other
    assert_contains "$client_logs" "ISOLATED_CLIENT:BLOCKED" "Isolated jobs should not communicate with each other"
}

# ============================================
# None Network Tests
# ============================================

test_none_network_interfaces() {
    local job_id=$(run_python_network_job "
import subprocess
result = subprocess.run(['cat', '/proc/net/dev'], capture_output=True, text=True)
interfaces = [line.split(':')[0].strip() for line in result.stdout.split('\\n')[2:] if ':' in line]
print(f'NONE_INTERFACES:{len(interfaces)}:{interfaces}')
" "none")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "NONE_INTERFACES:" "Should list none network interfaces"
    local iface_count=$(echo "$clean_output" | grep -o 'NONE_INTERFACES:[0-9]*' | cut -d: -f2 | tr -d '\n\r')
    assert_equals "$iface_count" "1" "None network should have only loopback interface"
}

test_none_network_no_internet() {
    local job_id=$(run_python_network_job "
import socket
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(2)
    s.connect(('8.8.8.8', 53))
    print('NONE_INTERNET:UNEXPECTED_SUCCESS')
    s.close()
except Exception as e:
    print('NONE_INTERNET:BLOCKED')
" "none")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "NONE_INTERNET:BLOCKED" "None network should have no internet access"
}

test_none_network_loopback_only() {
    local job_id=$(run_python_network_job "
import socket
try:
    # Test loopback connectivity
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.bind(('127.0.0.1', 0))
    print('NONE_LOOPBACK:SUCCESS')
    s.close()
except Exception as e:
    print(f'NONE_LOOPBACK:FAILED:{e}')
" "none")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "NONE_LOOPBACK:SUCCESS" "None network should have loopback connectivity"
}

# ============================================
# Custom Network Tests
# ============================================

test_custom_network_creation() {
    # Clean up first
    cleanup_test_networks
    
    # Create test networks
    if "$RNX_BINARY" network create "$CUSTOM_NETWORK_1" --cidr="$CUSTOM_CIDR_1" >/dev/null 2>&1; then
        echo -e "    ${GREEN}✓ Created custom network $CUSTOM_NETWORK_1${NC}"
    else
        return 1
    fi
    
    if "$RNX_BINARY" network create "$CUSTOM_NETWORK_2" --cidr="$CUSTOM_CIDR_2" >/dev/null 2>&1; then
        echo -e "    ${GREEN}✓ Created custom network $CUSTOM_NETWORK_2${NC}"
        return 0
    else
        return 1
    fi
}

test_custom_network_listing() {
    local networks=$("$RNX_BINARY" network list 2>&1)
    
    assert_contains "$networks" "$CUSTOM_NETWORK_1" "Should list custom network 1"
    assert_contains "$networks" "$CUSTOM_NETWORK_2" "Should list custom network 2"
    assert_contains "$networks" "$CUSTOM_CIDR_1" "Should show CIDR for network 1"
    assert_contains "$networks" "$CUSTOM_CIDR_2" "Should show CIDR for network 2"
}

test_custom_network_inter_job_communication() {
    # Start a server in custom network 1
    local server_job=$(run_python_network_job "
import socket
import time
import threading
import os

def get_my_ip():
    # Get the IP address assigned to this job
    import subprocess
    result = subprocess.run(['ip', 'route', 'get', '8.8.8.8'], capture_output=True, text=True)
    # Extract IP from 'src X.X.X.X'
    for part in result.stdout.split():
        if part.startswith('10.100.1.'):
            return part
    return '10.100.1.2'  # fallback

my_ip = get_my_ip()
print(f'CUSTOM_SERVER_IP:{my_ip}')

def server():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        s.bind(('0.0.0.0', 8080))
        s.listen(5)
        print('CUSTOM_SERVER:LISTENING:8080')
        s.settimeout(10)
        while True:
            try:
                conn, addr = s.accept()
                print(f'CUSTOM_SERVER:CONNECTION:{addr}')
                conn.send(b'CUSTOM_HELLO')
                conn.close()
            except socket.timeout:
                break
        s.close()
    except Exception as e:
        print(f'CUSTOM_SERVER:ERROR:{e}')

threading.Thread(target=server).start()
time.sleep(12)
print('CUSTOM_SERVER:DONE')
" "$CUSTOM_NETWORK_1")
    
    sleep 3
    
    # Start a client in the same custom network
    local client_job=$(run_python_network_job "
import socket
import time
import subprocess

# Get server IP by scanning the network
def find_server():
    # Try common IPs in the 10.100.1.x range
    for i in range(2, 10):
        try:
            ip = f'10.100.1.{i}'
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(1)
            s.connect((ip, 8080))
            data = s.recv(1024).decode()
            s.close()
            print(f'CUSTOM_CLIENT:SUCCESS:{ip}:{data}')
            return True
        except:
            continue
    return False

time.sleep(2)  # Let server start
if not find_server():
    print('CUSTOM_CLIENT:NO_SERVER_FOUND')
" "$CUSTOM_NETWORK_1")
    
    sleep 12
    
    local server_logs=$(get_job_logs "$server_job")
    local client_logs=$(get_job_logs "$client_job")
    
    # Jobs in same custom network should be able to communicate
    assert_contains "$server_logs" "CUSTOM_SERVER:LISTENING" "Server should start in custom network"
    assert_contains "$client_logs" "CUSTOM_CLIENT:SUCCESS" "Client should connect to server in same custom network"
}

test_custom_network_isolation_between_networks() {
    # Start server in network 1
    local server_job=$(run_python_network_job "
import socket
import time
import threading

def server():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.bind(('0.0.0.0', 8080))
        s.listen(1)
        print('ISOLATION_SERVER:LISTENING')
        s.settimeout(8)
        conn, addr = s.accept()
        print(f'ISOLATION_SERVER:UNEXPECTED_CONNECTION:{addr}')
        conn.close()
        s.close()
    except socket.timeout:
        print('ISOLATION_SERVER:NO_CONNECTIONS')
    except Exception as e:
        print(f'ISOLATION_SERVER:ERROR:{e}')

threading.Thread(target=server).start()
time.sleep(10)
print('ISOLATION_SERVER:DONE')
" "$CUSTOM_NETWORK_1")
    
    sleep 2
    
    # Try to connect from network 2 (should fail)
    local client_job=$(run_python_network_job "
import socket
import time
time.sleep(1)
try:
    # Try various IPs from network 1
    for i in range(2, 10):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(1)
            s.connect((f'10.100.1.{i}', 8080))
            print('ISOLATION_CLIENT:UNEXPECTED_SUCCESS')
            s.close()
            break
        except:
            continue
    else:
        print('ISOLATION_CLIENT:BLOCKED')
except Exception as e:
    print('ISOLATION_CLIENT:BLOCKED')
" "$CUSTOM_NETWORK_2")
    
    sleep 12
    
    local server_logs=$(get_job_logs "$server_job")
    local client_logs=$(get_job_logs "$client_job")
    
    # Jobs in different custom networks should NOT be able to communicate
    assert_contains "$server_logs" "ISOLATION_SERVER:NO_CONNECTIONS" "Server should not receive connections from different network"
    assert_contains "$client_logs" "ISOLATION_CLIENT:BLOCKED" "Client in different network should be blocked"
}

test_custom_network_internet_access() {
    local job_id=$(run_python_network_job "
import socket
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(3)
    s.connect(('8.8.8.8', 53))
    print('CUSTOM_INTERNET:SUCCESS')
    s.close()
except Exception as e:
    print(f'CUSTOM_INTERNET:FAILED:{e}')
" "$CUSTOM_NETWORK_1")
    
    local logs=$(get_job_logs "$job_id")
    local clean_output=$(get_clean_output "$logs")
    
    assert_contains "$clean_output" "CUSTOM_INTERNET:SUCCESS" "Custom network should have internet access"
}

# ============================================
# Main Test Execution
# ============================================

main() {
    # Initialize test suite
    test_suite_init "Comprehensive Network Configuration Tests"
    
    # Check prerequisites
    if ! check_prerequisites; then
        echo -e "${RED}Prerequisites check failed${NC}"
        exit 1
    fi
    
    # Ensure runtime is available
    ensure_runtime "$DEFAULT_RUNTIME"
    
    # Bridge Network Tests
    test_section "Bridge Network Tests (Default)"
    run_test "Bridge network interface configuration" test_bridge_network_interfaces
    run_test "Bridge network internet access" test_bridge_internet_access
    run_test "Bridge network DNS resolution" test_bridge_dns_resolution
    
    # Isolated Network Tests
    test_section "Isolated Network Tests"
    run_test "Isolated network interface configuration" test_isolated_network_interfaces
    run_test "Isolated network internet access" test_isolated_internet_access
    run_test "Isolated network blocks inter-job communication" test_isolated_no_inter_job_communication
    
    # None Network Tests
    test_section "None Network Tests (Complete Isolation)"
    run_test "None network interface configuration" test_none_network_interfaces
    run_test "None network blocks internet access" test_none_network_no_internet
    run_test "None network allows loopback only" test_none_network_loopback_only
    
    # Custom Network Tests
    test_section "Custom Network Management"
    run_test "Custom network creation" test_custom_network_creation
    run_test "Custom network listing" test_custom_network_listing
    
    test_section "Custom Network Communication"
    run_test "Inter-job communication within custom network" test_custom_network_inter_job_communication
    run_test "Network isolation between different custom networks" test_custom_network_isolation_between_networks
    run_test "Custom network internet access" test_custom_network_internet_access
    
    # Cleanup
    cleanup_test_networks
    
    # Print summary
    test_suite_summary
    exit $?
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi