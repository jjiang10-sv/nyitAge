from typing import List
from collections import deque
def shortestCellPath(grid: List[List[int]], sr: int, sc: int, tr: int, tc: int) -> int:
    #pass # your code goes here 
    visited = set()

    dq = deque([(sr,sc,0)])
    row_boundary = len(grid)
    col_boundary = len(grid[0])
    while len(dq) > 0:
        #print(dq.pop())
        (i,j,steps) = dq.pop()
        print("row and col ", i, j)
        # if (i,j) in visited:
        #     continue
        if (j+1) < col_boundary and grid[i][j+1] == 1:
            if (i, j+1) not in visited:
                dq.appendleft((i,j+1, steps+1))
        if (j-1) > -1 and grid[i][j-1] == 1:
            if (i,j-1) not in visited:
                dq.appendleft((i,j-1, steps+1))
        if (i+1) < row_boundary and grid[i+1][j] == 1:
            if (i+1, j) not in visited:
                dq.appendleft((i+1, j, steps+1))
        if (i-1) > -1 and grid[i-1][j] == 1:
            if (i-1,j) not in visited:
                dq.appendleft((i-1, j, steps+1))
        if i == tr and j == tc:
            return steps
        visited.add((i,j))
    return -1

from heapq import heappush, heappop

def dijkstra_grid(grid, sr, sc, tr, tc):
    """
    grid: 2D list of ints, each cell is the cost to step into that cell (0 = blocked)
    (sr, sc): start row, col
    (tr, tc): target row, col
    Returns: minimum cost to reach (tr, tc) from (sr, sc), or -1 if unreachable
    """
    rows, cols = len(grid), len(grid[0])
    if grid[sr][sc] == 0 or grid[tr][tc] == 0:
        return -1

    heap = []
    heappush(heap, (grid[sr][sc], sr, sc))  # (total_cost, row, col)
    visited = set()

    while heap:
        cost, r, c = heappop(heap)
        if (r, c) == (tr, tc):
            return cost
        if (r, c) in visited:
            continue
        visited.add((r, c))
        for dr, dc in [(-1,0), (1,0), (0,-1), (0,1)]:
            nr, nc = r + dr, c + dc
            if 0 <= nr < rows and 0 <= nc < cols and grid[nr][nc] != 0 and (nr, nc) not in visited:
                heappush(heap, (cost + grid[nr][nc], nr, nc))
    return -1

	
# 1 1 1 1
# 0 0 0 1
# 1 1 1 1

# (0,0)
# (2,0)


# 1 2 3 4
# 0 3 0 1
# 5 4 5 1

# (0,0)
# (2, 0)
# (0,1) -- (2-0 + abs(0-1)) = 3
# (0,2) -- (2-0 + abs(0-2))
# start -> go down
# visit = ()

# debug your code below
# grid = [[1, 1, 1, 1], [0, 0, 0, 1], [1, 1, 1, 1]]
# sr, sc, tr, tc = 0, 0, 2, 0
# print(shortestCellPath(grid, sr, sc, tr, tc))


def has_cycle_directed(graph):
    visited = set()
    rec_stack = set()

    def dfs(node):
        visited.add(node)
        rec_stack.add(node)

        for neighbor in graph[node]:
            if neighbor not in visited:
                if dfs(neighbor):
                    return True
            elif neighbor in rec_stack:
                return True

        rec_stack.remove(node)
        return False

    for node in graph:
        if node not in visited:
            if dfs(node):
                return True
    return False

def has_cycle_undirected(graph):
    visited = set()

    def dfs(node, parent):
        visited.add(node)
        for neighbor in graph[node]:
            if neighbor not in visited:
                if dfs(neighbor, node):
                    return True
            # the neighbor is already visited and not the parent, then cycle
            # the nature of undirected graph
            elif neighbor != parent:
                return True
        return False

    for node in graph:
        if node not in visited:
            if dfs(node, None):
                return True
    return False

graph = {
    0: [1, 2],
    1: [2],
    2: [3],
    3: [1]  # This creates a cycle: 1 → 2 → 3 → 1
}

print(has_cycle_undirected(graph=graph))
from typing import Dict, Union

DeepNestedDict = Dict[str, Union[str, 'DeepNestedDict']]


def flatten_dictionary(dictionary: DeepNestedDict) -> Dict[str, str]:
    result = {}
    def dfs(dict_input, key_concat):  
        for key in dict_input:
            if type(dict_input[key]) == dict:   
                key_concat  = key_concat+"." if key_concat !="" else key_concat
                key_concat += key
                dfs(dict_input[key],key_concat)
            else:
                #key_concat += key 
                final_key = key if key_concat == "" else key_concat+"."+key
                result[final_key] = dict_input[key]
                #key_concat -= key

    dfs(dictionary,"")
    return result
# kdsfhsak
# asdjh adslak final_key 

["key1", "key2", ["a","b", "c",["d", "e", [""]]]]
'''
1, use list
2, walk through the dict by keys
3, if value is a int or string, then add (key,value) into the list
4, if value is an dict, walk through the sub-dict by its keys until the value is a int or string
'''
# debug your code below
dict_input = {
    "Key1": "1",
    "Key2": {
        "a": "2",
        "b": "3",
        "c": {
            "d": "3",
            "e": {
                "": "1"
            }
        }
    }
}

print(flatten_dictionary(dict_input))
