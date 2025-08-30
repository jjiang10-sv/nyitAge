# 198. House Robber
# You are a professional robber planning to rob houses along a street. Each house has a certain amount of money stashed, the only constraint stopping you from robbing each of them is that adjacent houses have security systems connected and it will automatically contact the police if two adjacent houses were broken into on the same night.
# Given an integer array nums representing the amount of money of each house, return the maximum amount of money you can rob tonight without alerting the police.
# Example 1:
# Input: nums = [1,2,3,1]
# Output: 4
# Explanation: Rob house 1 (money = 1) and then rob house 3 (money = 3).
# Total amount you can rob = 1 + 3 = 4.
# Example 2:
# Input: nums = [2,7,9,3,1]
# Output: 12
# Explanation: Rob house 1 (money = 2), rob house 3 (money = 9) and rob house 5 (money = 1).
# Total amount you can rob = 2 + 9 + 1 = 12.

from typing import List
class Solution:
    def rob(self, nums: List[int]) -> int:
        if len(nums) == 1:
            return nums[0]
        elif len(nums) == 2:
            return max(nums)
        else:
            j = len(nums)-1
            return max([nums[j] + self.rob(nums[:j-1]), self.rob(nums[:j])])
    
    def rob_1(self, nums:List[int]) -> int:
        memo = [0] * len(nums)
        if len(nums) == 1:
            return nums[0]
        memo[0], memo[1] = nums[0], max(nums[0], nums[1])
        for i in range(2,len(nums)):
            memo[i] = max(memo[i-1], nums[i] + memo(i-2))
        return memo[-1]

    def rob_3(self, nums:List[int]) -> int:
        if len(nums) == 1:
            return nums[0]
        nums[1] = max(nums[0], nums[1])
        for i in range(2, len(nums)):
            nums[i] = max(nums[i-1], nums[i] + nums[i-2])
        return nums[-1]
    
    def rob_4(self, nums:List[int]) -> int:
        if len(nums) == 1:
            return nums[0]
        left_2, left_1 = nums[0], max(nums[0],nums[1])
        for i in range(2, len(nums)):
            tmp = left_1
            left_1 = max(nums[i] + left_2, left_1)
            left_2 = tmp
        return left_1
    

# You are given an array prices where prices[i] is the price of a given stock on the ith day.

# You want to maximize your profit by choosing a single day to buy one stock and choosing a different day in the future to sell that stock.

# Return the maximum profit you can achieve from this transaction. If you cannot achieve any profit, return 0.

 

# Example 1:

# Input: prices = [7,1,5,3,6,4]
# Output: 5
# Explanation: Buy on day 2 (price = 1) and sell on day 5 (price = 6), profit = 6-1 = 5.
# Note that buying on day 2 and selling on day 1 is not allowed because you must buy before you sell.
# Example 2:

# Input: prices = [7,6,4,3,1]
# Output: 0
# Explanation: In this case, no transactions are done and the max profit = 0.

class Solution1:
    # not DP
    def maxProfit(self, prices: List[int]) -> int:
        if len(prices) == 1:
            return 0
        max_profit , min_price = 0, prices[0]
        for i in range(1,len(prices)):
            max_profit = max(prices[i] - min_price, max_profit)
            min_price = min(min_price, prices[i]) 
        return max_profit
    # strict DP -- not quite fitting in this problem.
    def maxProfit(self, prices: List[int]) -> int:
        if len(prices) == 1:
            return 0
        min_price = prices[0]
        prices[0] = 0
        for i in range(1, len(prices)):
            tmp = prices[i]
            # max_profit in ith price
            prices[i] = max(prices[i-1],prices[i] - min_price )
            min_price = min(tmp, min_price)
        return prices[-1]