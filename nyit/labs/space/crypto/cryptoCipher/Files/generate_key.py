#!/bin/env python3
import random

s = "abcdefghijklmnopqrstuvwxyz"
list = random.sample(s, len(s))
key = ''.join(list)
print(key)

#onxbwfevuzijdgrqlpacstmykh

# tr "abcdefghijklmnopqrstuvwxyz" "onxbwfevuzijdgrqlpacstmykh" < plaintext.txt > ciphertext1.txt