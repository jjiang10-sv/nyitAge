class BinaryTree:
    def __init__(self, compare_fun):
        self.node = None
        self.left = None
        self.right = None
        self.less_fun = compare_fun

    def search(self, value):
        if self.node is None:
            return None
        if self.node == value:
            return self
        if self.less_fun(value, self.node):
            return self.left.search(value) if self.left else None
        else:
            return self.right.search(value) if self.right else None

    def insert(self, value):
        if self.node is None:
            self.node = value
            self.left = BinaryTree(self.less_fun)
            self.right = BinaryTree(self.less_fun)
            return
        if self.less_fun(value, self.node):
            self.left.insert(value)
        else:
            self.right.insert(value)

    def max(self):
        if self.node is None or self.right is None or self.right.node is None:
            return self.node
        return self.right.max()

    def min(self):
        if self.node is None or self.left is None or self.left.node is None:
            return self.node
        return self.left.min()
    
    def find_largest_smaller_key(self, num):
        tmp = self
        while tmp is not None:
            if tmp.right is not None:
                if tmp.right > num:
                    tmp = tmp.left
                elif tmp.right == num:
                    return tmp.node
                else:
                    tmp = tmp.right
            else:
                return tmp.node
            

# Absolute Value Sort
# Given an array of integers arr, write a function absSort(arr), 
# that sorts the array according to the absolute values of the numbers in arr. 
# If two numbers have the same absolute value, sort them according to sign, 
# where the negative numbers come before the positive numbers.

# Examples:

# input:  arr = [2, -7, -2, -2, 0]
# output: [0, -2, -2, 2, -7]
# Constraints:


class BinarySortTree:
    def __init__(self):
        self.node = []
        self.left = None
        self.right = None
        self.sorted_arr = []
    
    def insert(self, num):
        if len(self.node) == 0:
            self.node = [num]
            self.left = BinarySortTree()
            self.right = BinarySortTree()
        else:
            node_val = abs(self.node[0])
            insert_val = abs(num)
            if node_val == insert_val:
                if num < 0:
                    self.node = [num] + self.node
                else:
                    self.node.append(num)
            elif node_val > insert_val:
                self.left.insert(num)
            elif node_val < insert_val:
                self.right.insert(num)

    def in_order_walk(self, root):
        #result = []
        if root is not None:
            #return
            #print(root.node)
            self.in_order_walk(root.left)
            #result.append(self.node)
            for c in root.node:
                self.sorted_arr.append(c)
            self.in_order_walk(root.right)

# arr = [2, -7, -2, -2, 0,3,-1,3,-3,1,0,-5,6,6,5,-6]
# root = BinarySortTree()
# for c in arr:
#     root.insert(c)
# root.in_order_walk(root=root)
# print(root.sorted_arr)

# Input: s = "abcabcbb"
# Output: 3
# Explanation: The answer is "abc", with the length of 3.

# Input: s = "bbbbb"
# Output: 1
# Explanation: The answer is "b", with the length of 1.

# Input: s = ""
# Output: 0
# Explanation: The string is empty, so the answer is 0.



#         Array of Array Products
# Given an array of integers arr, you’re asked to calculate for each index i the product of all integers 
# except the integer at that index (i.e. except arr[i]). Implement a function arrayOfArrayProducts 
# that takes an array of integers and returns an array of the products.

# Solve without using division and analyze your solution's time and space complexities.

# Examples:

# input:  arr = [8, 10, 2]
# output: [20, 16, 80] # by calculating: [10*2, 8*2, 8*10]

# input:  arr = [2, 7, 3, 4]
# output: [84, 24, 56, 42] # by calculating: [7*3*4, 2*3*4, 2*7*4, 2*7*3]
# Constraints:

# [time limit] 5000ms

# [input] array.integer arr

# 0 ≤ arr.length ≤ 20
# [output] array.integer

def arrayOfArrayProducts(arr :list[int]):
    
    n = len(arr)
    if n <= 1:
        return []
    
    first, second = arr[0], arr[1]
    result = [second, first]
    for i in range(2,n):
        to_add_item = arr[i]

        last_item_in_result = result[len(result)-1]
        item_before_to_add = arr[i-1]
        result = [item * to_add_item for item in result]
        result.append(last_item_in_result*item_before_to_add)
    return result

arr = [8, 10, 2]
result = arrayOfArrayProducts(arr=arr)
print(result)
    # to_add_idx = 2
    # while to_add_idx < n:
    #     result = 

# (2,3)
# (2,3,4) 

#     2   7   3   4

# 2   1   14   6   8

# 7   14  21  28

# 3   42    

# 4

#     7   3   4
# 7       21  28

# 3   28 * 7  

# 4

from typing import List
class Solution:
    def getLongestSubsequence(self, words: List[str], groups: List[int]) -> List[str]:
        indexes = [0]

        for idx, group in range(1,len(groups)):
            print(idx, group)
            if indexes[-1] != group:
                indexes.append(idx)
        
        return [words[i] for i in indexes]

        