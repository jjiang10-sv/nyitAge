#!/usr/bin/env python3

from collections import Counter
import re

TOP_K  = 20
N_GRAM = 3

# Generate all the n-grams for value n
def ngrams(n, text):
    for i in range(len(text) -n + 1):
        # Ignore n-grams containing white space
        if not re.search(r'\s', text[i:i+n]):
           yield text[i:i+n]

# Read the data from the ciphertext
with open('ciphertext.txt') as f:
    text = f.read()

# Count, sort, and print out the n-grams
for N in range(N_GRAM):
   print("-------------------------------------")
   print("{}-gram (top {}):".format(N+1, TOP_K))
   counts = Counter(ngrams(N+1, text))        # Count
   sorted_counts = counts.most_common(TOP_K)  # Sort 
   for ngram, count in sorted_counts:                  
       print("{}: {}".format(ngram, count))   # Print
   print("===========", N)
   if N == 0:
       ngram_string = ""  # Initialize an empty string to store ngrams
       for ngram, _ in sorted_counts:
          ngram_string += ngram + ""  # Append each ngram to the string
       print(ngram_string) 

# -------------------------------------
# 1-gram (top 20):
# n: 488
# y: 373
# v: 348
# x: 291
# u: 280
# q: 276
# m: 264
# h: 235
# t: 183
# i: 166
# p: 156
# a: 116
# c: 104
# z: 95
# l: 90
# g: 83
# b: 83
# r: 82
# e: 76
# d: 59
# -------------------------------------
# 2-gram (top 20):
# yt: 115
# tn: 89
# mu: 74
# nh: 58
# vh: 57
# hn: 57
# vu: 56
# nq: 53
# xu: 52
# up: 46
# xh: 45
# yn: 44
# np: 44
# vy: 44
# nu: 42
# qy: 39
# vq: 33
# vi: 32
# gn: 32
# av: 31
# -------------------------------------
# 3-gram (top 20):
# ytn: 78
# vup: 30
# mur: 20
# ynh: 18
# xzy: 16
# mxu: 14
# gnq: 14
# ytv: 13
# nqy: 13
# vii: 13
# bxh: 13
# lvq: 12
# nuy: 12
# vyn: 12
# uvy: 11
# lmu: 11
# nvh: 11
# cmu: 11
# tmq: 10
# vhp: 10