def check_is_ip(s :str):

    # split it into chunks
    # check whether the length of chunks are 4; if not , then false
    # loop through the chunks:
    #   if the chunks length is > 1, then check whether the first char is 0
    #   convert the chunks into int. if convert failed, then false
    #   check whether the chunk value is in [0 , 255], if not false
    #   
    chunks = s.split(":")
    if len(chunks) != 4:
        return False
    for chunk in chunks:
        if len(chunk) > 1 :
            if not chunk[0].isdigit():
                return False
            if int(chunk[0]) == 0:
                return False
        chunk_num = 0
        try:
            chunk_num = int(chunk)
        except :
            return False
        if chunk_num < 0 or chunk_num > 255:
            return False
        
    return True

ips = [['192.168.123.456', False], ["1.2.3.0x1", False], ['123.24.59.99', True]]
for ip in ips:
    #print(check_is_ip(ip))
    print(check_is_ip(ip[0]), ip[1])
    assert check_is_ip(ip[0]) == ip[1]

# Shortest Cell Path
# In a given grid of 0s and 1s, we have some starting row and column sr, sc and a target row and column tr, tc. Return the length of the shortest path from sr, sc to tr, tc that walks along 1 values only.

# Each location in the path, including the start and the end, must be a 1. Each subsequent location in the path must be 4-directionally adjacent to the previous location.

# It is guaranteed that grid[sr][sc] = grid[tr][tc] = 1, and the starting and target positions are different.

# If the task is impossible, return -1.

# from typing import List

# import collections.deque
# def shortestCellPath(grid: List[List[int]], sr: int, sc: int, tr: int, tc: int) -> int:
#     #pass # your code goes here 

#     visited = set()

#     dp = deque((sr,sc,0))
#     i, j = sr, sc
    
#     while dp.pop():
        

	


# 1 2 3 4
# 0 3 0 1
# 5 4 5 1

# (0,0)
# (2, 0)
# (0,1) -- (2-0 + abs(0-1)) = 3
# (0,2) -- (2-0 + abs(0-2))
# start -> go down
# visit = ()

# from collections import deque

# def shortestCellPath(grid, sr, sc, tr, tc):
#     queue = deque()
#     queue.append((sr, sc, 0))
#     seen = set()
#     seen.add((sr, sc))
#     R, C = len(grid), len(grid[0])

#     while queue:
#         r, c, depth = queue.popleft()
#         if r == tr and c == tc:
#             return depth
#         for nr, nc in ((r-1, c), (r+1, c), (r, c-1), (r, c+1)):
#             if 0 <= nr < R and 0 <= nc < C and grid[nr][nc] == 1 and (nr, nc) not in seen:
#                 queue.append((nr, nc, depth + 1))
#                 seen.add((nr, nc))

#     return -1




	


# 1 2 3 4
# 0 3 0 1
# 5 4 5 1

# (0,0)
# (2, 0)
# (0,1) -- (2-0 + abs(0-1)) = 3
# (0,2) -- (2-0 + abs(0-2))
# start -> go down
# visit = ()

# # debug your code below
# grid = [[1, 1, 1, 1], [0, 0, 0, 1], [1, 1, 1, 1]]
# sr, sc, tr, tc = 0, 0, 2, 0
# print(shortestCellPath(grid, sr, sc, tr, tc))