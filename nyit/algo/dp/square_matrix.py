
from typing import List

class Solution:
    
# dynamic programming:
# 1, init a state matrix which has one more row than the matrix
# 2, loop the size, height and width and set the state matrix
# 3, thw higher size will depend on the lowser size and all the extra points as 1

    def countSquares(self, matrix: List[List[int]]) -> int:
        m,n = len(matrix), len(matrix[0])
        state_matrix = [[0]*n for _ in range(m+1)]
        max_size = min(m,n)
        total_count = 0
        for size in range(1,max_size+1):
            for h in range(size-1, m):
                for w in range(0, n):
                    if matrix[h][w] == 1 and state_matrix[h][w] >= (size-1):
                        to_add = True
                        #print("size as ", size)
                        i = size-1
                        if w+i < n and h-i >=0:
                            # check all the extra row items as 1
                            for k in range(w+i,w,-1):
                                if matrix[h][k] != 1:
                                    to_add = False
                                    break
                            if to_add:
                                # check all the extra col items as 1
                                for j in range(h-i,h):
                                    if matrix[j][w+i] != 1:
                                        to_add = False
                                        break
                            if to_add:
                                state_matrix[h+1][w] = size
                                total_count +=1
        return total_count
        