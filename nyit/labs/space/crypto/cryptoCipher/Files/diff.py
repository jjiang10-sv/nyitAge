#!/usr/bin/python3

decryted_files = ['decrypted_ecb.txt', 'decrypted_cbc.txt','decrypted_cfb.txt','decrypted_ofb.txt']

for file in decryted_files:

    with open('big_file.txt', 'rb') as f:
        f1 = f.read()
    with open(file, 'rb') as f:
        f2 = f.read()
    res = 0
    for i in range(min(len(f1), len(f2))):
        if f1[i] != f2[i]:
            res += 1
    print( file + " diff bytes: "+str(res+abs(len(f1)-len(f2))))