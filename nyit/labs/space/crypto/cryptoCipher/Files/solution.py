#!/usr/bin/python3
from sys import argv
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad

def read_keys(file_path):
    with open(file_path) as f:
        return [line.rstrip('\n') for line in f]

def find_key(data, ciphertext, iv, keys):
    for k in keys:
        if len(k) <= 16:
            key = k + '#' * (16 - len(k))  # Pad key to 16 bytes
            cipher = AES.new(key.encode('utf-8'), AES.MODE_CBC, iv)
            guess = cipher.encrypt(pad(data, 16))
            if guess == ciphertext:
                return key
    return None

def main():
    if len(argv) != 4:
        print("Usage: ./program <first> <second> <third>")
        return

    first, second, third = argv[1], argv[2], argv[3]

    assert len(first) == 21, "First argument must be 21 characters long."
    data = bytearray(first, encoding='utf-8')
    ciphertext = bytearray.fromhex(second)
    iv = bytearray.fromhex(third)

    keys = read_keys('./words.txt')
    key_found = find_key(data, ciphertext, iv, keys)

    if key_found:
        print("find the key:", key_found)
    else:
        print("cannot find the key!")

if __name__ == "__main__":
    main()