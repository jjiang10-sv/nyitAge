
# Hint
# You are given an array of strings ideas that represents a list of names to be used in the process of naming a company. The process of naming a company is as follows:

# Choose 2 distinct names from ideas, call them ideaA and ideaB.
# Swap the first letters of ideaA and ideaB with each other.
# If both of the new names are not found in the original ideas, then the name ideaA ideaB (the concatenation of ideaA and ideaB, separated by a space) is a valid company name.
# Otherwise, it is not a valid name.
# Return the number of distinct valid names for the company.

 

# Example 1:

# Input: ideas = ["coffee","donuts","time","toffee"]
# Output: 6
# Explanation: The following selections are valid:
# - ("coffee", "donuts"): The company name created is "doffee conuts".
# - ("donuts", "coffee"): The company name created is "conuts doffee".
# - ("donuts", "time"): The company name created is "tonuts dime".
# - ("donuts", "toffee"): The company name created is "tonuts doffee".
# - ("time", "donuts"): The company name created is "dime tonuts".
# - ("toffee", "donuts"): The company name created is "doffee tonuts".
# Therefore, there are a total of 6 distinct company names.

# The following are some examples of invalid selections:
# - ("coffee", "time"): The name "toffee" formed after swapping already exists in the original array.
# - ("time", "toffee"): Both names are still the same after swapping and exist in the original array.
# - ("coffee", "toffee"): Both names formed after swapping already exist in the original array.
# Example 2:

# Input: ideas = ["lack","back"]
# Output: 0
# Explanation: There are no valid selections. Therefore, 0 is returned.
 

# Constraints:

# 2 <= ideas.length <= 5 * 104
# 1 <= ideas[i].length <= 10
# ideas[i] consists of lowercase English letters.
# All the strings in ideas are unique.

# from typing import List
# class Solution:
#     def distinctNames(self, ideas: List[str]) -> int:

#         prefix_suffix_groups = {}
#         for idea in ideas:
#             first_char = idea[0]
#             rest_char = idea[1:]

#             if first_char in prefix_suffix_groups:
#                 prefix_suffix_groups[first_char].add(rest_char)
#             else:
#                 init_set = set()
#                 init_set.add(rest_char)
#                 prefix_suffix_groups[first_char] = init_set
                
#         total_count = 0
#         keys = list(prefix_suffix_groups.keys())
#         for i in range(0, len(keys)):
#             for j in range(i+1, len(keys)):
#                 unique_i = prefix_suffix_groups[keys[i]] - prefix_suffix_groups[keys[j]]
#                 unique_j = prefix_suffix_groups[keys[j]] - prefix_suffix_groups[keys[i]]
#                 total_count += len(unique_i)*len(unique_j)*2
#         return total_count

from typing import List
from collections import defaultdict

def count_ones(n: int) -> int:
    """Count the number of 1 bits in the binary representation of n."""
    count = 0
    while n:
        count += n & 1  # Add 1 if the least significant bit is 1
        n >>= 1         # Shift right by 1 to check the next bit
    return count
# this solution is not quite right with result way bigger
#  but bit manupilation is worthy of consideration.

class Solution:
    def distinctNames(self, ideas: List[str]) -> int:
        suffix_bitmap = defaultdict(int)
        for idea in ideas:
            first = ord(idea[0]) - ord('a')
            suffix = idea[1:]
            suffix_bitmap[suffix] |= (1 << first)

        # For each pair of starting letters, count unique suffixes
        result = 0
        for i in range(26):
            for j in range(i + 1, 26):
                count_i = count_j = 0
                for bm in suffix_bitmap.values():
                    # Suffix appears with i but not j
                    if (bm & (1 << i)) and not (bm & (1 << j)):
                        count_i += 1
                    # Suffix appears with j but not i
                    if (bm & (1 << j)) and not (bm & (1 << i)):
                        count_j += 1
                result += count_i * count_j * 2  # *2 for both (i, j) and (j, i)
        return result

# class Solution:
#     def distinctNames(self, ideas: List[str]) -> int:
#         suffix_bitmap = defaultdict(int)  # suffix -> 26-bit bitmap

#         for idea in ideas:
#             first = ord(idea[0]) - ord('a')
#             suffix = idea[1:]
#             suffix_bitmap[suffix] |= (1 << first)
#         total = 0
#         keys = list(suffix_bitmap.keys())
#         for i in range(0, len(keys)):
#             for j in range(i+1, len(keys)):
#                 first_letter_a = suffix_bitmap[keys[i]]
#                 first_letter_b = suffix_bitmap[keys[j]]
#                 common = first_letter_a ^ first_letter_b
#                 unique_first_letter_a = first_letter_a & common
#                 unique_first_letter_b = first_letter_b & common
#                 total += count_ones(unique_first_letter_a) * count_ones(unique_first_letter_b) * 2
#         return total


# Example usage:
ideas = ["coffee","donuts","time","toffee"]
sol = Solution()
print(sol.distinctNames(ideas))  # Output: 6



# Example usage:
print(count_ones(0b10111))  # Output: 4
print(count_ones(23))       # Output: 4

# from typing import List
# class Solution:
#     def distinctNames(self, ideas: List[str]) -> int:

#         dict1 , dict2 = {}, {}
#         idea_set = set()
#         for idea in ideas:
#             first_char = idea[0]
#             rest_char = idea[1:]
#             idea_set.add(idea)
#             if first_char in dict1:
#                 dict1[first_char].add(idea)
#             else:
#                 init_set = set()
#                 init_set.add(idea)
#                 dict1[first_char] = init_set
                
#             if rest_char in dict2:
#                 dict2[rest_char].add(idea)
#             else:
#                 init_set = set()
#                 init_set.add(idea)
#                 dict2[rest_char] = init_set
#         total_count = 0
#         #dict1_keys = set(dict1.keys())
#         dict1_keys_len = len(idea_set)
#         for idea in idea_set:
#             #same_prefix_ideas = dict1[idea[0]]
#             dict1_keys_len_tmp = dict1_keys_len
#             same_suffix_ideas, same_prefix_ideas = dict2[idea[1:]], dict1[idea[0]]
#             for idea_tmp in same_suffix_ideas:
#                 dict1_keys_len_tmp -= len(dict1[idea_tmp[0]])
#             # if remove itself from same_prefix_ideas

#             #same_prefix_ideas.remove(idea)
#             for idea_tmp in same_prefix_ideas:
#                 # not including the idea itself for it has been subtracted 
#                 if idea_tmp != idea:
#                     dict1_keys_len_tmp -= (len(dict2[idea_tmp[1:]])-1)
#             print(idea, dict1_keys_len_tmp)
#             total_count += dict1_keys_len_tmp
#             # not_count_first_char = set(idea[0])
#             # for idea in same_suffix_ideas:
#             #     not_count_first_char.add(idea[0])
#             # for idea in same_prefix_ideas:
#             #     not_count_first_char.add(idea[0])
#             # valid_pair_key_set = dict1_keys - not_count_first_char
#             # for first_char in valid_pair_key_set:
#             #     total_count += len(dict1[first_char])
#             # for first_char in dict1.keys():
#             #     if first_char not in not_count_first_char:
#             #         total_count += len(dict1[first_char])
#         if total_count < 0:
#             total_count *= -1
#         return total_count