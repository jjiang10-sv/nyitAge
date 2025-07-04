#!/bin/bash

# Define input and output files
INPUT_FILE="big_file.txt"
KEY="00112233445566778889aabbccddeeff"
IV="0102030405060708090a0b0c0d0e0f10"

# Encrypt using AES-128-CBC
openssl enc -aes-128-cbc -e -in "$INPUT_FILE" -out output_cbc.bin -K "$KEY" -iv "$IV"

# Encrypt using AES-128-CFB
openssl enc -aes-128-cfb -e -in "$INPUT_FILE" -out output_cfb.bin -K "$KEY" -iv "$IV"

# Encrypt using AES-128-OFB
openssl enc -aes-128-ofb -e -in "$INPUT_FILE" -out output_ofb.bin -K "$KEY" -iv "$IV"

echo "Encryption completed for all modes."