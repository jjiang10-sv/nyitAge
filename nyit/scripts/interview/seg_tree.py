
from typing import List
class Solution:
    def numOfUnplacedFruits(self, fruits: List[int], baskets: List[int]) -> int:
        print("fruits ==== ",fruits)
        print("baskets ==== ", baskets)
        idx = None
        start_fruit_quantity = fruits[0]
        fruits_ = []
        for i in range(len(baskets)):
            if baskets[i] >= start_fruit_quantity:
                idx = i
                break
        root_tree = None;
        if idx is not None:
            root_tree = SegmentTree(start_fruit_quantity, [0, idx-1], [idx+1, len(baskets)-1])
        else:
            root_tree = SegmentTree(start_fruit_quantity, [0, len(baskets)-1],[])
            fruits_.append(0)
        tmp_tree = root_tree
        for i in range(1,len(fruits)):
            fruit_weight = fruits[i]
            while tmp_tree is not None:
                if fruit_weight >= tmp_tree.val:
                    if tmp_tree.right_as_tree:
                        tmp_tree = tmp_tree.right
                    else:  
                        fruit_idx = tmp_tree.search_right(i,baskets ,fruits)
                        if fruit_idx:
                            fruits_.append(fruit_idx)
                        tmp_tree = None
                else:

                    if tmp_tree.left_as_tree:
                        #tmp_tree = tmp_tree.left
                        fruit_idx = tmp_tree.search_recursive(i,baskets ,fruits)
                        if fruit_idx is not None:
                            fruits_.append(fruit_idx) 
                        print(fruit_idx)
                        tmp_tree = None
                    else:  
                        found = tmp_tree.search_left(i,baskets ,fruits)
                        fruit_idx = None
                        if not found:
                            if tmp_tree.right_as_tree:
                                tmp_tree = tmp_tree.right
                            else:
                                fruit_idx = tmp_tree.search_right(i,baskets ,fruits)
                                if fruit_idx is not None:
                                    fruits_.append(fruit_idx) 
                                tmp_tree = None
                        else:
                            tmp_tree = None
            tmp_tree = root_tree
        
        return len(fruits_)
                            
                        
class SegmentTree:
    
    def __init__(self, val, left, right ):
        self.val = val
        def _format(item):
            if len(item) == 0:
                item = None
            elif item[0] == item[1]:
                item = [item[0]]
            elif item[0] > item[1]:
                item = None
            return item
        self.left = _format(left)
        self.right = _format(right)
        self.left_as_tree = False
        self.right_as_tree = False

    def search_recursive(self,fruit_idx, baskets,fruits):
        if self.left_as_tree is False:
            found = self.search_left(fruit_idx=fruit_idx, baskets=baskets,fruits=fruits)
            if found:
                return None
        if self.right_as_tree is False:
            fruit_idx = self.search_right(fruit_idx=fruit_idx, baskets=baskets,fruits=fruits)

            if self.left_as_tree is False:
                return fruit_idx
        if self.left_as_tree:
            self.left.search_recursive(fruit_idx, baskets,fruits)
        if self.right_as_tree:
            self.right.search_recursive(fruit_idx, baskets,fruits)

    def search_left(self, fruit_idx, baskets,fruits):
        if not self.left:
            return False

        fruit_idx_tmp = None
        for idx in self.left:
            if baskets[idx] >= fruits[fruit_idx]:
                fruit_idx_tmp = idx
                break
        if fruit_idx_tmp is not None:
            left = [self.left[0], fruit_idx_tmp-1]
            right_end = self.left[0]
            if len(self.left) == 2 :
                right_end = self.left[1]
            right = [fruit_idx_tmp +1, right_end]
            sub_tree = SegmentTree(fruits[fruit_idx], left,right)
            self.left = sub_tree
            self.left_as_tree = True
            return True
        return False

    def search_right(self, fruit_idx, baskets,fruits):
        if not self.right:
            return fruit_idx

        fruit_idx_tmp = None      
        for idx in self.right:
            if baskets[idx] >= fruits[fruit_idx]:
                fruit_idx_tmp = idx
                break
        if fruit_idx_tmp is not None:
            left = [self.right[0], fruit_idx_tmp-1]
            right_end = self.right[0]
            if len(self.right) == 2 :
                right_end = self.right[1]
            right = [fruit_idx_tmp +1, right_end]
            sub_tree = SegmentTree(fruits[fruit_idx], left,right)
            self.right = sub_tree
            self.right_as_tree = True
        else:
            return fruit_idx
        return None
    

solution = Solution()
fruits =[28,6,3]
baskets =[12,27,35]
print(solution.numOfUnplacedFruits(fruits=fruits,baskets=baskets))