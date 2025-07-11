#!/usr/bin/env python3
"""
Concurrent Rate Limiting Test for Solana Balance API
Tests IP-based rate limiting with concurrent requests
"""

import requests
import time
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed

API_URL = "http://localhost:8080/api/get-balance"
API_KEY = "test-api-key-1"
TEST_ADDRESS = "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"

def make_request(request_id, ip=None):
    """Make a single API request"""
    headers = {
        "Content-Type": "application/json",
        "X-API-Key": API_KEY
    }
    
    if ip:
        headers["X-Forwarded-For"] = ip
    
    data = {"wallets": [TEST_ADDRESS]}
    
    try:
        response = requests.post(API_URL, headers=headers, json=data, timeout=10)
        result = response.json()
        
        if result.get("success"):
            return f"Request {request_id}: SUCCESS (Balance: {result.get('balance', 'N/A')})"
        elif "Rate limit exceeded" in result.get("error", ""):
            return f"Request {request_id}: RATE LIMITED"
        else:
            return f"Request {request_id}: ERROR - {result.get('error', 'Unknown error')}"
    except Exception as e:
        return f"Request {request_id}: EXCEPTION - {str(e)}"

def test_single_ip_rate_limit():
    """Test rate limiting for a single IP"""
    print("🧪 Testing Single IP Rate Limiting")
    print("=" * 40)
    
    # Make 10 requests (should all succeed)
    print("\n1. Making 10 requests (should all succeed):")
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = [executor.submit(make_request, i) for i in range(1, 11)]
        for future in as_completed(futures):
            print(f"  {future.result()}")
    
    # Make 11th request (should be rate limited)
    print("\n2. Making 11th request (should be rate limited):")
    result = make_request(11)
    print(f"  {result}")

def test_multiple_ips():
    """Test rate limiting with multiple IPs"""
    print("\n🌐 Testing Multiple IP Rate Limiting")
    print("=" * 40)
    
    ips = ["192.168.1.100", "192.168.1.200", "10.0.0.1", "172.16.0.1"]
    
    for ip in ips:
        print(f"\nTesting IP: {ip}")
        with ThreadPoolExecutor(max_workers=3) as executor:
            futures = [executor.submit(make_request, i, ip) for i in range(1, 6)]
            for future in as_completed(futures):
                print(f"  {future.result()}")

def test_concurrent_same_ip():
    """Test concurrent requests from same IP"""
    print("\n⚡ Testing Concurrent Requests from Same IP")
    print("=" * 40)
    
    print("Making 15 concurrent requests from same IP:")
    with ThreadPoolExecutor(max_workers=15) as executor:
        futures = [executor.submit(make_request, i) for i in range(1, 16)]
        
        success_count = 0
        rate_limited_count = 0
        
        for future in as_completed(futures):
            result = future.result()
            print(f"  {result}")
            
            if "SUCCESS" in result:
                success_count += 1
            elif "RATE LIMITED" in result:
                rate_limited_count += 1
        
        print(f"\nResults:")
        print(f"  Successful: {success_count}")
        print(f"  Rate Limited: {rate_limited_count}")
        print(f"  Expected: 10 successful, 5 rate limited")

def test_rate_limit_reset():
    """Test rate limit reset after waiting"""
    print("\n⏰ Testing Rate Limit Reset")
    print("=" * 40)
    
    # Exhaust the rate limit
    print("1. Exhausting rate limit (making 10 requests):")
    for i in range(1, 11):
        result = make_request(i)
        print(f"  {result}")
    
    # Try 11th request (should be rate limited)
    print("\n2. Making 11th request (should be rate limited):")
    result = make_request(11)
    print(f"  {result}")
    
    # Wait and try again
    print("\n3. Waiting 5 seconds and trying again:")
    time.sleep(5)
    result = make_request(12)
    print(f"  {result}")
    
    print("\nNote: Rate limit resets after 1 minute. Wait 60 seconds for full reset.")

if __name__ == "__main__":
    print("🚀 Solana Balance API - Rate Limiting Tests")
    print("=" * 50)
    
    try:
        # Test health endpoint first
        response = requests.get("http://localhost:8080/health")
        if response.status_code == 200:
            print("✅ API is running and healthy")
        else:
            print("❌ API is not responding properly")
            exit(1)
    except Exception as e:
        print(f"❌ Cannot connect to API: {e}")
        print("Make sure the API server is running with: go run main_simple.go")
        exit(1)
    
    # Run tests
    test_single_ip_rate_limit()
    test_multiple_ips()
    test_concurrent_same_ip()
    test_rate_limit_reset()
    
    print("\n✅ All rate limiting tests completed!") 