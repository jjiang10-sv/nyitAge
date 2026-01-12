# queue.py

class QueueItem:
    def __init__(self, item, prev=None):
        self.item = item
        self.prev = prev

class Queue:
    def __init__(self):
        self.current = None
        self.last = None
        self.depth = 0

    def enqueue(self, item):
        if self.depth == 0:
            self.current = QueueItem(item)
            self.last = self.current
            self.depth += 1
            return
        q = QueueItem(item)
        self.last.prev = q
        self.last = q
        self.depth += 1

    def dequeue(self):
        if self.depth > 0:
            item = self.current.item
            self.current = self.current.prev
            self.depth -= 1
            return item
        return None

# Forward-linked version
class QueueItem1:
    def __init__(self, item, next=None):
        self.item = item
        self.next = next

class Queue1:
    def __init__(self):
        self.current = None
        self.last = None
        self.depth = 0

    def enqueue(self, item):
        if self.depth == 0:
            self.current = QueueItem1(item)
            self.last = self.current
            self.depth += 1
            return
        q = QueueItem1(item)
        self.last.next = q
        self.last = q
        self.depth += 1

    def dequeue(self):
        if self.depth == 0:
            return None
        item = self.current.item
        self.current = self.current.next
        self.depth -= 1
        return item
    


# from typing import List

# def get_indices_of_item_weights(arr: List[int], limit: int) -> List[int]:
#     # convert then list into an map, then acccess it by the k,v pair. the time complexity of
#     # accessing map is O(1); the value is an array containing the indexes of the key
#     items_map = {}
#     print("debugging....")
#     for i in range(len(arr)):
#         item = arr[i]
#         if item not in items_map:
#             items_map[item]=  [i]
#         else:
#             items_map[item].append(i)
#         i +=1
#     print(items_map)
#     for k,v in items_map.items():
        
#         remainer = limit - k
#         print("the remainder is ", remainer)
#         if remainer in items_map:
#             if remainer == k:
#                 return [v[0],v[1]]
#             else:
#                 return [v[0], items_map[remainer][0]]
#     return []
    
# arr = [4, 6, 10, 15, 16]
# lim = 21
# target = 15
# # result = get_indices_of_item_weights(arr=arr,limit=lim)
# # print(result)

# # from typing import List
# from typing import List

# def shifted_arr_search(shiftArr: List[int], num: int) -> int:
#     # find the shift val by cutting in the middle
    
#     low, high = 0, len(shiftArr)-1
#     mid = (low + high)//2
#     original_arr = None
#     while mid < high and mid > 0:
#         mid_val = shiftArr[mid]
#         left_val = shiftArr[mid-1]
#         print(mid_val,left_val, mid)
#         if mid_val < left_val :
#             original_arr =   shiftArr[mid:] + shiftArr[:mid]
#             break
#         elif mid_val > left_val:
#             # 4,5,6,7,8,1,2,3
#             mid += 1
#     print("the original array is ", original_arr)
#     return (binay_search(original_arr, num) + mid) % (high+1)



# def binay_search(arr: List[int], num:int) -> int:

#     n = len(arr)
#     low,high = 0, n-1
#     while low <= high:
#         mid = (low + high) // 2
#         print("low , mid , high ", low, mid, high)
#         print(arr[mid], num)
#         if high - low <=2:
#             for i in range(low,high+1):
#                 if arr[i] == num:
#                     return i
#         if arr[mid] == num:
#             return mid
#         elif arr[mid] > num:
#             high = mid -1
            
#         elif arr[mid] < num:
#             low = mid +1
#         print("low , high ", low, high)
#     return None

# # shiftArr = [2]
# # num = 2 # shiftArr is the
# # #[2],2
# # print(shifted_arr_search(shiftArr=shiftArr, num=num))





# def find_pairs(arr: list, num :int):
#     item_map = {}
#     ans = []
#     for i in range(len(arr)):
#         tmp = []
#         item = arr[i]
#         pair = item - num
#         item_map[item] = i
#         if pair in item_map:
#             #if i == (len(arr) - 1):

#             tmp = [item_map[pair], i]
#             ans.append(tmp)
#     return ans
      

def get_indices_of_item_weights(arr):
    
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

# input:  arr = [8, 10, 2]
# output: [20, 16, 80] # by calculating: [10*2, 8*2, 8*10]

# input:  arr = [2, 7, 3, 4]
# output: [84, 24, 56, 42] # by calculating: [7*3*4, 2*3*4, 2*7*4, 2*7*3]
# Constraints:
# arr = [2, 7, 3, 4]
# result = get_indices_of_item_weights(arr=arr)
# print(result)



def get_longest_substring(s):
    if len(s) == 0 :
        return 0
    visited = set(s[0])
    longest = 1
    
    l,r = 0,1
    while  r < len(s):
        char = s[r]
        if char not in visited:
            visited.add(s[r])
        else:
            # set("a","b","c")  b
            # "ab"  set("a") b
            longest = max(longest, len(visited))
            while char in visited:
                visited.remove(s[l])
                l += 1
            visited.add(char)
        r += 1
    return max(longest, len(visited))


def test_get_longest_substring():
    # Test 1: Empty string
    assert get_longest_substring("") == 0

    # Test 2: Single character
    assert get_longest_substring("a") == 1

    # Test 3: All unique characters
    assert get_longest_substring("abcdef") == 6

    # Test 4: All same characters
    assert get_longest_substring("aaaaaa") == 1

    # Test 5: Repeating pattern
    assert get_longest_substring("abcabcbb") == 3  # "abc"

    # Test 6: Substring at the end
    assert get_longest_substring("pwwkew") == 3  # "wke"

    # Test 7: Substring in the middle
    assert get_longest_substring("dvdf") == 3  # "vdf"

    # Test 8: Numbers and letters
    assert get_longest_substring("a1b2c3d4") == 8

    # Test 9: Special characters
    assert get_longest_substring("!@#!!@#") == 3  # "!@#"

    # Test 10: Long string with no repeats
    assert get_longest_substring("abcdefghijklmnopqrstuvwxyz") == 26

    print("All test cases passed.")

# Run the tests
#test_get_longest_substring()

