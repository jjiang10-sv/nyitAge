import subprocess

# Define input and output files
input_file = "big_file.txt"
key = "00112233445566778889aabbccddeeff"
iv = "0102030405060708090a0b0c0d0e0f10"

# Function to encrypt using OpenSSL
def encrypt(mode):
    command = [
        "openssl", "enc", f"-aes-128-{mode}", "-e",
        "-in", input_file, "-out", f"utput_{mode}.bin",
        "-K", key, "-iv", iv
    ]
    subprocess.run(command, check=True)
#Encrypt using AES-128-CBC
modes = ["cbc","cfb","ofb"]
for mode in modes:
    encrypt(mode=mode)
modes_no_iv = ["ecb"]

print("Encryption completed for all modes.")